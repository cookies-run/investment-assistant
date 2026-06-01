package repository

import (
	"stock-monitor/internal/model"
	"time"

	"gorm.io/gorm"
)

type AlertRepo struct {
	db *gorm.DB
}

func NewAlertRepo(db *gorm.DB) *AlertRepo {
	return &AlertRepo{db: db}
}

func (r *AlertRepo) GetAll(targetCode string, targetType model.TargetType, limit int) ([]model.AlertRecord, error) {
	var alerts []model.AlertRecord
	query := r.db.Order("triggered_at desc")
	if targetCode != "" {
		query = query.Where("target_code = ?", targetCode)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&alerts).Error
	return alerts, err
}

func (r *AlertRepo) Create(alert *model.AlertRecord) error {
	return r.db.Create(alert).Error
}

func (r *AlertRepo) HasAlertToday(targetCode string, targetType model.TargetType, alertType model.AlertType) (bool, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.Model(&model.AlertRecord{}).
		Where("target_code = ? AND target_type = ? AND alert_type = ? AND date(triggered_at) = ?", targetCode, targetType, alertType, today).
		Count(&count).Error
	return count > 0, err
}
