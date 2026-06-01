package api

import (
	"net/http"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type FundLotHandler struct {
	lotRepo  *repository.FundLotRepo
	fundRepo *repository.FundRepo
}

func NewFundLotHandler(lotRepo *repository.FundLotRepo, fundRepo *repository.FundRepo) *FundLotHandler {
	return &FundLotHandler{lotRepo: lotRepo, fundRepo: fundRepo}
}

func (h *FundLotHandler) List(c *gin.Context) {
	code := c.Param("code")
	lots, err := h.lotRepo.GetByFundCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lots)
}

func (h *FundLotHandler) Create(c *gin.Context) {
	code := c.Param("code")
	var req struct {
		Quantity    decimal.Decimal `json:"quantity" binding:"required"`
		Cost        decimal.Decimal `json:"cost"`
		PurchasedAt string          `json:"purchased_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fund, err := h.fundRepo.GetByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fund not found"})
		return
	}

	purchasedAt := time.Now()
	if req.PurchasedAt != "" {
		if t, err := time.Parse("2006-01-02", req.PurchasedAt); err == nil {
			purchasedAt = t
		}
	}

	lot := &model.FundLot{
		FundCode:    code,
		Quantity:    req.Quantity,
		Cost:        req.Cost,
		PurchasedAt: purchasedAt,
	}
	if err := h.lotRepo.Create(lot); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Sync aggregate hold_quantity and reset cumulative baseline
	newQty := fund.HoldQuantity.Add(req.Quantity)
	if err := h.fundRepo.Update(code, map[string]interface{}{
		"hold_quantity":      newQty,
		"last_rebalanced_at": purchasedAt,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lot)
}

func (h *FundLotHandler) Delete(c *gin.Context) {
	code := c.Param("code")
	id := c.Param("id")
	v, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lot id"})
		return
	}
	lotID := uint(v)

	lot, err := h.lotRepo.GetByID(lotID)
	if err != nil || lot.FundCode != code {
		c.JSON(http.StatusNotFound, gin.H{"error": "lot not found"})
		return
	}

	fund, err := h.fundRepo.GetByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fund not found"})
		return
	}

	if err := h.lotRepo.Delete(lotID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Sync aggregate hold_quantity
	newQty := fund.HoldQuantity.Sub(lot.Quantity)
	if newQty.IsNegative() {
		newQty = decimal.Zero
	}
	if err := h.fundRepo.Update(code, map[string]interface{}{"hold_quantity": newQty}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
