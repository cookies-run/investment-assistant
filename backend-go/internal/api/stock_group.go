package api

import (
	"net/http"
	"strconv"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"

	"github.com/gin-gonic/gin"
)

type StockGroupHandler struct {
	groupRepo *repository.StockGroupRepo
	itemRepo  *repository.StockGroupItemRepo
}

func NewStockGroupHandler(groupRepo *repository.StockGroupRepo, itemRepo *repository.StockGroupItemRepo) *StockGroupHandler {
	return &StockGroupHandler{groupRepo: groupRepo, itemRepo: itemRepo}
}

func (h *StockGroupHandler) ListGroups(c *gin.Context) {
	userID := getUserID(c)
	groups, err := h.groupRepo.GetAllWithItems(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *StockGroupHandler) CreateGroup(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group := &model.StockGroup{Name: req.Name, UserID: getUserID(c)}
	if err := h.groupRepo.Create(group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *StockGroupHandler) UpdateGroup(c *gin.Context) {
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

func (h *StockGroupHandler) DeleteGroup(c *gin.Context) {
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

func (h *StockGroupHandler) ReorderGroups(c *gin.Context) {
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

func (h *StockGroupHandler) CreateItem(c *gin.Context) {
	var req struct {
		GroupID   uint   `json:"group_id" binding:"required"`
		StockCode string `json:"stock_code" binding:"required"`
		StockName string `json:"stock_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item := &model.StockGroupItem{
		UserID:    getUserID(c),
		GroupID:   req.GroupID,
		StockCode: req.StockCode,
		StockName: req.StockName,
	}
	if err := h.itemRepo.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *StockGroupHandler) DeleteItem(c *gin.Context) {
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

func (h *StockGroupHandler) ReorderItems(c *gin.Context) {
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
