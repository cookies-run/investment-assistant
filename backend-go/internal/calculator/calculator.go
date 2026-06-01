package calculator

import (
	"github.com/shopspring/decimal"
)

func CalcStockChange(currentPrice, prevClose decimal.Decimal) decimal.Decimal {
	if prevClose.IsZero() {
		return decimal.Zero
	}
	return currentPrice.Sub(prevClose).Div(prevClose).Mul(decimal.NewFromInt(100)).Round(2)
}

type StockPnL struct {
	MarketValue    decimal.Decimal `json:"market_value"`
	TotalPnL       decimal.Decimal `json:"total_pnl"`
	TotalPnLPercent decimal.Decimal `json:"total_pnl_percent"`
}

func CalcStockPnL(buyPrice, currentPrice decimal.Decimal, quantity decimal.Decimal) StockPnL {
	q := quantity
	marketValue := currentPrice.Mul(q).Round(3)
	totalCost := buyPrice.Mul(q)
	totalPnL := marketValue.Sub(totalCost).Round(3)
	totalPnLPercent := decimal.Zero
	if !totalCost.IsZero() {
		totalPnLPercent = totalPnL.Div(totalCost).Mul(decimal.NewFromInt(100)).Round(2)
	}
	return StockPnL{
		MarketValue:     marketValue,
		TotalPnL:        totalPnL,
		TotalPnLPercent: totalPnLPercent,
	}
}

func CalcFundEstimatedChange(holdings []struct {
	ChangePercent decimal.Decimal
	HoldRatio     decimal.Decimal
}) decimal.Decimal {
	sum := decimal.Zero
	for _, h := range holdings {
		sum = sum.Add(h.ChangePercent.Mul(h.HoldRatio).Div(decimal.NewFromInt(100)))
	}
	return sum.Round(2)
}

func CalcCumulativeChange(records []decimal.Decimal) decimal.Decimal {
	sum := decimal.Zero
	for _, r := range records {
		sum = sum.Add(r)
	}
	return sum.Round(2)
}
