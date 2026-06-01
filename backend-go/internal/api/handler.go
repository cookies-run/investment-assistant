package api

import (
	"net/http"
	"stock-monitor/internal/datasource"
	"stock-monitor/internal/repository"
	"stock-monitor/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	stockRepo := repository.NewStockRepo(db)
	fundRepo := repository.NewFundRepo(db)
	holdingRepo := repository.NewHoldingRepo(db)
	alertRepo := repository.NewAlertRepo(db)
	dailyRecordRepo := repository.NewDailyRecordRepo(db)
	groupRepo := repository.NewMarketIndexGroupRepo(db)
	itemRepo := repository.NewMarketIndexItemRepo(db)
	fundLotRepo := repository.NewFundLotRepo(db)
	dailyCloseRepo := repository.NewDailyCloseRepo(db)
	stockGroupRepo := repository.NewStockGroupRepo(db)
	stockGroupItemRepo := repository.NewStockGroupItemRepo(db)
	fundGroupRepo := repository.NewFundGroupRepo(db)
	fundGroupItemRepo := repository.NewFundGroupItemRepo(db)

	stockService := service.NewStockService(stockRepo, dailyRecordRepo)
	fundService := service.NewFundService(fundRepo)
	statsService := service.NewStatsService(stockRepo, fundRepo)
	notificationConfigRepo := repository.NewNotificationConfigRepo(db)
	notificationService := service.NewNotificationService(notificationConfigRepo)

	stockHandler := NewStockHandler(stockService, stockRepo, db)
	fundHandler := NewFundHandler(fundService, fundRepo, dailyRecordRepo, db)
	notificationHandler := NewNotificationHandler(notificationService)
	// 设计说明：HoldingHandler 注入 dailyRecordRepo，用于在 Detail 接口中计算滚动复合累计收益（rolling_cumulative_return）。
	// 该指标是 14:50 决策引擎第 3/4 关（短线止盈/补仓轨道）的核心输入，必须基于历史日收益做复合乘积。
	holdingHandler := NewHoldingHandler(holdingRepo, fundRepo, fundLotRepo, dailyRecordRepo)
	alertHandler := NewAlertHandler(alertRepo)
	dailyRecordHandler := NewDailyRecordHandler(dailyRecordRepo)
	statsHandler := NewStatsHandler(statsService)
	marketHandler := NewMarketHandler(groupRepo, dailyCloseRepo)
	marketDashboardHandler := NewMarketDashboardHandler(groupRepo, itemRepo)
	fundLotHandler := NewFundLotHandler(fundLotRepo, fundRepo)
	stockGroupHandler := NewStockGroupHandler(stockGroupRepo, stockGroupItemRepo)
	fundGroupHandler := NewFundGroupHandler(fundGroupRepo, fundGroupItemRepo)

	api := r.Group("/api")
	{
		api.GET("/stocks", stockHandler.List)
		api.POST("/stocks", stockHandler.Create)
		api.PUT("/stocks/:code", stockHandler.Update)
		api.DELETE("/stocks/:code", stockHandler.Delete)
		api.GET("/stocks/detail/:code", stockHandler.Detail)
		api.GET("/stocks/search", func(c *gin.Context) {
			q := c.Query("q")
			results, err := datasource.SearchStocks(q)
			if err != nil {
				c.JSON(http.StatusOK, []interface{}{})
				return
			}
			c.JSON(http.StatusOK, results)
		})

		api.GET("/funds", fundHandler.List)
		api.POST("/funds", fundHandler.Create)
		api.PUT("/funds/:code", fundHandler.Update)
		api.DELETE("/funds/:code", fundHandler.Delete)
		api.GET("/funds/:code/daily-records", fundHandler.DailyRecords)
		api.GET("/funds/search", func(c *gin.Context) {
			q := c.Query("q")
			results, err := datasource.SearchFunds(q)
			if err != nil {
				c.JSON(http.StatusOK, []interface{}{})
				return
			}
			c.JSON(http.StatusOK, results)
		})

		api.GET("/fund-holdings/:code", holdingHandler.Get)
		api.GET("/fund-holdings/detail/:code", holdingHandler.Detail)
		api.POST("/fund-holdings/sync/:code", holdingHandler.Sync)

		api.GET("/funds/:code/lots", fundLotHandler.List)
		api.POST("/funds/:code/lots", fundLotHandler.Create)
		api.DELETE("/funds/:code/lots/:id", fundLotHandler.Delete)

		api.GET("/alerts", alertHandler.List)
		api.GET("/daily-records", dailyRecordHandler.List)
		api.GET("/stats", statsHandler.Get)
		api.GET("/market-dashboard", marketHandler.Get)

		// Market dashboard group/item management
		api.GET("/market-groups", marketDashboardHandler.ListGroups)
		api.POST("/market-groups", marketDashboardHandler.CreateGroup)
		api.PUT("/market-groups/:id", marketDashboardHandler.UpdateGroup)
		api.DELETE("/market-groups/:id", marketDashboardHandler.DeleteGroup)
		api.POST("/market-groups/reorder", marketDashboardHandler.ReorderGroups)

		api.POST("/market-items", marketDashboardHandler.CreateItem)
		api.DELETE("/market-items/:id", marketDashboardHandler.DeleteItem)
		api.POST("/market-items/reorder", marketDashboardHandler.ReorderItems)

		api.GET("/available-indices", marketDashboardHandler.ListAvailableIndices)

		// Stock group management
		api.GET("/stock-groups", stockGroupHandler.ListGroups)
		api.POST("/stock-groups", stockGroupHandler.CreateGroup)
		api.PUT("/stock-groups/:id", stockGroupHandler.UpdateGroup)
		api.DELETE("/stock-groups/:id", stockGroupHandler.DeleteGroup)
		api.POST("/stock-groups/reorder", stockGroupHandler.ReorderGroups)

		api.POST("/stock-group-items", stockGroupHandler.CreateItem)
		api.DELETE("/stock-group-items/:id", stockGroupHandler.DeleteItem)
		api.POST("/stock-group-items/reorder", stockGroupHandler.ReorderItems)

		// Fund group management
		api.GET("/fund-groups", fundGroupHandler.ListGroups)
		api.POST("/fund-groups", fundGroupHandler.CreateGroup)
		api.PUT("/fund-groups/:id", fundGroupHandler.UpdateGroup)
		api.DELETE("/fund-groups/:id", fundGroupHandler.DeleteGroup)
		api.POST("/fund-groups/reorder", fundGroupHandler.ReorderGroups)

		api.POST("/fund-group-items", fundGroupHandler.CreateItem)
		api.DELETE("/fund-group-items/:id", fundGroupHandler.DeleteItem)
		api.POST("/fund-group-items/reorder", fundGroupHandler.ReorderItems)

		// Notification settings
		api.GET("/notification-config", notificationHandler.GetConfig)
		api.POST("/notification-config", notificationHandler.SaveConfig)
	}
}
