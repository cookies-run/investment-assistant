package main

import (
	"bufio"
	"net/http"
	"os"
	"stock-monitor/internal/api"
	"stock-monitor/internal/repository"
	"stock-monitor/internal/schedule"
	"stock-monitor/pkg/envcrypt"
	"stock-monitor/pkg/logger"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Obfuscated passphrase used to decrypt .env.enc at startup.
func envPassphrase() string {
	return "Inv" + "est" + "ment" + "Asst" + "2026" + "Secure"
}

func loadEncryptedEnv() {
	encPath := ".env.enc"
	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		_ = godotenv.Load()
		return
	}
	encData, err := os.ReadFile(encPath)
	if err != nil {
		logger.Log.Warn("failed to read .env.enc", zap.Error(err))
		return
	}
	plain, err := envcrypt.Decrypt(encData, envPassphrase())
	if err != nil {
		logger.Log.Warn("failed to decrypt .env.enc", zap.Error(err))
		return
	}
	scanner := bufio.NewScanner(strings.NewReader(string(plain)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			_ = os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	logger.Log.Info("loaded encrypted env", zap.String("file", encPath))
}

func main() {
	logger.Init()
	defer logger.Sync()

	loadEncryptedEnv()

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
		frontendDist := os.Getenv("FRONTEND_DIST")
		if frontendDist == "" {
			frontendDist = "../frontend/dist"
		}
		r.Static("/assets", frontendDist+"/assets")
		r.GET("/", func(c *gin.Context) {
			c.File(frontendDist + "/index.html")
		})
		r.NoRoute(func(c *gin.Context) {
			c.File(frontendDist + "/index.html")
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
