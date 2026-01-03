package handlers

import (
	"database/sql"
	"nav-admin/models"
	"nav-admin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SiteHandler struct {
	DB *sql.DB
}

// GetByCategoryID 获取指定分类的所有站点
func (h *SiteHandler) GetByCategoryID(c *gin.Context) {
	catID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的分类ID")
		return
	}

	sites, err := models.GetSitesByCategoryID(h.DB, catID)
	if err != nil {
		utils.InternalServerError(c, "查询失败")
		return
	}

	utils.Success(c, sites)
}

// GetByID 根据ID获取站点
func (h *SiteHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	site, err := models.GetSiteByID(h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFound(c, "站点不存在")
		} else {
			utils.InternalServerError(c, "查询失败")
		}
		return
	}

	utils.Success(c, site)
}

// Create 创建站点
func (h *SiteHandler) Create(c *gin.Context) {
	var site models.Site
	if err := c.ShouldBindJSON(&site); err != nil {
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

	id, err := models.CreateSite(tx, &site)
	if err != nil {
		utils.InternalServerError(c, "创建失败")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	site.ID = int(id)
	utils.SuccessWithMessage(c, "创建成功", site)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}

// Update 更新站点
func (h *SiteHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var site models.Site
	if err := c.ShouldBindJSON(&site); err != nil {
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

	err = models.UpdateSite(tx, id, &site)
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

// Delete 删除站点
func (h *SiteHandler) Delete(c *gin.Context) {
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

	err = models.DeleteSite(tx, id)
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

// UpdateSort 更新站点排序
func (h *SiteHandler) UpdateSort(c *gin.Context) {
	var sortData struct {
		Items []struct {
			ID     int `json:"id"`
			SortNo int `json:"sort_no"`
		} `json:"items"`
	}

	if err := c.ShouldBindJSON(&sortData); err != nil {
		utils.BadRequest(c, "请求格式错误")
		return
	}

	// 安全验证1: 限制items数组长度，防止DoS攻击
	if len(sortData.Items) == 0 {
		utils.BadRequest(c, "排序列表不能为空")
		return
	}
	if len(sortData.Items) > 500 {
		utils.BadRequest(c, "排序列表过长，最多支持500项")
		return
	}

	// 安全验证2: 收集所有ID并验证sort_no范围
	ids := make([]int, 0, len(sortData.Items))
	for _, item := range sortData.Items {
		if item.ID <= 0 {
			utils.BadRequest(c, "无效的站点ID")
			return
		}
		if item.SortNo < 0 || item.SortNo > 10000 {
			utils.BadRequest(c, "排序值超出有效范围(0-10000)")
			return
		}
		ids = append(ids, item.ID)
	}

	// 安全验证3: 检查ID是否有重复
	idSet := make(map[int]bool)
	for _, id := range ids {
		if idSet[id] {
			utils.BadRequest(c, "排序列表中存在重复ID")
			return
		}
		idSet[id] = true
	}

	// 使用事务
	tx, err := h.DB.Begin()
	if err != nil {
		utils.InternalServerError(c, "事务开始失败")
		return
	}
	defer tx.Rollback()

	// 安全验证4: 验证所有ID都存在，并且属于同一分类
	var expectedCatID int = -1
	for _, id := range ids {
		var catID int
		err := tx.QueryRow("SELECT cat_id FROM sites WHERE id = ?", id).Scan(&catID)
		if err != nil {
			utils.BadRequest(c, "站点ID不存在: "+strconv.Itoa(id))
			return
		}
		// 记录第一个站点的分类ID
		if expectedCatID == -1 {
			expectedCatID = catID
		}
		// 验证所有站点属于同一分类
		if catID != expectedCatID {
			utils.BadRequest(c, "不允许跨分类修改站点排序")
			return
		}
	}

	// 执行更新
	for _, item := range sortData.Items {
		result, err := tx.Exec("UPDATE sites SET sort_no = ? WHERE id = ?", item.SortNo, item.ID)
		if err != nil {
			utils.InternalServerError(c, "更新排序失败")
			return
		}
		// 安全验证5: 确认更新影响了1行
		affected, _ := result.RowsAffected()
		if affected != 1 {
			utils.InternalServerError(c, "更新排序异常")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	utils.SuccessWithMessage(c, "排序更新成功", nil)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}
