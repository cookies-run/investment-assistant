package repository

import (
	"stock-monitor/internal/model"

	"gorm.io/gorm"
)

type FundRepo struct {
	db *gorm.DB
}

func NewFundRepo(db *gorm.DB) *FundRepo {
	return &FundRepo{db: db}
}

func (r *FundRepo) GetAll() ([]model.Fund, error) {
	var funds []model.Fund
	err := r.db.Find(&funds).Error
	return funds, err
}

func (r *FundRepo) GetByCode(code string) (*model.Fund, error) {
	var fund model.Fund
	err := r.db.First(&fund, "fund_code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &fund, nil
}

func (r *FundRepo) Create(fund *model.Fund) error {
	return r.db.Create(fund).Error
}

func (r *FundRepo) Update(code string, updates map[string]interface{}) error {
	// 设计说明：Updates 接收前端传来的 map，其中可能包含 ListWithStats 注入的衍生字段
	//（如 current_nav、change_percent）。这些字段不在 funds 表中，直接传入会导致
	// "no such column" 错误。此处通过白名单过滤，仅保留模型真实存在的列。
	allowed := map[string]bool{
		"fund_code":              true,
		"fund_name":              true,
		"hold_quantity":          true,
		"hold_cost":              true,
		"daily_profit_line":      true,
		"daily_loss_line":        true,
		"cumulative_profit_line": true,
		"cumulative_loss_line":   true,
		"cumulative_days":        true,
		"long_term_profit_line":  true,
		"long_term_loss_line":    true,
		"capital_scale_preset":   true,
		"fund_type":              true,
		"related_index_symbol":   true,
		"base_currency":          true,
		"is_active":              true,
		"last_rebalanced_at":     true,
	}
	filtered := make(map[string]interface{}, len(updates))
	for k, v := range updates {
		if allowed[k] {
			filtered[k] = v
		}
	}
	return r.db.Model(&model.Fund{}).Where("fund_code = ?", code).Updates(filtered).Error
}

func (r *FundRepo) Delete(code string) error {
	return r.db.Delete(&model.Fund{}, "fund_code = ?", code).Error
}

func (r *FundRepo) GetActive() ([]model.Fund, error) {
	var funds []model.Fund
	err := r.db.Where("is_active = ?", true).Find(&funds).Error
	return funds, err
}
