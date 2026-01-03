package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"nav-admin/config"
	"nav-admin/models"
	"nav-admin/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	DB *sql.DB
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		utils.BadRequest(c, "请求格式错误")
		return
	}

	// 获取用户
	user, err := models.GetUserByUsername(h.DB, loginData.Username)
	if err != nil {
		utils.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 验证密码
	if !user.VerifyPassword(loginData.Password) {
		utils.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 设置session cookie
	sessionToken := fmt.Sprintf("%s_%d", user.Username, time.Now().Unix())
	c.SetCookie("session", sessionToken, config.AppConfig.Session.MaxAge, "/", "", false, true)

	utils.SuccessWithMessage(c, "登录成功", gin.H{
		"username": user.Username,
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("session", "", -1, "/", "", false, true)
	utils.SuccessWithMessage(c, "登出成功", nil)
}

// CheckAuth 检查登录状态
func (h *AuthHandler) CheckAuth(c *gin.Context) {
	session, err := c.Cookie("session")
	if err != nil || session == "" {
		c.JSON(http.StatusOK, gin.H{
			"authenticated": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword     string `json:"old_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求格式错误")
		return
	}

	// 安全验证1: 新密码长度检查
	if len(req.NewPassword) < 6 {
		utils.BadRequest(c, "新密码长度不能少于6位")
		return
	}
	if len(req.NewPassword) > 50 {
		utils.BadRequest(c, "新密码长度不能超过50位")
		return
	}

	// 安全验证2: 确认密码匹配
	if req.NewPassword != req.ConfirmPassword {
		utils.BadRequest(c, "两次输入的新密码不一致")
		return
	}

	// 安全验证3: 新旧密码不能相同
	if req.OldPassword == req.NewPassword {
		utils.BadRequest(c, "新密码不能与旧密码相同")
		return
	}

	// 获取当前用户 (从session中解析用户名)
	session, _ := c.Cookie("session")
	// session格式: username_timestamp
	username := "admin" // 默认用户名
	if len(session) > 0 {
		for i := len(session) - 1; i >= 0; i-- {
			if session[i] == '_' {
				username = session[:i]
				break
			}
		}
	}

	// 验证旧密码
	user, err := models.GetUserByUsername(h.DB, username)
	if err != nil {
		utils.InternalServerError(c, "用户不存在")
		return
	}

	if !user.VerifyPassword(req.OldPassword) {
		utils.BadRequest(c, "旧密码错误")
		return
	}

	// 更新密码
	if err := models.UpdatePassword(h.DB, username, req.NewPassword); err != nil {
		utils.InternalServerError(c, "密码更新失败")
		return
	}

	utils.SuccessWithMessage(c, "密码修改成功，请重新登录", nil)
}
