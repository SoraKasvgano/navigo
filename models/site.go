package models

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
)

type Site struct {
	ID     int    `json:"id,omitempty"`
	CatID  int    `json:"cat_id,omitempty"`
	Name   string `json:"name"`
	Href   string `json:"href"`
	Desc   string `json:"desc"`
	Logo   string `json:"logo"`
	SortNo int    `json:"sort_no"`
}

// GetSitesByCategoryID 获取指定分类的所有站点
func GetSitesByCategoryID(db interface{}, catID int) ([]Site, error) {
	var rows *sql.Rows
	var err error

	// 支持 *sql.DB 和 *sql.Tx
	switch v := db.(type) {
	case *sql.DB:
		rows, err = v.Query("SELECT id, cat_id, name, href, description, logo, sort_no FROM sites WHERE cat_id = ? ORDER BY sort_no, id", catID)
	case *sql.Tx:
		rows, err = v.Query("SELECT id, cat_id, name, href, description, logo, sort_no FROM sites WHERE cat_id = ? ORDER BY sort_no, id", catID)
	default:
		return nil, sql.ErrConnDone
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var site Site
		if err := rows.Scan(&site.ID, &site.CatID, &site.Name, &site.Href, &site.Desc, &site.Logo, &site.SortNo); err != nil {
			continue
		}
		sites = append(sites, site)
	}
	return sites, nil
}

// GetSiteByID 根据ID获取站点
func GetSiteByID(db interface{}, id int) (*Site, error) {
	site := &Site{}
	var err error

	// 支持 *sql.DB 和 *sql.Tx
	switch v := db.(type) {
	case *sql.DB:
		err = v.QueryRow(
			"SELECT id, cat_id, name, href, description, logo, sort_no FROM sites WHERE id = ?",
			id,
		).Scan(&site.ID, &site.CatID, &site.Name, &site.Href, &site.Desc, &site.Logo, &site.SortNo)
	case *sql.Tx:
		err = v.QueryRow(
			"SELECT id, cat_id, name, href, description, logo, sort_no FROM sites WHERE id = ?",
			id,
		).Scan(&site.ID, &site.CatID, &site.Name, &site.Href, &site.Desc, &site.Logo, &site.SortNo)
	default:
		return nil, sql.ErrConnDone
	}

	if err != nil {
		return nil, err
	}
	return site, nil
}

// CreateSite 创建站点
func CreateSite(tx *sql.Tx, site *Site) (int64, error) {
	// 获取该分类下的最大排序号
	var maxSortNo int
	err := tx.QueryRow("SELECT COALESCE(MAX(sort_no), -1) FROM sites WHERE cat_id = ?", site.CatID).Scan(&maxSortNo)
	if err != nil {
		return 0, err
	}

	site.SortNo = maxSortNo + 1

	result, err := tx.Exec(
		"INSERT INTO sites (cat_id, name, href, description, logo, sort_no) VALUES (?, ?, ?, ?, ?, ?)",
		site.CatID, site.Name, site.Href, site.Desc, site.Logo, site.SortNo,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateSite 更新站点
func UpdateSite(tx *sql.Tx, id int, site *Site) error {
	// 先获取旧的站点信息，用于删除旧文件
	oldSite, err := GetSiteByID(tx, id)
	if err != nil {
		return err
	}

	// 如果href改变了，且旧的href是上传的文件，则删除旧文件
	if oldSite.Href != site.Href && isUploadedFile(oldSite.Href) {
		DeleteSiteFile(oldSite.Href)
	}

	_, err = tx.Exec(
		"UPDATE sites SET name = ?, href = ?, description = ?, logo = ? WHERE id = ?",
		site.Name, site.Href, site.Desc, site.Logo, id,
	)
	return err
}

// DeleteSite 删除站点
func DeleteSite(tx *sql.Tx, id int) error {
	// 先获取站点信息，用于删除关联文件
	site, err := GetSiteByID(tx, id)
	if err != nil {
		return err
	}

	// 删除关联的上传文件
	if err := DeleteSiteFile(site.Href); err != nil {
		// 记录错误但继续执行
	}

	_, err = tx.Exec("DELETE FROM sites WHERE id = ?", id)
	return err
}

// UpdateSiteSortNo 更新站点排序
func UpdateSiteSortNo(tx *sql.Tx, id int, sortNo int) error {
	_, err := tx.Exec("UPDATE sites SET sort_no = ? WHERE id = ?", sortNo, id)
	return err
}

// DeleteSiteFile 删除站点关联的上传文件
func DeleteSiteFile(href string) error {
	if !isUploadedFile(href) {
		return nil
	}

	// 提取文件路径
	filePath := strings.TrimPrefix(href, "/uploads/")
	fullPath := filepath.Join("./uploads", filePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}

	// 删除文件
	return os.Remove(fullPath)
}

// isUploadedFile 判断是否是上传的文件
func isUploadedFile(href string) bool {
	return strings.HasPrefix(href, "/uploads/") || strings.HasPrefix(href, "./uploads/")
}
