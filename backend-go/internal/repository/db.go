package repository

import (
	"os"
	"path/filepath"
	"stock-monitor/internal/model"
	"stock-monitor/pkg/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func resolveDBPath() string {
	// Priority 1: explicit DB_PATH from Tauri sidecar
	if env := os.Getenv("DB_PATH"); env != "" {
		return env
	}
	// Priority 2: legacy DATA_DIR override
	if env := os.Getenv("DATA_DIR"); env != "" {
		return filepath.Join(env, "stock_monitor.db")
	}
	// Priority 3: Tauri sidecar mode (auto-detect user config dir)
	if os.Getenv("TAURI_SIDECAR") == "1" || os.Getenv("TAURI_SIDEARCAR") == "1" {
		configDir, err := os.UserConfigDir()
		if err == nil {
			dir := filepath.Join(configDir, "投资助手")
			os.MkdirAll(dir, 0755)
			return filepath.Join(dir, "stock_monitor.db")
		}
	}
	// Standalone mode: relative to backend-go directory
	return "../data/stock_monitor.db"
}

func InitDB() *gorm.DB {
	var err error
	dbPath := resolveDBPath()
	logger.Log.Info("opening database", zap.String("path", dbPath))
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("failed to connect database", zap.Error(err))
	}

	if err := DB.AutoMigrate(
		&model.Stock{},
		&model.Fund{},
		&model.FundHolding{},
		&model.FundLot{},
		&model.DailyRecord{},
		&model.AlertRecord{},
		&model.MarketIndexGroup{},
		&model.MarketIndexItem{},
		&model.StockGroup{},
		&model.StockGroupItem{},
		&model.FundGroup{},
		&model.FundGroupItem{},
		&model.DailyClose{},
		&model.NotificationConfig{},
	); err != nil {
		logger.Log.Fatal("failed to migrate database", zap.Error(err))
	}

	// Backfill: create default FundLot for existing funds without lots
	backfillFundLots(DB)
	// Backfill: set default values for new Fund fields
	backfillFundDefaults(DB)

	return DB
}

func backfillFundDefaults(db *gorm.DB) {
	// Set default long_term_profit_line, long_term_loss_line, capital_scale_preset for existing funds
	err := db.Model(&model.Fund{}).Where("long_term_profit_line = 0").Update("long_term_profit_line", decimal.NewFromFloat(20.0)).Error
	if err != nil {
		logger.Log.Warn("failed to backfill long_term_profit_line", zap.Error(err))
	}
	err = db.Model(&model.Fund{}).Where("long_term_loss_line = 0").Update("long_term_loss_line", decimal.NewFromFloat(15.0)).Error
	if err != nil {
		logger.Log.Warn("failed to backfill long_term_loss_line", zap.Error(err))
	}
	err = db.Model(&model.Fund{}).Where("capital_scale_preset = ? OR capital_scale_preset = ?", "", "custom").Update("capital_scale_preset", "MEDIUM").Error
	if err != nil {
		logger.Log.Warn("failed to backfill capital_scale_preset", zap.Error(err))
	}
	// Backfill new fields for index fund support
	err = db.Model(&model.Fund{}).Where("fund_type = ? OR fund_type = ?", "", "custom").Update("fund_type", "ACTIVE").Error
	if err != nil {
		logger.Log.Warn("failed to backfill fund_type", zap.Error(err))
	}
	err = db.Model(&model.Fund{}).Where("base_currency = ? OR base_currency = ?", "", "custom").Update("base_currency", "CNY").Error
	if err != nil {
		logger.Log.Warn("failed to backfill base_currency", zap.Error(err))
	}
}

func backfillFundLots(db *gorm.DB) {
	var funds []model.Fund
	if err := db.Find(&funds).Error; err != nil {
		logger.Log.Warn("failed to fetch funds for lot backfill", zap.Error(err))
		return
	}
	for _, f := range funds {
		var count int64
		if err := db.Model(&model.FundLot{}).Where("fund_code = ?", f.FundCode).Count(&count).Error; err != nil {
			continue
		}
		if count > 0 {
			continue
		}
		lot := model.FundLot{
			FundCode:    f.FundCode,
			Quantity:    f.HoldQuantity,
			Cost:        f.HoldCost,
			PurchasedAt: f.CreatedAt,
		}
		if err := db.Create(&lot).Error; err != nil {
			logger.Log.Warn("failed to create default fund lot", zap.String("fund", f.FundCode), zap.Error(err))
		}
	}
}
