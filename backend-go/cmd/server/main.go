package main

import (
	"net/http"
	"os"
	"stock-monitor/internal/api"
	"stock-monitor/internal/repository"
	"stock-monitor/internal/schedule"
	"stock-monitor/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	logger.Init()
	defer logger.Sync()

	db := repository.InitDB()

	scheduler := schedule.New(db)
	scheduler.Start()
	defer scheduler.Stop()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	api.RegisterRoutes(r, db)

	// In standalone mode, serve frontend static files.
	// In Tauri sidecar mode, Tauri WebView loads its own frontend; skip static file serving.
	if os.Getenv("TAURI_SIDEARCAR") != "1" {
		r.Static("/assets", "../frontend/dist/assets")
		r.GET("/", func(c *gin.Context) {
			c.File("../frontend/dist/index.html")
		})
		r.NoRoute(func(c *gin.Context) {
			c.File("../frontend/dist/index.html")
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	addr := ":" + port

	logger.Log.Info("Server starting", zap.String("addr", addr), zap.String("tauri", os.Getenv("TAURI_SIDEARCAR")))
	if err := r.Run(addr); err != nil {
		logger.Log.Fatal("server failed", zap.Error(err))
	}
}
