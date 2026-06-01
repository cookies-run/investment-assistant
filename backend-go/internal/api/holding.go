package api

import (
	"net/http"
	"stock-monitor/internal/datasource"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type HoldingHandler struct {
	repo            *repository.HoldingRepo
	fundRepo        *repository.FundRepo
	lotRepo         *repository.FundLotRepo
	dailyRecordRepo *repository.DailyRecordRepo
}

func NewHoldingHandler(repo *repository.HoldingRepo, fundRepo *repository.FundRepo, lotRepo *repository.FundLotRepo, dailyRecordRepo *repository.DailyRecordRepo) *HoldingHandler {
	return &HoldingHandler{repo: repo, fundRepo: fundRepo, lotRepo: lotRepo, dailyRecordRepo: dailyRecordRepo}
}

func (h *HoldingHandler) Get(c *gin.Context) {
	code := c.Param("code")
	holdings, err := h.repo.GetByFundCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, holdings)
}

type holdingDetailItem struct {
	StockCode    string          `json:"stock_code"`
	StockName    string          `json:"stock_name"`
	HoldRatio    decimal.Decimal `json:"hold_ratio"`
	CurrentPrice decimal.Decimal `json:"current_price"`
	ChangePercent decimal.Decimal `json:"change_percent"`
	Contribution decimal.Decimal `json:"contribution"`
}

type lotSummary struct {
	TotalQuantity      decimal.Decimal `json:"total_quantity"`
	FeeFreeQuantity    decimal.Decimal `json:"fee_free_quantity"`    // >= 30 days
	FeeReducedQuantity decimal.Decimal `json:"fee_reduced_quantity"` // 7~30 days
	FeePenaltyQuantity decimal.Decimal `json:"fee_penalty_quantity"` // < 7 days
	Lots               []model.FundLot `json:"lots"`
}

// holdingDetailResponse 是基金详情接口的核心返回结构。
// 设计说明：本结构承载了 14:50 量化决策引擎所需的全部输入变量，
// 分为四个维度：基金实时数据、用户持仓状态、大盘趋势代理变量、用户策略配置。
// 后端在此一次性计算好所有衍生指标（如滚动复合收益、总收益率、最小持有天数），
// 前端决策树只需做纯逻辑判断，不做复杂计算，确保决策引擎与 UI 完全解耦。
type holdingDetailResponse struct {
	FundCode                string              `json:"fund_code"`
	FundName                string              `json:"fund_name"`
	HoldCost                decimal.Decimal     `json:"hold_cost"`
	HoldQuantity            decimal.Decimal     `json:"hold_quantity"`
	DailyProfitLine         decimal.Decimal     `json:"daily_profit_line"`
	DailyLossLine           decimal.Decimal     `json:"daily_loss_line"`
	CumulativeProfitLine    decimal.Decimal     `json:"cumulative_profit_line"`
	CumulativeLossLine      decimal.Decimal     `json:"cumulative_loss_line"`
	CumulativeDays          int                 `json:"cumulative_days"`
	LongTermProfitLine      decimal.Decimal     `json:"long_term_profit_line"`
	LongTermLossLine        decimal.Decimal     `json:"long_term_loss_line"`
	CapitalScalePreset      string              `json:"capital_scale_preset"`
	FundType                string              `json:"fund_type"`
	RelatedIndexSymbol      string              `json:"related_index_symbol"`
	BaseCurrency            string              `json:"base_currency"`
	IsActive                bool                `json:"is_active"`
	EstimatedChange         decimal.Decimal     `json:"estimated_change"`
	ActualChange            decimal.Decimal     `json:"actual_change"`
	Deviation               decimal.Decimal     `json:"deviation"`
	CurrentNav              decimal.Decimal     `json:"current_nav"`
	TotalReturnEst          decimal.Decimal     `json:"total_return_est"`
	MinHoldDays             int                 `json:"min_hold_days"`
	RollingCumulativeReturn decimal.Decimal     `json:"rolling_cumulative_return"`
	RelatedIndexReturn      decimal.Decimal     `json:"related_index_return"`
	FxDailyChange           decimal.Decimal     `json:"fx_daily_change"`
	TrackingErrorIndex      decimal.Decimal     `json:"tracking_error_index"`
	EstDailyChange          decimal.Decimal     `json:"est_daily_change"`
	SyncedAt                time.Time           `json:"synced_at"`
	CreatedAt               time.Time           `json:"created_at"`
	ReportDate              time.Time           `json:"report_date"`
	LotSummary              lotSummary          `json:"lot_summary"`
	Holdings                []holdingDetailItem `json:"holdings"`
}

func (h *HoldingHandler) Detail(c *gin.Context) {
	code := c.Param("code")

	fund, err := h.fundRepo.GetByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fund not found"})
		return
	}

	holdings, err := h.repo.GetByFundCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var stockCodes []string
	for _, h := range holdings {
		stockCodes = append(stockCodes, h.StockCode)
	}

	spotMap := make(map[string]datasource.StockSpot)
	if len(stockCodes) > 0 {
		spots, _ := datasource.FetchStockSpot(stockCodes)
		for _, s := range spots {
			spotMap[s.Code] = s
		}
	}

	items := make([]holdingDetailItem, 0)
	var estimatedChange decimal.Decimal

	for _, holding := range holdings {
		item := holdingDetailItem{
			StockCode: holding.StockCode,
			StockName: holding.StockName,
			HoldRatio: holding.HoldRatio,
		}

		if spot, ok := spotMap[holding.StockCode]; ok {
			item.CurrentPrice = spot.CurrentPrice
			item.ChangePercent = spot.ChangePercent
			// contribution = hold_ratio * change_percent / 100
			item.Contribution = holding.HoldRatio.Mul(spot.ChangePercent).Div(decimal.NewFromInt(100)).Round(4)
			estimatedChange = estimatedChange.Add(item.Contribution)
		}

		items = append(items, item)
	}

	// Fetch actual fund estimation for deviation calculation
	actualChange := decimal.Zero
	currentNav := decimal.Zero
	est, _ := datasource.FetchFundEstimation(fund.FundCode)
	if est != nil {
		actualChange = est.EstChange
		currentNav = est.EstNAV
	}
	deviation := estimatedChange.Sub(actualChange).Round(4)

	// ═══════════════════════════════════════════════════════
	// 指数基金模式：以关联指数实时行情 + 汇率折算替代基金自身估值
	// ═══════════════════════════════════════════════════════
	relatedIndexReturn := decimal.Zero
	fxDailyChange := decimal.Zero
	trackingErrorIndex := decimal.Zero
	estDailyChange := actualChange

	if fund.FundType == "INDEX" && fund.RelatedIndexSymbol != "" {
		idxItem, _ := datasource.FetchIndexRealtime(fund.RelatedIndexSymbol)
		relatedIndexReturn = idxItem.ChangePercent
		if fund.BaseCurrency != "" && fund.BaseCurrency != "CNY" {
			fxDailyChange, _ = datasource.FetchFXChange(fund.BaseCurrency)
		}
		// est_daily_change = index_return + fx_change * (-1)
		// 人民币升值 → fx_change > 0 → est 被拉低；人民币贬值 → fx_change < 0 → est 被抬高
		estDailyChange = relatedIndexReturn.Add(fxDailyChange.Mul(decimal.NewFromInt(-1))).Round(4)
		trackingErrorIndex = actualChange.Sub(estDailyChange).Round(4)
	}

	// Calculate lot summary for redemption fee analysis
	lotSummary := calcLotSummary(h.lotRepo, code)

	// Calculate min_hold_days: 最近一笔买入（purchased_at 最新）的持有天数
	minHoldDays := 0
	for _, lot := range lotSummary.Lots {
		days := int(time.Now().Sub(lot.PurchasedAt).Hours() / 24)
		if days < minHoldDays || minHoldDays == 0 {
			minHoldDays = days
		}
	}

	// Calculate total_return_est = (currentNav - holdCost) / holdCost * 100
	totalReturnEst := decimal.Zero
	if fund.HoldCost.GreaterThan(decimal.Zero) && currentNav.GreaterThan(decimal.Zero) {
		totalReturnEst = currentNav.Sub(fund.HoldCost).Div(fund.HoldCost).Mul(decimal.NewFromInt(100)).Round(4)
	}

	// Calculate rolling cumulative return (compound): prod(1 + r_i/100) - 1, then *100
	rollingCumulativeReturn := decimal.Zero
	if fund.CumulativeDays > 0 {
		recentRecords, _ := h.dailyRecordRepo.GetRecent(code, model.TargetTypeFund, fund.CumulativeDays)
		if len(recentRecords) > 0 {
			product := decimal.NewFromFloat(1.0)
			for _, r := range recentRecords {
				rDecimal := r.ChangePercent.Div(decimal.NewFromInt(100))
				product = product.Mul(decimal.NewFromFloat(1.0).Add(rDecimal))
			}
			rollingCumulativeReturn = product.Sub(decimal.NewFromFloat(1.0)).Mul(decimal.NewFromInt(100)).Round(4)
		}
	}

	// 设计说明：Detail 接口是 14:50 决策引擎的唯一数据源。
	// 所有需要前端做判断的衍生指标（total_return_est, min_hold_days, rolling_cumulative_return）
	// 都在后端一次性计算完毕，前端 analyze 函数只做纯逻辑分支判断，不触及复杂数学计算。
	resp := holdingDetailResponse{
		FundCode:                fund.FundCode,
		FundName:                fund.FundName,
		HoldCost:                fund.HoldCost,
		HoldQuantity:            fund.HoldQuantity,
		DailyProfitLine:         fund.DailyProfitLine,
		DailyLossLine:           fund.DailyLossLine,
		CumulativeProfitLine:    fund.CumulativeProfitLine,
		CumulativeLossLine:      fund.CumulativeLossLine,
		CumulativeDays:          fund.CumulativeDays,
		LongTermProfitLine:      fund.LongTermProfitLine,
		LongTermLossLine:        fund.LongTermLossLine,
		CapitalScalePreset:      fund.CapitalScalePreset,
		FundType:                fund.FundType,
		RelatedIndexSymbol:      fund.RelatedIndexSymbol,
		BaseCurrency:            fund.BaseCurrency,
		IsActive:                fund.IsActive,
		EstimatedChange:         estimatedChange.Round(4),
		ActualChange:            actualChange,
		Deviation:               deviation,
		CurrentNav:              currentNav,
		TotalReturnEst:          totalReturnEst,
		MinHoldDays:             minHoldDays,
		RollingCumulativeReturn: rollingCumulativeReturn,
		RelatedIndexReturn:      relatedIndexReturn,
		FxDailyChange:           fxDailyChange,
		TrackingErrorIndex:      trackingErrorIndex,
		EstDailyChange:          estDailyChange,
		SyncedAt:                time.Now(),
		CreatedAt:               fund.CreatedAt,
		LotSummary:              lotSummary,
		Holdings:                items,
	}
	if len(holdings) > 0 {
		resp.SyncedAt = holdings[0].SyncedAt
		resp.ReportDate = holdings[0].ReportDate
	}

	c.JSON(http.StatusOK, resp)
}

func calcLotSummary(lotRepo *repository.FundLotRepo, code string) lotSummary {
	lots, err := lotRepo.GetByFundCode(code)
	if err != nil {
		return lotSummary{Lots: []model.FundLot{}}
	}

	now := time.Now()
	var total, feeFree, feeReduced, feePenalty decimal.Decimal

	for _, lot := range lots {
		total = total.Add(lot.Quantity)
		days := int(now.Sub(lot.PurchasedAt).Hours() / 24)
		switch {
		case days >= 30:
			feeFree = feeFree.Add(lot.Quantity)
		case days >= 7:
			feeReduced = feeReduced.Add(lot.Quantity)
		default:
			feePenalty = feePenalty.Add(lot.Quantity)
		}
	}

	return lotSummary{
		TotalQuantity:      total,
		FeeFreeQuantity:    feeFree,
		FeeReducedQuantity: feeReduced,
		FeePenaltyQuantity: feePenalty,
		Lots:               lots,
	}
}

func (h *HoldingHandler) Sync(c *gin.Context) {
	code := c.Param("code")
	data, err := datasource.FetchFundHoldings(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No holdings data found"})
		return
	}

	var holdings []model.FundHolding
	for _, d := range data {
		holdings = append(holdings, model.FundHolding{
			FundCode:   d.FundCode,
			StockCode:  d.StockCode,
			StockName:  d.StockName,
			HoldRatio:  d.HoldRatio,
			ReportDate: parseDate(d.ReportDate),
		})
	}

	if err := h.repo.Sync(code, holdings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "count": len(holdings)})
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	if t.IsZero() {
		return time.Now()
	}
	return t
}
