package api

import (
	"net/http"
	"stock-monitor/internal/datasource"
	"stock-monitor/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

type MarketHandler struct {
	groupRepo      *repository.MarketIndexGroupRepo
	dailyCloseRepo *repository.DailyCloseRepo
}

func NewMarketHandler(groupRepo *repository.MarketIndexGroupRepo, dailyCloseRepo *repository.DailyCloseRepo) *MarketHandler {
	return &MarketHandler{groupRepo: groupRepo, dailyCloseRepo: dailyCloseRepo}
}

func (h *MarketHandler) Get(c *gin.Context) {
	groups, err := h.groupRepo.GetAllWithItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(groups) == 0 {
		// Return empty dashboard if no groups configured
		c.JSON(http.StatusOK, gin.H{
			"update_time": "",
			"categories":  []interface{}{},
		})
		return
	}

	dashboard, err := datasource.FetchMarketDashboard(groups)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save today's close prices and compute MA20
	today := time.Now().Format("2006-01-02")
	for i := range dashboard.Categories {
		for j := range dashboard.Categories[i].Items {
			item := &dashboard.Categories[i].Items[j]
			// Save/upsert today's close price
			_ = h.dailyCloseRepo.Save(item.Symbol, today, item.Price.InexactFloat64(), "")

			// Query last 20 days
			closes, _ := h.dailyCloseRepo.GetLastNDays(item.Symbol, 20)
			if len(closes) > 0 {
				var sum float64
				for _, dc := range closes {
					sum += dc.ClosePrice
				}
				ma := sum / float64(len(closes))
				item.MA20 = &ma
			}
		}
	}

	c.JSON(http.StatusOK, dashboard)
}
