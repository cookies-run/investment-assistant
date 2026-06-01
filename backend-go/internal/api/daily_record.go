package api

import (
	"net/http"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DailyRecordHandler struct {
	repo *repository.DailyRecordRepo
}

func NewDailyRecordHandler(repo *repository.DailyRecordRepo) *DailyRecordHandler {
	return &DailyRecordHandler{repo: repo}
}

func (h *DailyRecordHandler) List(c *gin.Context) {
	targetCode := c.Query("target_code")
	targetType := c.Query("target_type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	if limit <= 0 {
		limit = 30
	}

	var tt model.TargetType
	if targetType != "" {
		tt = model.TargetType(targetType)
	}

	records, err := h.repo.GetByTarget(targetCode, tt, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}
