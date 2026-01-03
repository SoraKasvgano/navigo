package handlers

import (
	"fmt"
	"io"
	"nav-admin/config"
	"nav-admin/utils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

// UploadFile 上传文件（支持logo和下载文件）
func (h *UploadHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "未找到上传文件")
		return
	}

	// 检查文件大小
	if file.Size > config.AppConfig.Upload.MaxSize {
		utils.BadRequest(c, fmt.Sprintf("文件大小超过限制（最大%dMB）", config.AppConfig.Upload.MaxSize/1024/1024))
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	uploadType := c.PostForm("type") // logo 或 document
	if uploadType == "" {
		uploadType = c.DefaultQuery("type", "file")
	}

	// 根据上传类型验证文件格式
	var allowedExts []string
	if uploadType == "logo" {
		allowedExts = []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico"}
	} else if uploadType == "document" {
		allowedExts = []string{".txt", ".pdf", ".ppt", ".pptx", ".xls", ".xlsx", ".doc", ".docx", ".rar", ".zip", ".7z"}
	} else {
		// 默认允许所有配置的类型
		allowedExts = config.AppConfig.Upload.AllowedTypes
	}

	allowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			allowed = true
			break
		}
	}

	if !allowed {
		if uploadType == "logo" {
			utils.BadRequest(c, "图标只支持 png, jpg, jpeg, gif, webp, ico 格式")
		} else if uploadType == "document" {
			utils.BadRequest(c, "文件只支持 txt, pdf, ppt, pptx, xls, xlsx, doc, docx, rar, zip, 7z 格式")
		} else {
			utils.BadRequest(c, "不支持的文件类型")
		}
		return
	}

	// 生成唯一文件名
	timestamp := time.Now().Format("20060102150405")
	randomStr := fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	newFilename := fmt.Sprintf("%s_%s%s", timestamp, randomStr, ext)

	// 确定保存路径
	var savePath string
	if uploadType == "logo" {
		savePath = filepath.Join(config.AppConfig.Upload.Path, "logos", newFilename)
		os.MkdirAll(filepath.Join(config.AppConfig.Upload.Path, "logos"), 0755)
	} else {
		savePath = filepath.Join(config.AppConfig.Upload.Path, "files", newFilename)
		os.MkdirAll(filepath.Join(config.AppConfig.Upload.Path, "files"), 0755)
	}

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.InternalServerError(c, "文件保存失败")
		return
	}

	// 返回文件访问路径
	var accessPath string
	if uploadType == "logo" {
		accessPath = fmt.Sprintf("/uploads/logos/%s", newFilename)
	} else {
		accessPath = fmt.Sprintf("/uploads/files/%s", newFilename)
	}

	utils.SuccessWithMessage(c, "上传成功", gin.H{
		"filename":     file.Filename,
		"name":         newFilename,
		"size":         file.Size,
		"url":          accessPath,
		"path":         accessPath,
		"originalName": file.Filename,
	})
}

// DeleteFile 删除文件
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	// 支持 path 或 filename 参数
	filePath := c.Query("path")
	filename := c.Query("filename")

	if filePath == "" && filename == "" {
		utils.BadRequest(c, "文件路径不能为空")
		return
	}

	var fullPath string
	if filePath != "" {
		// 安全检查：只允许删除uploads目录下的文件
		if !strings.HasPrefix(filePath, "/uploads/") {
			utils.BadRequest(c, "无效的文件路径")
			return
		}
		// 构建完整路径
		fullPath = filepath.Join(".", filePath)
	} else {
		// 根据文件名在uploads目录下搜索
		// 先检查files目录，再检查logos目录
		fullPath = filepath.Join(config.AppConfig.Upload.Path, "files", filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fullPath = filepath.Join(config.AppConfig.Upload.Path, "logos", filename)
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		utils.NotFound(c, "文件不存在")
		return
	}

	// 删除文件
	if err := os.Remove(fullPath); err != nil {
		utils.InternalServerError(c, "删除文件失败")
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// ListFiles 列出上传的文件
func (h *UploadHandler) ListFiles(c *gin.Context) {
	uploadType := c.DefaultQuery("type", "all")
	var fileList []gin.H

	// 读取文件的辅助函数
	readDir := func(dirPath string, urlPrefix string) {
		files, err := os.ReadDir(dirPath)
		if err != nil {
			return
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			accessPath := fmt.Sprintf("%s%s", urlPrefix, file.Name())

			fileList = append(fileList, gin.H{
				"name":        file.Name(),
				"size":        info.Size(),
				"modTime":     info.ModTime().Format("2006-01-02 15:04:05"),
				"uploaded_at": info.ModTime().Format("2006-01-02 15:04:05"),
				"path":        accessPath,
				"url":         accessPath,
			})
		}
	}

	if uploadType == "logo" {
		readDir(filepath.Join(config.AppConfig.Upload.Path, "logos"), "/uploads/logos/")
	} else if uploadType == "file" {
		readDir(filepath.Join(config.AppConfig.Upload.Path, "files"), "/uploads/files/")
	} else {
		// all - 读取所有目录
		readDir(filepath.Join(config.AppConfig.Upload.Path, "logos"), "/uploads/logos/")
		readDir(filepath.Join(config.AppConfig.Upload.Path, "files"), "/uploads/files/")
	}

	utils.Success(c, fileList)
}

// DownloadFile 下载文件
func (h *UploadHandler) DownloadFile(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		utils.BadRequest(c, "文件路径不能为空")
		return
	}

	// 安全检查
	if !strings.HasPrefix(filePath, "/uploads/") {
		utils.BadRequest(c, "无效的文件路径")
		return
	}

	// 构建完整路径
	fullPath := filepath.Join(".", filePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		utils.NotFound(c, "文件不存在")
		return
	}

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		utils.InternalServerError(c, "打开文件失败")
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		utils.InternalServerError(c, "获取文件信息失败")
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// 发送文件
	io.Copy(c.Writer, file)
}
