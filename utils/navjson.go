package utils

import (
	"database/sql"
	"encoding/json"
	"log"
	"nav-admin/config"
	"os"
	"path/filepath"
	"sync"
)

// 文件锁，防止并发写入
var navJSONMutex sync.Mutex

// NavJSONCategory 导航分类结构（用于JSON输出）
type NavJSONCategory struct {
	ID       string        `json:"_id"`
	Classify string        `json:"classify"`
	Icon     string        `json:"icon"`
	Sites    []NavJSONSite `json:"sites"`
}

// NavJSONSite 站点结构（用于JSON输出）
type NavJSONSite struct {
	Name string `json:"name"`
	Href string `json:"href"`
	Desc string `json:"desc"`
	Logo string `json:"logo"`
}

// NavJSONAnnouncement 公告结构（用于JSON输出）
type NavJSONAnnouncement struct {
	ID        int    `json:"id"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

// NavJSONAnnouncementConfig 公告配置结构（用于JSON输出）
type NavJSONAnnouncementConfig struct {
	ID            string                `json:"_id"`
	Type          string                `json:"type"`
	Interval      int                   `json:"interval"`
	Announcements []NavJSONAnnouncement `json:"announcements"`
}

// NavJSONPageConfig 页面配置结构（用于JSON输出）
type NavJSONPageConfig struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Logo       string `json:"logo"`
	FooterText string `json:"footer_text"`
	ICP        string `json:"icp"`
}

// GenerateNavJSON 从数据库生成nav.json文件
// 每次数据变更后调用此函数更新静态JSON文件
func GenerateNavJSON(db *sql.DB) error {
	navJSONMutex.Lock()
	defer navJSONMutex.Unlock()

	var result []interface{}

	// 1. 获取页面配置
	pageConfig, err := getPageConfigForJSON(db)
	if err != nil {
		log.Printf("获取页面配置失败: %v", err)
		// 使用默认配置
		pageConfig = &NavJSONPageConfig{
			Type:       "page_config",
			Title:      "网址导航",
			Subtitle:   "常用网址一键直达",
			Logo:       "/static/logo.png",
			FooterText: "",
			ICP:        "",
		}
	}
	result = append(result, pageConfig)

	// 2. 获取公告配置
	announcementConfig, err := getAnnouncementConfigForJSON(db)
	if err != nil {
		log.Printf("获取公告配置失败: %v", err)
		// 使用默认配置
		announcementConfig = &NavJSONAnnouncementConfig{
			ID:            "announcement_config",
			Type:          "announcement_config",
			Interval:      5000,
			Announcements: []NavJSONAnnouncement{},
		}
	}
	result = append(result, announcementConfig)

	// 3. 获取所有分类及其站点
	categories, err := getCategoriesForJSON(db)
	if err != nil {
		log.Printf("获取分类失败: %v", err)
	} else {
		for _, cat := range categories {
			result = append(result, cat)
		}
	}

	// 3. 序列化为JSON（带缩进，便于阅读）
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	// 4. 写入文件
	outputPath := config.AppConfig.Nav.JSONPath
	if outputPath == "" {
		outputPath = "../static/nav.json"
	}

	// 确保目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入文件
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return err
	}

	log.Printf("nav.json 已更新: %s", outputPath)
	return nil
}

// getAnnouncementConfigForJSON 获取公告配置（用于JSON输出）
func getAnnouncementConfigForJSON(db *sql.DB) (*NavJSONAnnouncementConfig, error) {
	// 获取轮播间隔
	var interval int
	err := db.QueryRow("SELECT interval FROM announcement_config WHERE id = 1").Scan(&interval)
	if err != nil {
		interval = 5000
	}

	// 获取公告列表
	rows, err := db.Query("SELECT id, timestamp, content FROM announcements ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announcements []NavJSONAnnouncement
	for rows.Next() {
		var ann NavJSONAnnouncement
		if err := rows.Scan(&ann.ID, &ann.Timestamp, &ann.Content); err != nil {
			continue
		}
		announcements = append(announcements, ann)
	}

	if announcements == nil {
		announcements = []NavJSONAnnouncement{}
	}

	return &NavJSONAnnouncementConfig{
		ID:            "announcement_config",
		Type:          "announcement_config",
		Interval:      interval,
		Announcements: announcements,
	}, nil
}

// getCategoriesForJSON 获取分类列表（用于JSON输出）
func getCategoriesForJSON(db *sql.DB) ([]NavJSONCategory, error) {
	// 获取所有分类
	rows, err := db.Query("SELECT id, id_str, classify, icon FROM categories ORDER BY sort_no, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []NavJSONCategory
	for rows.Next() {
		var id int
		var cat NavJSONCategory
		if err := rows.Scan(&id, &cat.ID, &cat.Classify, &cat.Icon); err != nil {
			continue
		}

		// 获取该分类下的站点
		sites, err := getSitesForJSON(db, id)
		if err != nil {
			cat.Sites = []NavJSONSite{}
		} else {
			cat.Sites = sites
		}

		categories = append(categories, cat)
	}

	return categories, nil
}

// getSitesForJSON 获取站点列表（用于JSON输出）
func getSitesForJSON(db *sql.DB, categoryID int) ([]NavJSONSite, error) {
	rows, err := db.Query(
		"SELECT name, href, description, logo FROM sites WHERE cat_id = ? ORDER BY sort_no, id",
		categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []NavJSONSite
	for rows.Next() {
		var site NavJSONSite
		var desc, logo sql.NullString
		if err := rows.Scan(&site.Name, &site.Href, &desc, &logo); err != nil {
			continue
		}
		site.Desc = desc.String
		site.Logo = logo.String
		sites = append(sites, site)
	}

	if sites == nil {
		sites = []NavJSONSite{}
	}

	return sites, nil
}

// getPageConfigForJSON 获取页面配置（用于JSON输出）
func getPageConfigForJSON(db *sql.DB) (*NavJSONPageConfig, error) {
	var config NavJSONPageConfig
	config.Type = "page_config"

	err := db.QueryRow("SELECT title, subtitle, logo, footer_text, icp FROM page_config WHERE id = 1").Scan(
		&config.Title, &config.Subtitle, &config.Logo, &config.FooterText, &config.ICP,
	)

	if err != nil {
		return nil, err
	}

	return &config, nil
}
