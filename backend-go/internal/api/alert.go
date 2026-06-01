package api

import (
	"net/http"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AlertHandler struct {
	repo *repository.AlertRepo
}

func NewAlertHandler(repo *repository.AlertRepo) *AlertHandler {
	return &AlertHandler{repo: repo}
}

func (h *AlertHandler) List(c *gin.Context) {
	targetCode := c.Query("target_code")
	targetType := c.Query("target_type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit <= 0 {
		limit = 100
	}

	var tt model.TargetType
	if targetType != "" {
		tt = model.TargetType(targetType)
	}

	alerts, err := h.repo.GetAll(targetCode, tt, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}
