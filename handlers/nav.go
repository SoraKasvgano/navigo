package handlers

import (
	"database/sql"
	"nav-admin/models"
	"nav-admin/utils"

	"github.com/gin-gonic/gin"
)

type NavHandler struct {
	DB *sql.DB
}

// GetNavData 获取完整的导航数据（用于前端展示）
func (h *NavHandler) GetNavData(c *gin.Context) {
	// 获取页面配置
	pageConfig, err := models.GetPageConfig(h.DB)
	if err != nil {
		utils.InternalServerError(c, "查询页面配置失败")
		return
	}

	// 获取公告配置
	announcementConfig, err := models.GetAnnouncementConfig(h.DB)
	if err != nil {
		utils.InternalServerError(c, "查询公告配置失败")
		return
	}

	// 获取所有分类
	categories, err := models.GetAllCategories(h.DB)
	if err != nil {
		utils.InternalServerError(c, "查询分类失败")
		return
	}

	// 为每个分类获取站点
	for i := range categories {
		sites, err := models.GetSitesByCategoryID(h.DB, categories[i].ID)
		if err != nil {
			continue
		}
		categories[i].Sites = sites
	}

	// 构建返回数据，页面配置和公告配置放在前面
	result := []interface{}{
		map[string]interface{}{
			"type":        "page_config",
			"title":       pageConfig.Title,
			"subtitle":    pageConfig.Subtitle,
			"logo":        pageConfig.Logo,
			"footer_text": pageConfig.FooterText,
			"icp":         pageConfig.ICP,
		},
		announcementConfig,
	}
	for _, cat := range categories {
		result = append(result, cat)
	}

	utils.Success(c, result)
}

// GetPageConfig 获取页面配置
func (h *NavHandler) GetPageConfig(c *gin.Context) {
	config, err := models.GetPageConfig(h.DB)
	if err != nil {
		utils.InternalServerError(c, "获取页面配置失败")
		return
	}
	utils.Success(c, config)
}

// UpdatePageConfig 更新页面配置
func (h *NavHandler) UpdatePageConfig(c *gin.Context) {
	var config models.PageConfig

	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求格式错误")
		return
	}

	// 使用事务
	tx, err := h.DB.Begin()
	if err != nil {
		utils.InternalServerError(c, "事务开始失败")
		return
	}
	defer tx.Rollback()

	if err := models.UpdatePageConfig(tx, &config); err != nil {
		utils.InternalServerError(c, "更新页面配置失败")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	utils.SuccessWithMessage(c, "更新成功", config)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}

// ExportData 导出所有数据为JSON
func (h *NavHandler) ExportData(c *gin.Context) {
	// 获取完整的导航数据
	announcementConfig, err := models.GetAnnouncementConfig(h.DB)
	if err != nil {
		utils.InternalServerError(c, "查询公告配置失败")
		return
	}

	categories, err := models.GetAllCategories(h.DB)
	if err != nil {
		utils.InternalServerError(c, "查询分类失败")
		return
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

	// 设置响应头，触发下载
	c.Header("Content-Disposition", "attachment; filename=nav_data.json")
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(200, result)
}

// ImportData 导入JSON数据
func (h *NavHandler) ImportData(c *gin.Context) {
	var data []map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.BadRequest(c, "请求格式错误")
		return
	}

	// 使用事务
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

	sortNo := 0
	for _, item := range data {
		// 检查是否是公告配置
		if typeVal, ok := item["type"].(string); ok && typeVal == "announcement_config" {
			// 处理公告配置
			if interval, ok := item["interval"].(float64); ok {
				if err := models.UpdateAnnouncementInterval(tx, int(interval)); err != nil {
					utils.InternalServerError(c, "更新公告配置失败")
					return
				}
			}

			// 处理公告列表
			if announcements, ok := item["announcements"].([]interface{}); ok {
				for _, annItem := range announcements {
					if annMap, ok := annItem.(map[string]interface{}); ok {
						ann := &models.Announcement{
							Timestamp: annMap["timestamp"].(string),
							Content:   annMap["content"].(string),
						}
						if _, err := models.CreateAnnouncement(tx, ann); err != nil {
							utils.InternalServerError(c, "创建公告失败")
							return
						}
					}
				}
			}
			continue
		}

		// 处理普通分类
		cat := &models.Category{
			IDStr:    item["_id"].(string),
			Classify: item["classify"].(string),
			Icon:     item["icon"].(string),
			SortNo:   sortNo,
		}

		catID, err := models.CreateCategory(tx, cat)
		if err != nil {
			utils.InternalServerError(c, "创建分类失败")
			return
		}
		sortNo++

		// 处理站点
		if sites, ok := item["sites"].([]interface{}); ok {
			siteSortNo := 0
			for _, siteItem := range sites {
				if siteMap, ok := siteItem.(map[string]interface{}); ok {
					site := &models.Site{
						CatID:  int(catID),
						Name:   siteMap["name"].(string),
						Href:   siteMap["href"].(string),
						SortNo: siteSortNo,
					}

					if desc, ok := siteMap["desc"].(string); ok {
						site.Desc = desc
					}
					if logo, ok := siteMap["logo"].(string); ok {
						site.Logo = logo
					}

					if _, err := models.CreateSite(tx, site); err != nil {
						utils.InternalServerError(c, "创建站点失败")
						return
					}
					siteSortNo++
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	utils.SuccessWithMessage(c, "导入成功", nil)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}
