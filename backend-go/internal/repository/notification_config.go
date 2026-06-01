package repository

import (
	"stock-monitor/internal/model"

	"gorm.io/gorm"
)

type NotificationConfigRepo struct {
	db *gorm.DB
}

func NewNotificationConfigRepo(db *gorm.DB) *NotificationConfigRepo {
	return &NotificationConfigRepo{db: db}
}

func (r *NotificationConfigRepo) Get() (*model.NotificationConfig, error) {
	var cfg model.NotificationConfig
	err := r.db.First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &model.NotificationConfig{}, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *NotificationConfigRepo) Save(cfg *model.NotificationConfig) error {
	var existing model.NotificationConfig
	if err := r.db.First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		return r.db.Save(cfg).Error
	}
	return r.db.Create(cfg).Error
}
