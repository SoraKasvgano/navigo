package utils

import (
	"database/sql"
	"log"
	"nav-admin/models"

	_ "modernc.org/sqlite"
)

// InitDB 初始化数据库
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 创建表
	if err := createTables(db); err != nil {
		return nil, err
	}

	// 创建默认用户
	if err := models.CreateDefaultUser(db); err != nil {
		log.Printf("创建默认用户失败: %v", err)
	} else {
		log.Println("默认管理员账号: admin/admin")
	}

	// 初始化公告配置
	if err := initAnnouncementConfig(db); err != nil {
		log.Printf("初始化公告配置失败: %v", err)
	}

	// 初始化页面配置
	if err := initPageConfig(db); err != nil {
		log.Printf("初始化页面配置失败: %v", err)
	}

	// 生成初始nav.json
	if err := GenerateNavJSON(db); err != nil {
		log.Printf("生成nav.json失败: %v", err)
	}

	log.Println("数据库初始化完成")
	return db, nil
}

// createTables 创建数据库表
func createTables(db *sql.DB) error {
	// 用户表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 分类表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			id_str TEXT NOT NULL,
			classify TEXT NOT NULL,
			icon TEXT NOT NULL,
			sort_no INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}

	// 站点表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cat_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			href TEXT NOT NULL,
			description TEXT,
			logo TEXT,
			sort_no INTEGER DEFAULT 0,
			FOREIGN KEY (cat_id) REFERENCES categories(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// 公告表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS announcements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			content TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// 公告配置表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS announcement_config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			interval INTEGER DEFAULT 5000
		)
	`)
	if err != nil {
		return err
	}

	// 页面配置表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS page_config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			title TEXT DEFAULT '网址导航',
			subtitle TEXT DEFAULT '常用网址一键直达',
			logo TEXT DEFAULT '/static/logo.png',
			footer_text TEXT DEFAULT '',
			icp TEXT DEFAULT ''
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// initAnnouncementConfig 初始化公告配置
func initAnnouncementConfig(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM announcement_config").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = db.Exec("INSERT INTO announcement_config (id, interval) VALUES (1, 5000)")
		return err
	}

	return nil
}

// initPageConfig 初始化页面配置
func initPageConfig(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM page_config").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = db.Exec(`
			INSERT INTO page_config (id, title, subtitle, logo, footer_text, icp)
			VALUES (1, '网址导航', '常用网址一键直达', '/static/logo.png',
			'', '')
		`)
		return err
	}

	return nil
}
