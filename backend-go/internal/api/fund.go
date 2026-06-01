package api

import (
	"net/http"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"stock-monitor/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type FundHandler struct {
	service         *service.FundService
	repo            *repository.FundRepo
	dailyRecordRepo *repository.DailyRecordRepo
	db              *gorm.DB
}

func NewFundHandler(service *service.FundService, repo *repository.FundRepo, dailyRecordRepo *repository.DailyRecordRepo, db *gorm.DB) *FundHandler {
	return &FundHandler{service: service, repo: repo, dailyRecordRepo: dailyRecordRepo, db: db}
}

func (h *FundHandler) List(c *gin.Context) {
	results, err := h.service.ListWithStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *FundHandler) Create(c *gin.Context) {
	var req struct {
		FundCode             string          `json:"fund_code" binding:"required"`
		FundName             string          `json:"fund_name" binding:"required"`
		HoldQuantity         decimal.Decimal `json:"hold_quantity"`
		HoldCost             decimal.Decimal `json:"hold_cost"`
		DailyProfitLine      decimal.Decimal `json:"daily_profit_line"`
		DailyLossLine        decimal.Decimal `json:"daily_loss_line"`
		CumulativeProfitLine decimal.Decimal `json:"cumulative_profit_line"`
		CumulativeLossLine   decimal.Decimal `json:"cumulative_loss_line"`
		CumulativeDays       int             `json:"cumulative_days"`
		FundType             string          `json:"fund_type"`
		RelatedIndexSymbol   string          `json:"related_index_symbol"`
		BaseCurrency         string          `json:"base_currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, _ := h.repo.GetByCode(req.FundCode)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fund already exists"})
		return
	}

	fund := &model.Fund{
		FundCode:             req.FundCode,
		FundName:             req.FundName,
		HoldQuantity:         req.HoldQuantity,
		HoldCost:             req.HoldCost,
		DailyProfitLine:      req.DailyProfitLine,
		DailyLossLine:        req.DailyLossLine,
		CumulativeProfitLine: req.CumulativeProfitLine,
		CumulativeLossLine:   req.CumulativeLossLine,
		CumulativeDays:       req.CumulativeDays,
		FundType:             req.FundType,
		RelatedIndexSymbol:   req.RelatedIndexSymbol,
		BaseCurrency:         req.BaseCurrency,
	}
	if err := h.service.Create(fund); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fund)
}

func (h *FundHandler) Update(c *gin.Context) {
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
	fund, err := h.repo.GetByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fund not found"})
		return
	}
	c.JSON(http.StatusOK, fund)
}

func (h *FundHandler) Delete(c *gin.Context) {
	code := c.Param("code")
	if err := h.service.Delete(code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Where("fund_code = ?", code).Delete(&model.FundGroupItem{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *FundHandler) DailyRecords(c *gin.Context) {
	code := c.Param("code")
	fund, err := h.repo.GetByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fund not found"})
		return
	}

	since := fund.CreatedAt
	if fund.LastRebalancedAt != nil {
		since = *fund.LastRebalancedAt
	}

	records, err := h.dailyRecordRepo.GetSince(code, model.TargetTypeFund, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}
