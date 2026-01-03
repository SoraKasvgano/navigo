package handlers

import (
	"database/sql"
	"nav-admin/models"
	"nav-admin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	DB *sql.DB
}

// GetAll 获取所有分类
func (h *CategoryHandler) GetAll(c *gin.Context) {
	categories, err := models.GetAllCategories(h.DB)
	if err != nil {
		utils.InternalServerError(c, "查询失败")
		return
	}

	utils.Success(c, categories)
}

// GetByID 根据ID获取分类
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	category, err := models.GetCategoryByID(h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFound(c, "分类不存在")
		} else {
			utils.InternalServerError(c, "查询失败")
		}
		return
	}

	// 获取该分类下的站点
	sites, err := models.GetSitesByCategoryID(h.DB, id)
	if err == nil {
		category.Sites = sites
	}

	utils.Success(c, category)
}

// Create 创建分类
func (h *CategoryHandler) Create(c *gin.Context) {
	var cat models.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
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

	id, err := models.CreateCategory(tx, &cat)
	if err != nil {
		utils.InternalServerError(c, "创建失败")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.InternalServerError(c, "提交事务失败")
		return
	}

	cat.ID = int(id)
	utils.SuccessWithMessage(c, "创建成功", cat)

	// 异步更新nav.json
	go utils.GenerateNavJSON(h.DB)
}

// Update 更新分类
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	var cat models.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
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

	err = models.UpdateCategory(tx, id, &cat)
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

// Delete 删除分类
func (h *CategoryHandler) Delete(c *gin.Context) {
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

	err = models.DeleteCategory(tx, id)
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

// UpdateSort 更新分类排序
func (h *CategoryHandler) UpdateSort(c *gin.Context) {
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
	if len(sortData.Items) > 100 {
		utils.BadRequest(c, "排序列表过长，最多支持100项")
		return
	}

	// 安全验证2: 收集所有ID并验证sort_no范围
	ids := make([]int, 0, len(sortData.Items))
	for _, item := range sortData.Items {
		if item.ID <= 0 {
			utils.BadRequest(c, "无效的分类ID")
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

	// 安全验证4: 验证所有ID都存在于数据库中
	for _, id := range ids {
		var exists int
		err := tx.QueryRow("SELECT 1 FROM categories WHERE id = ?", id).Scan(&exists)
		if err != nil {
			utils.BadRequest(c, "分类ID不存在: "+strconv.Itoa(id))
			return
		}
	}

	// 执行更新
	for _, item := range sortData.Items {
		result, err := tx.Exec("UPDATE categories SET sort_no = ? WHERE id = ?", item.SortNo, item.ID)
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
