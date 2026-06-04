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
	userRepo := repository.NewUserRepo(db)

	authHandler := NewAuthHandler(userRepo)
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

	// Auth routes (no middleware)
	api := r.Group("/api")
	{
		api.POST("/auth/quick-register", authHandler.QuickRegister)
		api.POST("/auth/email/send", authHandler.SendEmailCode)
		api.POST("/auth/email/login", authHandler.EmailLogin)
	}

	// Protected routes
	authorized := r.Group("/api")
	authorized.Use(AuthMiddleware())
	{
		authorized.GET("/auth/me", authHandler.Me)
		authorized.GET("/stocks", stockHandler.List)
		authorized.POST("/stocks", stockHandler.Create)
		authorized.PUT("/stocks/:code", stockHandler.Update)
		authorized.DELETE("/stocks/:code", stockHandler.Delete)
		authorized.GET("/stocks/detail/:code", stockHandler.Detail)
		authorized.GET("/stocks/search", func(c *gin.Context) {
			q := c.Query("q")
			results, err := datasource.SearchStocks(q)
			if err != nil {
				c.JSON(http.StatusOK, []interface{}{})
				return
			}
			c.JSON(http.StatusOK, results)
		})

		authorized.GET("/funds", fundHandler.List)
		authorized.POST("/funds", fundHandler.Create)
		authorized.PUT("/funds/:code", fundHandler.Update)
		authorized.DELETE("/funds/:code", fundHandler.Delete)
		authorized.GET("/funds/:code/daily-records", fundHandler.DailyRecords)
		authorized.GET("/funds/search", func(c *gin.Context) {
			q := c.Query("q")
			results, err := datasource.SearchFunds(q)
			if err != nil {
				c.JSON(http.StatusOK, []interface{}{})
				return
			}
			c.JSON(http.StatusOK, results)
		})

		authorized.GET("/fund-holdings/:code", holdingHandler.Get)
		authorized.GET("/fund-holdings/detail/:code", holdingHandler.Detail)
		authorized.POST("/fund-holdings/sync/:code", holdingHandler.Sync)

		authorized.GET("/funds/:code/lots", fundLotHandler.List)
		authorized.POST("/funds/:code/lots", fundLotHandler.Create)
		authorized.DELETE("/funds/:code/lots/:id", fundLotHandler.Delete)

		authorized.GET("/alerts", alertHandler.List)
		authorized.GET("/daily-records", dailyRecordHandler.List)
		authorized.GET("/stats", statsHandler.Get)
		authorized.GET("/market-dashboard", marketHandler.Get)

		// Market dashboard group/item management
		authorized.GET("/market-groups", marketDashboardHandler.ListGroups)
		authorized.POST("/market-groups", marketDashboardHandler.CreateGroup)
		authorized.PUT("/market-groups/:id", marketDashboardHandler.UpdateGroup)
		authorized.DELETE("/market-groups/:id", marketDashboardHandler.DeleteGroup)
		authorized.POST("/market-groups/reorder", marketDashboardHandler.ReorderGroups)

		authorized.POST("/market-items", marketDashboardHandler.CreateItem)
		authorized.DELETE("/market-items/:id", marketDashboardHandler.DeleteItem)
		authorized.POST("/market-items/reorder", marketDashboardHandler.ReorderItems)

		authorized.GET("/available-indices", marketDashboardHandler.ListAvailableIndices)

		// Stock group management
		authorized.GET("/stock-groups", stockGroupHandler.ListGroups)
		authorized.POST("/stock-groups", stockGroupHandler.CreateGroup)
		authorized.PUT("/stock-groups/:id", stockGroupHandler.UpdateGroup)
		authorized.DELETE("/stock-groups/:id", stockGroupHandler.DeleteGroup)
		authorized.POST("/stock-groups/reorder", stockGroupHandler.ReorderGroups)

		authorized.POST("/stock-group-items", stockGroupHandler.CreateItem)
		authorized.DELETE("/stock-group-items/:id", stockGroupHandler.DeleteItem)
		authorized.POST("/stock-group-items/reorder", stockGroupHandler.ReorderItems)

		// Fund group management
		authorized.GET("/fund-groups", fundGroupHandler.ListGroups)
		authorized.POST("/fund-groups", fundGroupHandler.CreateGroup)
		authorized.PUT("/fund-groups/:id", fundGroupHandler.UpdateGroup)
		authorized.DELETE("/fund-groups/:id", fundGroupHandler.DeleteGroup)
		authorized.POST("/fund-groups/reorder", fundGroupHandler.ReorderGroups)

		authorized.POST("/fund-group-items", fundGroupHandler.CreateItem)
		authorized.DELETE("/fund-group-items/:id", fundGroupHandler.DeleteItem)
		authorized.POST("/fund-group-items/reorder", fundGroupHandler.ReorderItems)

		// Notification settings
		authorized.GET("/notification-config", notificationHandler.GetConfig)
		authorized.POST("/notification-config", notificationHandler.SaveConfig)
	}
}
