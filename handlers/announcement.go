package handlers

import (
	"database/sql"
	"nav-admin/models"
	"nav-admin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AnnouncementHandler struct {
	DB *sql.DB
}

// GetAll 获取所有公告
func (h *AnnouncementHandler) GetAll(c *gin.Context) {
	announcements, err := models.GetAllAnnouncements(h.DB)
	if err != nil {
		utils.InternalServerError(c, "查询失败")
		return
	}

	utils.Success(c, announcements)
}

// GetByID 根据ID获取公告
func (h *AnnouncementHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	announcement, err := models.GetAnnouncementByID(h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFound(c, "公告不存在")
		} else {
			utils.InternalServerError(c, "查询失败")
		}
		return
	}

	utils.Success(c, announcement)
}

// Create 创建公告
func (h *AnnouncementHandler) Create(c *gin.Context) {
	var ann models.Announcement
	if err := c.ShouldBindJSON(&ann); err != nil {
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

	id, err := models.CreateAnnouncement(tx, &ann)
	if err != nil {
		utils.InternalServerError(c, "创建失败")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	ann.ID = int(id)
	utils.SuccessWithMessage(c, "创建成功", ann)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}

// Update 更新公告
func (h *AnnouncementHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var ann models.Announcement
	if err := c.ShouldBindJSON(&ann); err != nil {
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

	err = models.UpdateAnnouncement(tx, id, &ann)
	if err != nil {
		utils.InternalServerError(c, "更新失败")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}

// Delete 删除公告
func (h *AnnouncementHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	// 使用事务
	tx, err := h.DB.Begin()
	if err != nil {
		utils.InternalServerError(c, "事务开始失败")
		return
	}
	defer tx.Rollback()

	err = models.DeleteAnnouncement(tx, id)
	if err != nil {
		utils.InternalServerError(c, "删除失败")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}

// GetConfig 获取公告配置
func (h *AnnouncementHandler) GetConfig(c *gin.Context) {
	interval, err := models.GetAnnouncementInterval(h.DB)
	if err != nil {
		utils.InternalServerError(c, "查询失败")
		return
	}

	utils.Success(c, gin.H{"interval": interval})
}

// UpdateConfig 更新公告配置
func (h *AnnouncementHandler) UpdateConfig(c *gin.Context) {
	var config struct {
		Interval int `json:"interval" binding:"required"`
	}

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

	err = models.UpdateAnnouncementInterval(tx, config.Interval)
	if err != nil {
		utils.InternalServerError(c, "更新失败")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}
