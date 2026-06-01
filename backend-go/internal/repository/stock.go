package repository

import (
	"stock-monitor/internal/model"

	"gorm.io/gorm"
)

type StockRepo struct {
	db *gorm.DB
}

func NewStockRepo(db *gorm.DB) *StockRepo {
	return &StockRepo{db: db}
}

func (r *StockRepo) GetAll() ([]model.Stock, error) {
	var stocks []model.Stock
	err := r.db.Find(&stocks).Error
	return stocks, err
}

func (r *StockRepo) GetByCode(code string) (*model.Stock, error) {
	var stock model.Stock
	err := r.db.First(&stock, "stock_code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func (r *StockRepo) Create(stock *model.Stock) error {
	return r.db.Create(stock).Error
}

func (r *StockRepo) Update(code string, updates map[string]interface{}) error {
	allowed := map[string]bool{
		"stock_code": true, "stock_name": true, "buy_price": true, "hold_quantity": true,
		"daily_profit_line": true, "daily_loss_line": true, "cumulative_profit_line": true,
		"cumulative_loss_line": true, "cumulative_days": true, "monitor_interval": true,
		"is_active": true,
	}
	filtered := make(map[string]interface{})
	for k, v := range updates {
		if allowed[k] {
			filtered[k] = v
		}
	}
	return r.db.Model(&model.Stock{}).Where("stock_code = ?", code).Updates(filtered).Error
}

func (r *StockRepo) Delete(code string) error {
	return r.db.Delete(&model.Stock{}, "stock_code = ?", code).Error
}

func (r *StockRepo) GetActive() ([]model.Stock, error) {
	var stocks []model.Stock
	err := r.db.Where("is_active = ?", true).Find(&stocks).Error
	return stocks, err
}
