package repository

import (
	"stock-monitor/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type FundLotRepo struct {
	db *gorm.DB
}

func NewFundLotRepo(db *gorm.DB) *FundLotRepo {
	return &FundLotRepo{db: db}
}

func (r *FundLotRepo) GetByFundCode(code string) ([]model.FundLot, error) {
	var lots []model.FundLot
	err := r.db.Where("fund_code = ?", code).Order("purchased_at asc").Find(&lots).Error
	return lots, err
}

func (r *FundLotRepo) Create(lot *model.FundLot) error {
	return r.db.Create(lot).Error
}

func (r *FundLotRepo) Delete(id uint) error {
	return r.db.Delete(&model.FundLot{}, id).Error
}

func (r *FundLotRepo) GetByID(id uint) (*model.FundLot, error) {
	var lot model.FundLot
	err := r.db.First(&lot, id).Error
	if err != nil {
		return nil, err
	}
	return &lot, nil
}

func (r *FundLotRepo) GetTotalQuantity(code string) (decimal.Decimal, error) {
	var result struct {
		Total decimal.Decimal
	}
	err := r.db.Model(&model.FundLot{}).
		Select("COALESCE(SUM(quantity), 0) as total").
		Where("fund_code = ?", code).
		Scan(&result).Error
	return result.Total, err
}
