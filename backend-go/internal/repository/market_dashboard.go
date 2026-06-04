package repository

import (
	"stock-monitor/internal/model"

	"gorm.io/gorm"
)

type MarketIndexGroupRepo struct {
	db *gorm.DB
}

func NewMarketIndexGroupRepo(db *gorm.DB) *MarketIndexGroupRepo {
	return &MarketIndexGroupRepo{db: db}
}

func (r *MarketIndexGroupRepo) GetAllWithItems(userID uint) ([]model.MarketIndexGroup, error) {
	var groups []model.MarketIndexGroup
	err := r.db.Where("user_id = ?", userID).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID).Order("sort_order ASC, id ASC")
	}).Order("sort_order ASC, id ASC").Find(&groups).Error
	return groups, err
}

func (r *MarketIndexGroupRepo) GetByID(userID, id uint) (*model.MarketIndexGroup, error) {
	var group model.MarketIndexGroup
	err := r.db.Where("user_id = ?", userID).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}).First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *MarketIndexGroupRepo) Create(group *model.MarketIndexGroup) error {
	return r.db.Create(group).Error
}

func (r *MarketIndexGroupRepo) Update(userID, id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.MarketIndexGroup{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates).Error
}

func (r *MarketIndexGroupRepo) Delete(userID, id uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.MarketIndexGroup{}).Error
}

func (r *MarketIndexGroupRepo) Reorder(userID uint, ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.MarketIndexGroup{}).Where("id = ? AND user_id = ?", id, userID).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type MarketIndexItemRepo struct {
	db *gorm.DB
}

func NewMarketIndexItemRepo(db *gorm.DB) *MarketIndexItemRepo {
	return &MarketIndexItemRepo{db: db}
}

func (r *MarketIndexItemRepo) Create(item *model.MarketIndexItem) error {
	return r.db.Create(item).Error
}

func (r *MarketIndexItemRepo) Delete(userID, id uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.MarketIndexItem{}).Error
}

func (r *MarketIndexItemRepo) Reorder(userID uint, ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.MarketIndexItem{}).Where("id = ? AND user_id = ?", id, userID).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type DailyCloseRepo struct {
	db *gorm.DB
}

func NewDailyCloseRepo(db *gorm.DB) *DailyCloseRepo {
	return &DailyCloseRepo{db: db}
}

func (r *DailyCloseRepo) Save(symbol, date string, closePrice float64, sourceType string) error {
	var existing model.DailyClose
	err := r.db.Where("symbol = ? AND date = ?", symbol, date).First(&existing).Error
	if err == nil {
		existing.ClosePrice = closePrice
		existing.SourceType = sourceType
		return r.db.Save(&existing).Error
	}
	return r.db.Create(&model.DailyClose{
		Symbol:     symbol,
		Date:       date,
		ClosePrice: closePrice,
		SourceType: sourceType,
	}).Error
}

func (r *DailyCloseRepo) GetLastNDays(symbol string, n int) ([]model.DailyClose, error) {
	var closes []model.DailyClose
	err := r.db.Where("symbol = ?", symbol).Order("date DESC").Limit(n).Find(&closes).Error
	return closes, err
}
