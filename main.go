package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"nav-admin/config"
	"nav-admin/handlers"
	"nav-admin/middleware"
	"nav-admin/utils"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

var db *sql.DB

func main() {
	// 初始化配置
	config.Init()

	// 初始化数据库
	var err error
	db, err = utils.InitDB(config.AppConfig.Database.Path)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer db.Close()

	// 设置Gin模式
	gin.SetMode(config.AppConfig.Server.Mode)

	r := gin.Default()

	// 设置信任的代理
	r.SetTrustedProxies(nil)

	// nav.json 使用运行时生成的文件
	r.StaticFile("/nav.json", config.AppConfig.Nav.JSONPath)

	// 静态文件服务（从嵌入的文件系统）
	staticSubFS, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/static", http.FS(staticSubFS))

	// 上传文件服务
	r.Static("/uploads", config.AppConfig.Upload.Path)

	// 初始化处理器
	authHandler := &handlers.AuthHandler{DB: db}
	categoryHandler := &handlers.CategoryHandler{DB: db}
	siteHandler := &handlers.SiteHandler{DB: db}
	announcementHandler := &handlers.AnnouncementHandler{DB: db}
	uploadHandler := &handlers.UploadHandler{}
	navHandler := &handlers.NavHandler{DB: db}
	backupHandler := &handlers.BackupHandler{DB: db}

	// 前端页面路由
	r.GET("/", func(c *gin.Context) {
		data, _ := templatesFS.ReadFile("templates/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	r.GET("/admin", func(c *gin.Context) {
		data, _ := templatesFS.ReadFile("templates/admin.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	r.GET("/login", func(c *gin.Context) {
		data, _ := templatesFS.ReadFile("templates/login.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// API路由
	api := r.Group("/api")
	{
		// 公开接口
		api.POST("/login", authHandler.Login)
		api.GET("/check-auth", authHandler.CheckAuth)
		api.GET("/nav", navHandler.GetNavData) // 获取导航数据（前端展示用）

		// 需要认证的管理接口
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		{
			// 认证相关
			admin.POST("/logout", authHandler.Logout)
			admin.PUT("/change-password", authHandler.ChangePassword)

			// 分类管理
			admin.GET("/categories", categoryHandler.GetAll)
			admin.GET("/categories/:id", categoryHandler.GetByID)
			admin.POST("/categories", categoryHandler.Create)
			admin.PUT("/categories/:id", categoryHandler.Update)
			admin.DELETE("/categories/:id", categoryHandler.Delete)
			admin.PUT("/categories/sort", categoryHandler.UpdateSort)

			// 站点管理
			admin.GET("/categories/:id/sites", siteHandler.GetByCategoryID)
			admin.GET("/sites/:id", siteHandler.GetByID)
			admin.POST("/sites", siteHandler.Create)
			admin.PUT("/sites/:id", siteHandler.Update)
			admin.DELETE("/sites/:id", siteHandler.Delete)
			admin.PUT("/sites/sort", siteHandler.UpdateSort)

			// 公告管理
			admin.GET("/announcements", announcementHandler.GetAll)
			admin.GET("/announcements/:id", announcementHandler.GetByID)
			admin.POST("/announcements", announcementHandler.Create)
			admin.PUT("/announcements/:id", announcementHandler.Update)
			admin.DELETE("/announcements/:id", announcementHandler.Delete)

			// 公告配置
			admin.GET("/announcement-config", announcementHandler.GetConfig)
			admin.PUT("/announcement-config", announcementHandler.UpdateConfig)

			// 页面配置
			admin.GET("/page-config", navHandler.GetPageConfig)
			admin.PUT("/page-config", navHandler.UpdatePageConfig)

			// 文件上传
			admin.POST("/upload", uploadHandler.UploadFile)
			admin.DELETE("/upload", uploadHandler.DeleteFile)
			admin.GET("/files", uploadHandler.ListFiles)

			// 数据导入导出
			admin.GET("/export", navHandler.ExportData)
			admin.POST("/import", navHandler.ImportData)

			// 完整备份（包含上传文件的zip）
			admin.GET("/backup/export", backupHandler.ExportBackup)
			admin.POST("/backup/import", backupHandler.ImportBackup)
		}
	}

	// 启动服务器
	addr := ":" + config.AppConfig.Server.Port
	log.Printf("服务器启动在 http://localhost%s", addr)
	log.Printf("管理后台: http://localhost%s/admin", addr)
	log.Printf("登录页面: http://localhost%s/login", addr)
	log.Println("默认账号: admin / admin")

	if err := r.Run(addr); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
