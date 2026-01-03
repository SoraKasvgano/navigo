package models

import (
	"database/sql"
)

type PageConfig struct {
	ID         int    `json:"id,omitempty"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Logo       string `json:"logo"`
	FooterText string `json:"footer_text"`
	ICP        string `json:"icp"`
}

// GetPageConfig 获取页面配置
func GetPageConfig(db *sql.DB) (*PageConfig, error) {
	config := &PageConfig{}
	err := db.QueryRow(
		"SELECT id, title, subtitle, logo, footer_text, icp FROM page_config WHERE id = 1",
	).Scan(&config.ID, &config.Title, &config.Subtitle, &config.Logo, &config.FooterText, &config.ICP)

	if err == sql.ErrNoRows {
		// 如果没有配置，创建默认配置
		if err := initPageConfig(db); err != nil {
			return getDefaultPageConfig(), nil
		}
		return GetPageConfig(db)
	}

	if err != nil {
		return getDefaultPageConfig(), nil
	}

	return config, nil
}

// UpdatePageConfig 更新页面配置
func UpdatePageConfig(tx *sql.Tx, config *PageConfig) error {
	// 先检查是否存在配置
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM page_config WHERE id = 1").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// 不存在则插入
		_, err = tx.Exec(
			`INSERT INTO page_config (id, title, subtitle, logo, footer_text, icp)
			VALUES (1, ?, ?, ?, ?, ?)`,
			config.Title, config.Subtitle, config.Logo, config.FooterText, config.ICP,
		)
	} else {
		// 存在则更新
		_, err = tx.Exec(
			`UPDATE page_config SET title = ?, subtitle = ?, logo = ?, footer_text = ?, icp = ? WHERE id = 1`,
			config.Title, config.Subtitle, config.Logo, config.FooterText, config.ICP,
		)
	}
	return err
}

// initPageConfig 初始化页面配置
func initPageConfig(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO page_config (id, title, subtitle, logo, footer_text, icp)
		VALUES (1, '网址导航', '常用网址一键直达', '/static/logo.png',
		'', '')
	`)
	return err
}

// getDefaultPageConfig 获取默认配置
func getDefaultPageConfig() *PageConfig {
	return &PageConfig{
		ID:         1,
		Title:      "网址导航",
		Subtitle:   "常用网址一键直达",
		Logo:       "/static/logo.png",
		FooterText: "",
		ICP:        "",
	}
}
