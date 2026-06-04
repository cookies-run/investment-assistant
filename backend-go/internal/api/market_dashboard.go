package api

import (
	"net/http"
	"strconv"
	"stock-monitor/internal/datasource"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"

	"github.com/gin-gonic/gin"
)

type MarketDashboardHandler struct {
	groupRepo *repository.MarketIndexGroupRepo
	itemRepo  *repository.MarketIndexItemRepo
}

func NewMarketDashboardHandler(groupRepo *repository.MarketIndexGroupRepo, itemRepo *repository.MarketIndexItemRepo) *MarketDashboardHandler {
	return &MarketDashboardHandler{groupRepo: groupRepo, itemRepo: itemRepo}
}

func (h *MarketDashboardHandler) ListGroups(c *gin.Context) {
	userID := getUserID(c)
	groups, err := h.groupRepo.GetAllWithItems(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *MarketDashboardHandler) CreateGroup(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group := &model.MarketIndexGroup{Name: req.Name, UserID: getUserID(c)}
	if err := h.groupRepo.Create(group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *MarketDashboardHandler) UpdateGroup(c *gin.Context) {
	userID := getUserID(c)
	id64, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groupRepo.Update(userID, id, map[string]interface{}{"name": req.Name}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	group, err := h.groupRepo.GetByID(userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *MarketDashboardHandler) DeleteGroup(c *gin.Context) {
	userID := getUserID(c)
	id64, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.groupRepo.Delete(userID, uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MarketDashboardHandler) ReorderGroups(c *gin.Context) {
	userID := getUserID(c)
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groupRepo.Reorder(userID, req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MarketDashboardHandler) CreateItem(c *gin.Context) {
	var req struct {
		GroupID    uint   `json:"group_id" binding:"required"`
		Symbol     string `json:"symbol" binding:"required"`
		Name       string `json:"name" binding:"required"`
		SourceType string `json:"source_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item := &model.MarketIndexItem{
		UserID:     getUserID(c),
		GroupID:    req.GroupID,
		Symbol:     req.Symbol,
		Name:       req.Name,
		SourceType: req.SourceType,
	}
	if err := h.itemRepo.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *MarketDashboardHandler) DeleteItem(c *gin.Context) {
	userID := getUserID(c)
	id64, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.itemRepo.Delete(userID, uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MarketDashboardHandler) ReorderItems(c *gin.Context) {
	userID := getUserID(c)
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.itemRepo.Reorder(userID, req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MarketDashboardHandler) ListAvailableIndices(c *gin.Context) {
	indices := datasource.GetAvailableIndices()
	c.JSON(http.StatusOK, indices)
}
