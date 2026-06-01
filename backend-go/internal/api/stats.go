package api

import (
	"net/http"
	"stock-monitor/internal/service"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	service *service.StatsService
}

func NewStatsHandler(service *service.StatsService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) Get(c *gin.Context) {
	stats, err := h.service.GetPortfolioStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
