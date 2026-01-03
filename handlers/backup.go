package handlers

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"nav-admin/config"
	"nav-admin/models"
	"nav-admin/utils"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type BackupHandler struct {
	DB *sql.DB
}

// 允许的文件扩展名（用于安全验证）
var allowedBackupExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".ico":  true,
	".svg":  true,
	".txt":  true,
	".pdf":  true,
	".ppt":  true,
	".pptx": true,
	".xls":  true,
	".xlsx": true,
	".doc":  true,
	".docx": true,
	".rar":  true,
	".zip":  true,
	".7z":   true,
}

// ExportBackup 导出完整备份为zip文件
func (h *BackupHandler) ExportBackup(c *gin.Context) {
	// 创建内存中的zip缓冲区
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// 1. 导出nav.json数据
	navData, err := h.getNavData()
	if err != nil {
		utils.InternalServerError(c, "获取导航数据失败: "+err.Error())
		return
	}

	navJSON, err := json.MarshalIndent(navData, "", "  ")
	if err != nil {
		utils.InternalServerError(c, "序列化导航数据失败")
		return
	}

	// 将nav.json写入zip
	navFile, err := zipWriter.Create("nav.json")
	if err != nil {
		utils.InternalServerError(c, "创建nav.json失败")
		return
	}
	navFile.Write(navJSON)

	// 2. 添加uploads目录下的文件
	uploadPath := config.AppConfig.Upload.Path
	err = h.addDirectoryToZip(zipWriter, uploadPath, "uploads")
	if err != nil {
		utils.InternalServerError(c, "添加上传文件失败: "+err.Error())
		return
	}

	// 关闭zip writer
	if err := zipWriter.Close(); err != nil {
		utils.InternalServerError(c, "创建zip文件失败")
		return
	}

	// 设置响应头，触发下载
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("nav_backup_%s.zip", timestamp)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Length", fmt.Sprintf("%d", buf.Len()))
	c.Data(200, "application/zip", buf.Bytes())
}

// ImportBackup 从zip文件导入备份
func (h *BackupHandler) ImportBackup(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "未找到上传文件")
		return
	}

	// 检查文件大小（限制为50MB）
	maxSize := int64(50 * 1024 * 1024)
	if file.Size > maxSize {
		utils.BadRequest(c, "文件大小超过限制（最大50MB）")
		return
	}

	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		utils.BadRequest(c, "只支持zip格式文件")
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		utils.InternalServerError(c, "打开上传文件失败")
		return
	}
	defer src.Close()

	// 读取文件内容到内存
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		utils.InternalServerError(c, "读取文件失败")
		return
	}

	// 验证并解析zip文件
	zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		utils.BadRequest(c, "无效的zip文件格式")
		return
	}

	// 安全验证zip内容
	if err := h.validateZipContent(zipReader); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 查找并读取nav.json
	var navData []map[string]interface{}
	var hasNavJSON bool
	for _, f := range zipReader.File {
		if f.Name == "nav.json" {
			hasNavJSON = true
			rc, err := f.Open()
			if err != nil {
				utils.InternalServerError(c, "读取nav.json失败")
				return
			}
			defer rc.Close()

			if err := json.NewDecoder(rc).Decode(&navData); err != nil {
				utils.BadRequest(c, "nav.json格式无效")
				return
			}
			break
		}
	}

	if !hasNavJSON {
		utils.BadRequest(c, "zip文件中缺少nav.json")
		return
	}

	// 验证nav.json内容格式
	if err := h.validateNavJSON(navData); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 开始导入数据
	tx, err := h.DB.Begin()
	if err != nil {
		utils.InternalServerError(c, "事务开始失败")
		return
	}
	defer tx.Rollback()

	// 清空现有数据
	if _, err := tx.Exec("DELETE FROM sites"); err != nil {
		utils.InternalServerError(c, "清空站点失败")
		return
	}
	if _, err := tx.Exec("DELETE FROM categories"); err != nil {
		utils.InternalServerError(c, "清空分类失败")
		return
	}
	if _, err := tx.Exec("DELETE FROM announcements"); err != nil {
		utils.InternalServerError(c, "清空公告失败")
		return
	}

	// 导入nav.json数据
	if err := h.importNavData(tx, navData); err != nil {
		utils.InternalServerError(c, "导入数据失败: "+err.Error())
		return
	}

	// 提交数据库事务
	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	// 解压uploads目录下的文件（在数据库事务成功后）
	if err := h.extractUploadsFromZip(zipReader); err != nil {
		// 记录错误但不阻止整个导入
		fmt.Printf("解压上传文件时出错: %v\n", err)
	}

	utils.SuccessWithMessage(c, "备份导入成功", nil)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}

// getNavData 获取完整的导航数据
func (h *BackupHandler) getNavData() ([]interface{}, error) {
	announcementConfig, err := models.GetAnnouncementConfig(h.DB)
	if err != nil {
		return nil, err
	}

	categories, err := models.GetAllCategories(h.DB)
	if err != nil {
		return nil, err
	}

	for i := range categories {
		sites, err := models.GetSitesByCategoryID(h.DB, categories[i].ID)
		if err != nil {
			continue
		}
		categories[i].Sites = sites
	}

	result := []interface{}{announcementConfig}
	for _, cat := range categories {
		result = append(result, cat)
	}

	return result, nil
}

// addDirectoryToZip 将目录添加到zip文件
func (h *BackupHandler) addDirectoryToZip(zipWriter *zip.Writer, srcPath string, destPath string) error {
	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略不存在的目录
		}

		// 跳过目录本身
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return nil
		}

		// 创建zip中的路径
		zipPath := filepath.Join(destPath, relPath)
		zipPath = filepath.ToSlash(zipPath) // 统一使用正斜杠

		// 打开源文件
		srcFile, err := os.Open(path)
		if err != nil {
			return nil // 忽略无法打开的文件
		}
		defer srcFile.Close()

		// 创建zip中的文件
		writer, err := zipWriter.Create(zipPath)
		if err != nil {
			return nil
		}

		// 复制文件内容
		_, err = io.Copy(writer, srcFile)
		return err
	})
}

// validateZipContent 验证zip文件内容的安全性
func (h *BackupHandler) validateZipContent(zipReader *zip.Reader) error {
	hasNavJSON := false

	for _, f := range zipReader.File {
		name := f.Name

		// 检查路径穿越攻击
		if strings.Contains(name, "..") {
			return fmt.Errorf("检测到非法路径: %s", name)
		}

		// 清理路径并检查
		cleanName := filepath.Clean(name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return fmt.Errorf("检测到非法路径: %s", name)
		}

		// 跳过目录
		if f.FileInfo().IsDir() {
			continue
		}

		// 检查nav.json
		if name == "nav.json" {
			hasNavJSON = true
			// 检查文件大小（限制为10MB）
			if f.UncompressedSize64 > 10*1024*1024 {
				return fmt.Errorf("nav.json文件过大")
			}
			continue
		}

		// 检查uploads目录下的文件
		if strings.HasPrefix(name, "uploads/") {
			// 必须在logos或files子目录下
			if !strings.HasPrefix(name, "uploads/logos/") && !strings.HasPrefix(name, "uploads/files/") {
				return fmt.Errorf("非法的上传文件路径: %s", name)
			}

			// 检查文件扩展名
			ext := strings.ToLower(filepath.Ext(name))
			if !allowedBackupExtensions[ext] {
				return fmt.Errorf("不允许的文件类型: %s", ext)
			}

			// 检查单个文件大小（限制为10MB）
			if f.UncompressedSize64 > 10*1024*1024 {
				return fmt.Errorf("文件过大: %s", name)
			}

			// 检查文件名是否符合预期格式（防止特殊字符注入）
			baseName := filepath.Base(name)
			if !isValidFilename(baseName) {
				return fmt.Errorf("非法的文件名: %s", baseName)
			}

			continue
		}

		// 其他文件一律拒绝
		return fmt.Errorf("包含非法文件: %s", name)
	}

	if !hasNavJSON {
		return fmt.Errorf("缺少nav.json文件")
	}

	return nil
}

// isValidFilename 检查文件名是否合法
func isValidFilename(name string) bool {
	// 只允许字母、数字、下划线、点、横线
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
	return validPattern.MatchString(name)
}

// validateNavJSON 验证nav.json的内容格式
func (h *BackupHandler) validateNavJSON(data []map[string]interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("nav.json数据为空")
	}

	for _, item := range data {
		// 检查公告配置
		if typeVal, ok := item["type"].(string); ok && typeVal == "announcement_config" {
			// 验证公告配置结构
			if _, ok := item["_id"].(string); !ok {
				return fmt.Errorf("公告配置缺少_id字段")
			}
			continue
		}

		// 检查分类结构
		if _, ok := item["_id"].(string); !ok {
			return fmt.Errorf("分类缺少_id字段")
		}
		if _, ok := item["classify"].(string); !ok {
			return fmt.Errorf("分类缺少classify字段")
		}

		// 检查站点数组
		if sites, ok := item["sites"].([]interface{}); ok {
			for i, site := range sites {
				siteMap, ok := site.(map[string]interface{})
				if !ok {
					return fmt.Errorf("第%d个站点格式错误", i+1)
				}
				if _, ok := siteMap["name"].(string); !ok {
					return fmt.Errorf("站点缺少name字段")
				}
				if _, ok := siteMap["href"].(string); !ok {
					return fmt.Errorf("站点缺少href字段")
				}

				// 验证logo路径安全性
				if logo, ok := siteMap["logo"].(string); ok && logo != "" {
					if strings.Contains(logo, "..") {
						return fmt.Errorf("检测到非法logo路径")
					}
				}

				// 验证href路径安全性（如果是本地文件）
				if href, ok := siteMap["href"].(string); ok {
					if strings.HasPrefix(href, "/uploads/") && strings.Contains(href, "..") {
						return fmt.Errorf("检测到非法href路径")
					}
				}
			}
		}
	}

	return nil
}

// importNavData 导入nav.json数据到数据库
func (h *BackupHandler) importNavData(tx *sql.Tx, data []map[string]interface{}) error {
	sortNo := 0
	for _, item := range data {
		// 检查是否是公告配置
		if typeVal, ok := item["type"].(string); ok && typeVal == "announcement_config" {
			// 处理公告配置
			if interval, ok := item["interval"].(float64); ok {
				if err := models.UpdateAnnouncementInterval(tx, int(interval)); err != nil {
					return fmt.Errorf("更新公告配置失败: %v", err)
				}
			}

			// 处理公告列表
			if announcements, ok := item["announcements"].([]interface{}); ok {
				for _, annItem := range announcements {
					if annMap, ok := annItem.(map[string]interface{}); ok {
						timestamp, _ := annMap["timestamp"].(string)
						content, _ := annMap["content"].(string)
						ann := &models.Announcement{
							Timestamp: timestamp,
							Content:   content,
						}
						if _, err := models.CreateAnnouncement(tx, ann); err != nil {
							return fmt.Errorf("创建公告失败: %v", err)
						}
					}
				}
			}
			continue
		}

		// 处理普通分类
		idStr, _ := item["_id"].(string)
		classify, _ := item["classify"].(string)
		icon, _ := item["icon"].(string)

		cat := &models.Category{
			IDStr:    idStr,
			Classify: classify,
			Icon:     icon,
			SortNo:   sortNo,
		}

		catID, err := models.CreateCategory(tx, cat)
		if err != nil {
			return fmt.Errorf("创建分类失败: %v", err)
		}
		sortNo++

		// 处理站点
		if sites, ok := item["sites"].([]interface{}); ok {
			siteSortNo := 0
			for _, siteItem := range sites {
				if siteMap, ok := siteItem.(map[string]interface{}); ok {
					name, _ := siteMap["name"].(string)
					href, _ := siteMap["href"].(string)
					desc, _ := siteMap["desc"].(string)
					logo, _ := siteMap["logo"].(string)

					site := &models.Site{
						CatID:  int(catID),
						Name:   name,
						Href:   href,
						Desc:   desc,
						Logo:   logo,
						SortNo: siteSortNo,
					}

					if _, err := models.CreateSite(tx, site); err != nil {
						return fmt.Errorf("创建站点失败: %v", err)
					}
					siteSortNo++
				}
			}
		}
	}

	return nil
}

// extractUploadsFromZip 从zip中解压uploads目录的文件
func (h *BackupHandler) extractUploadsFromZip(zipReader *zip.Reader) error {
	uploadPath := config.AppConfig.Upload.Path

	// 确保logos和files目录存在
	os.MkdirAll(filepath.Join(uploadPath, "logos"), 0755)
	os.MkdirAll(filepath.Join(uploadPath, "files"), 0755)

	for _, f := range zipReader.File {
		// 只处理uploads目录下的文件
		if !strings.HasPrefix(f.Name, "uploads/") {
			continue
		}

		// 跳过目录
		if f.FileInfo().IsDir() {
			continue
		}

		// 计算目标路径
		relPath := strings.TrimPrefix(f.Name, "uploads/")
		destPath := filepath.Join(uploadPath, relPath)

		// 再次验证路径安全性
		cleanPath := filepath.Clean(destPath)
		absUploadPath, _ := filepath.Abs(uploadPath)
		absDestPath, _ := filepath.Abs(cleanPath)
		if !strings.HasPrefix(absDestPath, absUploadPath) {
			continue // 跳过可能的路径穿越
		}

		// 确保父目录存在
		os.MkdirAll(filepath.Dir(destPath), 0755)

		// 打开zip中的文件
		rc, err := f.Open()
		if err != nil {
			continue
		}

		// 创建目标文件
		destFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			continue
		}

		// 复制内容（限制大小）
		_, err = io.CopyN(destFile, rc, 10*1024*1024) // 限制10MB
		destFile.Close()
		rc.Close()

		if err != nil && err != io.EOF {
			os.Remove(destPath) // 清理失败的文件
		}
	}

	return nil
}
