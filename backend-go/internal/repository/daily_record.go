package repository

import (
	"strconv"
	"stock-monitor/internal/model"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DailyRecordRepo struct {
	db *gorm.DB
}

func NewDailyRecordRepo(db *gorm.DB) *DailyRecordRepo {
	return &DailyRecordRepo{db: db}
}

func (r *DailyRecordRepo) GetByTarget(targetCode string, targetType model.TargetType, limit int) ([]model.DailyRecord, error) {
	var records []model.DailyRecord
	query := r.db.Where("target_code = ? AND target_type = ?", targetCode, targetType).Order("trade_date desc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&records).Error
	return records, err
}

func (r *DailyRecordRepo) GetRecent(targetCode string, targetType model.TargetType, days int) ([]model.DailyRecord, error) {
	var records []model.DailyRecord
	lookback := days * 2
	err := r.db.Where("target_code = ? AND target_type = ? AND trade_date >= date('now', '-"+strconv.Itoa(lookback)+" days')", targetCode, targetType).
		Order("trade_date desc").Limit(days).Find(&records).Error
	return records, err
}

func (r *DailyRecordRepo) CreateOrUpdate(record *model.DailyRecord) error {
	var existing model.DailyRecord
	err := r.db.Where("target_code = ? AND target_type = ? AND trade_date = ?", record.TargetCode, record.TargetType, record.TradeDate).First(&existing).Error
	if err == nil {
		return r.db.Model(&existing).Updates(map[string]interface{}{
			"open_price":         record.OpenPrice,
			"close_price":        record.ClosePrice,
			"change_percent":     record.ChangePercent,
			"cumulative_percent": record.CumulativePercent,
		}).Error
	}
	return r.db.Create(record).Error
}

func (r *DailyRecordRepo) UpdateClosePrice(targetCode string, targetType model.TargetType, tradeDate time.Time, closePrice decimal.Decimal) error {
	return r.db.Model(&model.DailyRecord{}).
		Where("target_code = ? AND target_type = ? AND trade_date = ?", targetCode, targetType, tradeDate).
		Update("close_price", closePrice).Error
}

func (r *DailyRecordRepo) GetSince(targetCode string, targetType model.TargetType, since time.Time) ([]model.DailyRecord, error) {
	var records []model.DailyRecord
	err := r.db.Where("target_code = ? AND target_type = ? AND trade_date >= ?", targetCode, targetType, since).
		Order("trade_date desc").Find(&records).Error
	return records, err
}
