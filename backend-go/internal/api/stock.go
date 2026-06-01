package api

import (
	"net/http"
	"stock-monitor/internal/datasource"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"stock-monitor/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type StockHandler struct {
	service *service.StockService
	repo    *repository.StockRepo
	db      *gorm.DB
}

func NewStockHandler(service *service.StockService, repo *repository.StockRepo, db *gorm.DB) *StockHandler {
	return &StockHandler{service: service, repo: repo, db: db}
}

func (h *StockHandler) List(c *gin.Context) {
	results, err := h.service.ListWithStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *StockHandler) Create(c *gin.Context) {
	var req struct {
		StockCode            string          `json:"stock_code" binding:"required"`
		StockName            string          `json:"stock_name" binding:"required"`
		BuyPrice             decimal.Decimal `json:"buy_price"`
		HoldQuantity         decimal.Decimal `json:"hold_quantity"`
		DailyProfitLine      decimal.Decimal `json:"daily_profit_line"`
		DailyLossLine        decimal.Decimal `json:"daily_loss_line"`
		CumulativeProfitLine decimal.Decimal `json:"cumulative_profit_line"`
		CumulativeLossLine   decimal.Decimal `json:"cumulative_loss_line"`
		CumulativeDays       int             `json:"cumulative_days"`
		MonitorInterval      int             `json:"monitor_interval"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, _ := h.repo.GetByCode(req.StockCode)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stock already exists"})
		return
	}

	stock := &model.Stock{
		StockCode:            req.StockCode,
		StockName:            req.StockName,
		BuyPrice:             req.BuyPrice,
		HoldQuantity:         req.HoldQuantity,
		DailyProfitLine:      req.DailyProfitLine,
		DailyLossLine:        req.DailyLossLine,
		CumulativeProfitLine: req.CumulativeProfitLine,
		CumulativeLossLine:   req.CumulativeLossLine,
		CumulativeDays:       req.CumulativeDays,
		MonitorInterval:      req.MonitorInterval,
	}
	if err := h.service.Create(stock); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stock)
}

func (h *StockHandler) Update(c *gin.Context) {
	code := c.Param("code")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Update(code, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stock, err := h.repo.GetByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}
	c.JSON(http.StatusOK, stock)
}

func (h *StockHandler) Delete(c *gin.Context) {
	code := c.Param("code")
	if err := h.service.Delete(code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Where("stock_code = ?", code).Delete(&model.StockGroupItem{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *StockHandler) Detail(c *gin.Context) {
	code := c.Param("code")
	stock, err := h.repo.GetByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}

	spot, err := datasource.FetchStockSpotEnhanced(code)
	if err != nil || spot == nil {
		c.JSON(http.StatusOK, gin.H{
			"stock_code":     stock.StockCode,
			"stock_name":     stock.StockName,
			"current_price":  nil,
			"change_percent": nil,
			"daily_klines":   nil,
			"minute_trends":  nil,
		})
		return
	}

	dailyKlines, _ := datasource.FetchStockTrends(code)
	minuteTrends, _ := datasource.FetchStockMinute(code)

	c.JSON(http.StatusOK, gin.H{
		"stock_code":              stock.StockCode,
		"stock_name":              stock.StockName,
		"buy_price":               stock.BuyPrice,
		"hold_quantity":           stock.HoldQuantity,
		"current_price":           spot.CurrentPrice,
		"change_percent":          spot.ChangePercent,
		"volume":                  spot.Volume,
		"turnover":                spot.Turnover,
		"open":                    spot.OpenPrice,
		"prev_close":              spot.PrevClose,
		"high":                    spot.High,
		"low":                     spot.Low,
		"buy_sell_diff":           spot.BuySellDiff,
		"buy_levels":              spot.BuyLevels,
		"sell_levels":             spot.SellLevels,
		"active_buy_vol":          spot.ActiveBuyVol,
		"active_sell_vol":         spot.ActiveSellVol,
		"large_order_net_inflow":  spot.LargeOrderNetInflow,
		"minute_turnover_rate":    spot.MinuteTurnoverRate,
		"minute_price_volatility": spot.MinutePriceVolatility,
		"daily_profit_line":       stock.DailyProfitLine,
		"daily_loss_line":         stock.DailyLossLine,
		"cumulative_profit_line":  stock.CumulativeProfitLine,
		"cumulative_loss_line":    stock.CumulativeLossLine,
		"cumulative_days":         stock.CumulativeDays,
		"monitor_interval":        stock.MonitorInterval,
		"is_active":               stock.IsActive,
		"daily_klines":            dailyKlines,
		"minute_trends":           minuteTrends,
	})
}
