package service

import (
	"stock-monitor/internal/datasource"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"

	"github.com/shopspring/decimal"
)

type FundService struct {
	repo *repository.FundRepo
}

func NewFundService(repo *repository.FundRepo) *FundService {
	return &FundService{repo: repo}
}

func (s *FundService) ListWithStats() ([]map[string]interface{}, error) {
	funds, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(funds))
	for _, f := range funds {
		est, _ := datasource.FetchFundEstimation(f.FundCode)
		currentNAV := decimal.Zero
		changePercent := decimal.Zero
		if est != nil {
			currentNAV = est.EstNAV
			changePercent = est.EstChange
		}

		result := map[string]interface{}{
			"fund_code":              f.FundCode,
			"fund_name":              f.FundName,
			"hold_cost":              f.HoldCost,
			"hold_quantity":          f.HoldQuantity,
			"daily_profit_line":      f.DailyProfitLine,
			"daily_loss_line":        f.DailyLossLine,
			"cumulative_profit_line": f.CumulativeProfitLine,
			"cumulative_loss_line":   f.CumulativeLossLine,
			"cumulative_days":        f.CumulativeDays,
			"long_term_profit_line":  f.LongTermProfitLine,
			"long_term_loss_line":    f.LongTermLossLine,
			"capital_scale_preset":   f.CapitalScalePreset,
			"fund_type":              f.FundType,
			"related_index_symbol":   f.RelatedIndexSymbol,
			"base_currency":          f.BaseCurrency,
			"is_active":              f.IsActive,
			"created_at":             f.CreatedAt,
			"updated_at":             f.UpdatedAt,
			"current_nav":            nil,
			"change_percent":         nil,
			"market_value":           nil,
			"total_pnl":              nil,
			"total_pnl_percent":      nil,
		}

		if currentNAV.IsPositive() {
			result["current_nav"] = currentNAV
			result["change_percent"] = changePercent
			if f.HoldCost.IsPositive() && f.HoldQuantity.IsPositive() {
				marketValue := currentNAV.Mul(f.HoldQuantity).Round(3)
				totalCost := f.HoldCost.Mul(f.HoldQuantity)
				totalPnL := marketValue.Sub(totalCost).Round(3)
				totalPnLPercent := decimal.Zero
				if !totalCost.IsZero() {
					totalPnLPercent = totalPnL.Div(totalCost).Mul(decimal.NewFromInt(100)).Round(2)
				}
				result["market_value"] = marketValue
				result["total_pnl"] = totalPnL
				result["total_pnl_percent"] = totalPnLPercent
			}
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *FundService) Create(fund *model.Fund) error {
	return s.repo.Create(fund)
}

func (s *FundService) Update(code string, updates map[string]interface{}) error {
	return s.repo.Update(code, updates)
}

func (s *FundService) Delete(code string) error {
	return s.repo.Delete(code)
}

func (s *FundService) GetByCode(code string) (*model.Fund, error) {
	return s.repo.GetByCode(code)
}
