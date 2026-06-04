package api

import (
	"net/http"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"stock-monitor/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo *repository.UserRepo
}

func NewAuthHandler(userRepo *repository.UserRepo) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

// QuickRegister creates a user with just a nickname (no password)
func (h *AuthHandler) QuickRegister(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := &model.User{Nickname: req.Nickname}
	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	token, err := service.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"user":     user,
	})
}

// WechatLogin placeholder for WeChat OAuth login
// Requires WECHAT_APPID and WECHAT_SECRET environment variables
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: exchange code for openid via WeChat API
	// For now, return an error indicating configuration is needed
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "WeChat login requires WECHAT_APPID and WECHAT_SECRET configuration",
	})
}

// Me returns the current user info
func (h *AuthHandler) Me(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	cl := claims.(*service.Claims)
	user, err := h.userRepo.GetByID(cl.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
