package service

import (
	"stock-monitor/internal/calculator"
	"stock-monitor/internal/datasource"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
)

type StockService struct {
	repo       *repository.StockRepo
	recordRepo *repository.DailyRecordRepo
}

func NewStockService(repo *repository.StockRepo, recordRepo *repository.DailyRecordRepo) *StockService {
	return &StockService{repo: repo, recordRepo: recordRepo}
}

func (s *StockService) ListWithStats() ([]map[string]interface{}, error) {
	stocks, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(stocks))
	for _, st := range stocks {
		codes = append(codes, st.StockCode)
	}

	spotData := make(map[string]datasource.StockSpot)
	if len(codes) > 0 {
		spots, err := datasource.FetchStockSpot(codes)
		if err == nil {
			for _, sp := range spots {
				spotData[sp.Code] = sp
			}
		}
	}

	results := make([]map[string]interface{}, 0, len(stocks))
	for _, st := range stocks {
		spot := spotData[st.StockCode]
		result := map[string]interface{}{
			"stock_code":             st.StockCode,
			"stock_name":             st.StockName,
			"buy_price":              st.BuyPrice,
			"hold_quantity":          st.HoldQuantity,
			"daily_profit_line":      st.DailyProfitLine,
			"daily_loss_line":        st.DailyLossLine,
			"cumulative_profit_line": st.CumulativeProfitLine,
			"cumulative_loss_line":   st.CumulativeLossLine,
			"cumulative_days":        st.CumulativeDays,
			"monitor_interval":       st.MonitorInterval,
			"is_active":              st.IsActive,
			"created_at":             st.CreatedAt,
			"updated_at":             st.UpdatedAt,
			"current_price":          nil,
			"change_percent":         nil,
			"market_value":           nil,
			"total_pnl":              nil,
			"total_pnl_percent":      nil,
			"volume":                 nil,
			"buy_volume":             nil,
			"sell_volume":            nil,
			"buy_sell_diff":          nil,
			"trends":                 nil,
		}

		if spot.CurrentPrice.IsPositive() {
			result["current_price"] = spot.CurrentPrice
			result["change_percent"] = spot.ChangePercent
			result["volume"] = spot.Volume
			result["buy_volume"] = spot.BuyVolume
			result["sell_volume"] = spot.SellVolume
			result["buy_sell_diff"] = spot.BuySellDiff
			if len(spot.Trends) > 0 {
				result["trends"] = spot.Trends
			} else {
				trends, _ := datasource.FetchStockTrends(st.StockCode)
				if len(trends) > 0 {
					result["trends"] = trends
				}
			}
			if st.BuyPrice.IsPositive() && st.HoldQuantity.IsPositive() {
				pnl := calculator.CalcStockPnL(st.BuyPrice, spot.CurrentPrice, st.HoldQuantity)
				result["market_value"] = pnl.MarketValue
				result["total_pnl"] = pnl.TotalPnL
				result["total_pnl_percent"] = pnl.TotalPnLPercent
			}
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *StockService) Create(stock *model.Stock) error {
	return s.repo.Create(stock)
}

func (s *StockService) Update(code string, updates map[string]interface{}) error {
	return s.repo.Update(code, updates)
}

func (s *StockService) Delete(code string) error {
	return s.repo.Delete(code)
}

func (s *StockService) GetByCode(code string) (*model.Stock, error) {
	return s.repo.GetByCode(code)
}
