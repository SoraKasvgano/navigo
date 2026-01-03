package models

import (
	"database/sql"
	"time"
)

type Announcement struct {
	ID        int    `json:"id"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

type AnnouncementConfig struct {
	IDStr         string         `json:"_id"`
	Type          string         `json:"type"`
	Interval      int            `json:"interval"`
	Announcements []Announcement `json:"announcements"`
}

// GetAllAnnouncements 获取所有公告
func GetAllAnnouncements(db *sql.DB) ([]Announcement, error) {
	rows, err := db.Query("SELECT id, timestamp, content FROM announcements ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announcements []Announcement
	for rows.Next() {
		var ann Announcement
		if err := rows.Scan(&ann.ID, &ann.Timestamp, &ann.Content); err != nil {
			continue
		}
		announcements = append(announcements, ann)
	}
	return announcements, nil
}

// GetAnnouncementByID 根据ID获取公告
func GetAnnouncementByID(db *sql.DB, id int) (*Announcement, error) {
	ann := &Announcement{}
	err := db.QueryRow(
		"SELECT id, timestamp, content FROM announcements WHERE id = ?",
		id,
	).Scan(&ann.ID, &ann.Timestamp, &ann.Content)

	if err != nil {
		return nil, err
	}
	return ann, nil
}

// CreateAnnouncement 创建公告
func CreateAnnouncement(tx *sql.Tx, ann *Announcement) (int64, error) {
	if ann.Timestamp == "" {
		ann.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}

	result, err := tx.Exec(
		"INSERT INTO announcements (timestamp, content) VALUES (?, ?)",
		ann.Timestamp, ann.Content,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateAnnouncement 更新公告
func UpdateAnnouncement(tx *sql.Tx, id int, ann *Announcement) error {
	_, err := tx.Exec(
		"UPDATE announcements SET timestamp = ?, content = ? WHERE id = ?",
		ann.Timestamp, ann.Content, id,
	)
	return err
}

// DeleteAnnouncement 删除公告
func DeleteAnnouncement(tx *sql.Tx, id int) error {
	_, err := tx.Exec("DELETE FROM announcements WHERE id = ?", id)
	return err
}

// GetAnnouncementInterval 获取公告轮播间隔
func GetAnnouncementInterval(db *sql.DB) (int, error) {
	var interval int
	err := db.QueryRow("SELECT interval FROM announcement_config WHERE id = 1").Scan(&interval)
	if err != nil {
		return 5000, err
	}
	return interval, nil
}

// UpdateAnnouncementInterval 更新公告轮播间隔
func UpdateAnnouncementInterval(tx *sql.Tx, interval int) error {
	_, err := tx.Exec("UPDATE announcement_config SET interval = ? WHERE id = 1", interval)
	return err
}

// GetAnnouncementConfig 获取完整的公告配置
func GetAnnouncementConfig(db *sql.DB) (*AnnouncementConfig, error) {
	interval, err := GetAnnouncementInterval(db)
	if err != nil {
		interval = 5000
	}

	announcements, err := GetAllAnnouncements(db)
	if err != nil {
		return nil, err
	}

	return &AnnouncementConfig{
		IDStr:         "announcement_config",
		Type:          "announcement_config",
		Interval:      interval,
		Announcements: announcements,
	}, nil
}
