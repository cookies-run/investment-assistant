package api

import (
	"net/http"
	"stock-monitor/internal/model"
	"stock-monitor/internal/service"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) GetConfig(c *gin.Context) {
	cfg, err := h.service.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *NotificationHandler) SaveConfig(c *gin.Context) {
	var req struct {
		FeishuWebhook string `json:"feishu_webhook"`
		EnableFeishu  bool   `json:"enable_feishu"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := &model.NotificationConfig{
		FeishuWebhook: req.FeishuWebhook,
		EnableFeishu:  req.EnableFeishu,
	}
	if err := h.service.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}
