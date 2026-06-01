package service

import (
	"stock-monitor/internal/datasource"
	"stock-monitor/internal/repository"

	"github.com/shopspring/decimal"
)

type StatsService struct {
	stockRepo *repository.StockRepo
	fundRepo  *repository.FundRepo
}

func NewStatsService(stockRepo *repository.StockRepo, fundRepo *repository.FundRepo) *StatsService {
	return &StatsService{stockRepo: stockRepo, fundRepo: fundRepo}
}

func (s *StatsService) GetPortfolioStats() (map[string]interface{}, error) {
	stocks, err := s.stockRepo.GetActive()
	if err != nil {
		return nil, err
	}
	funds, err := s.fundRepo.GetActive()
	if err != nil {
		return nil, err
	}

	totalCost := decimal.Zero
	totalMarketValue := decimal.Zero

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

	for _, st := range stocks {
		spot := spotData[st.StockCode]
		if spot.CurrentPrice.IsPositive() && st.BuyPrice.IsPositive() && st.HoldQuantity.IsPositive() {
			totalCost = totalCost.Add(st.BuyPrice.Mul(st.HoldQuantity))
			totalMarketValue = totalMarketValue.Add(spot.CurrentPrice.Mul(st.HoldQuantity))
		}
	}

	for _, f := range funds {
		est, _ := datasource.FetchFundEstimation(f.FundCode)
		if est != nil && est.EstNAV.IsPositive() && f.HoldCost.IsPositive() && f.HoldQuantity.IsPositive() {
			totalCost = totalCost.Add(f.HoldCost.Mul(f.HoldQuantity))
			totalMarketValue = totalMarketValue.Add(est.EstNAV.Mul(f.HoldQuantity))
		}
	}

	totalPnL := totalMarketValue.Sub(totalCost)
	totalPnLPercent := decimal.Zero
	if !totalCost.IsZero() {
		totalPnLPercent = totalPnL.Div(totalCost).Mul(decimal.NewFromInt(100)).Round(2)
	}

	return map[string]interface{}{
		"total_cost":         totalCost.Round(3),
		"total_market_value": totalMarketValue.Round(3),
		"total_pnl":          totalPnL.Round(3),
		"total_pnl_percent":  totalPnLPercent,
		"stock_count":        len(stocks),
		"fund_count":         len(funds),
	}, nil
}
