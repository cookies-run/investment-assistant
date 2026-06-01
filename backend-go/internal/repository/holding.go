package repository

import (
	"stock-monitor/internal/model"

	"gorm.io/gorm"
)

type HoldingRepo struct {
	db *gorm.DB
}

func NewHoldingRepo(db *gorm.DB) *HoldingRepo {
	return &HoldingRepo{db: db}
}

func (r *HoldingRepo) GetByFundCode(fundCode string) ([]model.FundHolding, error) {
	var holdings []model.FundHolding
	err := r.db.Where("fund_code = ?", fundCode).Order("synced_at desc").Find(&holdings).Error
	return holdings, err
}

func (r *HoldingRepo) Sync(fundCode string, holdings []model.FundHolding) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("fund_code = ?", fundCode).Delete(&model.FundHolding{}).Error; err != nil {
			return err
		}
		if len(holdings) > 0 {
			if err := tx.CreateInBatches(holdings, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
