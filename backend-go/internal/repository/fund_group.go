package repository

import (
	"stock-monitor/internal/model"

	"gorm.io/gorm"
)

type FundGroupRepo struct {
	db *gorm.DB
}

func NewFundGroupRepo(db *gorm.DB) *FundGroupRepo {
	return &FundGroupRepo{db: db}
}

func (r *FundGroupRepo) GetAllWithItems(userID uint) ([]model.FundGroup, error) {
	var groups []model.FundGroup
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).Order("sort_order ASC, id ASC").Find(&groups).Error
	return groups, err
}

func (r *FundGroupRepo) GetByID(userID, id uint) (*model.FundGroup, error) {
	var group model.FundGroup
	err := r.db.Preload("Items").First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *FundGroupRepo) Create(group *model.FundGroup) error {
	return r.db.Create(group).Error
}

func (r *FundGroupRepo) Update(userID, id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.FundGroup{}).Where("id = ?", id).Updates(updates).Error
}

func (r *FundGroupRepo) Delete(userID, id uint) error {
	return r.db.Delete(&model.FundGroup{}, id).Error
}

func (r *FundGroupRepo) Reorder(userID uint, ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.FundGroup{}).Where("id = ?", id).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type FundGroupItemRepo struct {
	db *gorm.DB
}

func NewFundGroupItemRepo(db *gorm.DB) *FundGroupItemRepo {
	return &FundGroupItemRepo{db: db}
}

func (r *FundGroupItemRepo) Create(item *model.FundGroupItem) error {
	return r.db.Create(item).Error
}

func (r *FundGroupItemRepo) Delete(userID, id uint) error {
	return r.db.Delete(&model.FundGroupItem{}, id).Error
}

func (r *FundGroupItemRepo) Reorder(userID uint, ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.FundGroupItem{}).Where("id = ?", id).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
