package models

import (
	"database/sql"
)

type Category struct {
	ID       int    `json:"id,omitempty"`
	IDStr    string `json:"_id"`
	Classify string `json:"classify"`
	Icon     string `json:"icon"`
	SortNo   int    `json:"sort_no"`
	Sites    []Site `json:"sites,omitempty"`
}

// GetAllCategories 获取所有分类
func GetAllCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT id, id_str, classify, icon, sort_no FROM categories ORDER BY sort_no, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.IDStr, &cat.Classify, &cat.Icon, &cat.SortNo); err != nil {
			continue
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

// GetCategoryByID 根据ID获取分类
func GetCategoryByID(db *sql.DB, id int) (*Category, error) {
	cat := &Category{}
	err := db.QueryRow(
		"SELECT id, id_str, classify, icon, sort_no FROM categories WHERE id = ?",
		id,
	).Scan(&cat.ID, &cat.IDStr, &cat.Classify, &cat.Icon, &cat.SortNo)

	if err != nil {
		return nil, err
	}
	return cat, nil
}

// CreateCategory 创建分类
func CreateCategory(tx *sql.Tx, cat *Category) (int64, error) {
	// 获取最大排序号
	var maxSortNo int
	err := tx.QueryRow("SELECT COALESCE(MAX(sort_no), -1) FROM categories").Scan(&maxSortNo)
	if err != nil {
		return 0, err
	}

	cat.SortNo = maxSortNo + 1

	result, err := tx.Exec(
		"INSERT INTO categories (id_str, classify, icon, sort_no) VALUES (?, ?, ?, ?)",
		cat.IDStr, cat.Classify, cat.Icon, cat.SortNo,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateCategory 更新分类
func UpdateCategory(tx *sql.Tx, id int, cat *Category) error {
	_, err := tx.Exec(
		"UPDATE categories SET id_str = ?, classify = ?, icon = ? WHERE id = ?",
		cat.IDStr, cat.Classify, cat.Icon, id,
	)
	return err
}

// DeleteCategory 删除分类（会级联删除站点）
func DeleteCategory(tx *sql.Tx, id int) error {
	// 先获取该分类下的所有站点，删除关联的上传文件
	sites, err := GetSitesByCategoryID(tx, id)
	if err != nil {
		return err
	}

	// 删除站点关联的文件
	for _, site := range sites {
		if err := DeleteSiteFile(site.Href); err != nil {
			// 记录错误但继续执行
			continue
		}
	}

	// 删除分类（会自动级联删除站点）
	_, err = tx.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}

// UpdateCategorySortNo 更新分类排序
func UpdateCategorySortNo(tx *sql.Tx, id int, sortNo int) error {
	_, err := tx.Exec("UPDATE categories SET sort_no = ? WHERE id = ?", sortNo, id)
	return err
}
