package repository

import (
	"stock-monitor/internal/model"

	"gorm.io/gorm"
)

type StockGroupRepo struct {
	db *gorm.DB
}

func NewStockGroupRepo(db *gorm.DB) *StockGroupRepo {
	return &StockGroupRepo{db: db}
}

func (r *StockGroupRepo) GetAllWithItems() ([]model.StockGroup, error) {
	var groups []model.StockGroup
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).Order("sort_order ASC, id ASC").Find(&groups).Error
	return groups, err
}

func (r *StockGroupRepo) GetByID(id uint) (*model.StockGroup, error) {
	var group model.StockGroup
	err := r.db.Preload("Items").First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *StockGroupRepo) Create(group *model.StockGroup) error {
	return r.db.Create(group).Error
}

func (r *StockGroupRepo) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.StockGroup{}).Where("id = ?", id).Updates(updates).Error
}

func (r *StockGroupRepo) Delete(id uint) error {
	return r.db.Delete(&model.StockGroup{}, id).Error
}

func (r *StockGroupRepo) Reorder(ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.StockGroup{}).Where("id = ?", id).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type StockGroupItemRepo struct {
	db *gorm.DB
}

func NewStockGroupItemRepo(db *gorm.DB) *StockGroupItemRepo {
	return &StockGroupItemRepo{db: db}
}

func (r *StockGroupItemRepo) Create(item *model.StockGroupItem) error {
	return r.db.Create(item).Error
}

func (r *StockGroupItemRepo) Delete(id uint) error {
	return r.db.Delete(&model.StockGroupItem{}, id).Error
}

func (r *StockGroupItemRepo) Reorder(ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.StockGroupItem{}).Where("id = ?", id).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
