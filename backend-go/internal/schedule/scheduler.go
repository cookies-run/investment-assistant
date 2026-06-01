package schedule

import (
	"stock-monitor/internal/calculator"
	"stock-monitor/internal/datasource"
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"stock-monitor/internal/service"
	"stock-monitor/internal/strategy"
	"stock-monitor/pkg/logger"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Scheduler struct {
	cron *cron.Cron
	db   *gorm.DB
}

func New(db *gorm.DB) *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
		db:   db,
	}
}

func (s *Scheduler) Start() {
	// stock_monitor: every minute during trading hours
	s.cron.AddFunc("0 * 9-11,13-14 * * 1-5", s.jobStockMonitor)
	// fund_estimation: 14:50 on weekdays
	s.cron.AddFunc("0 50 14 * * 1-5", s.jobFundEstimation)
	// fund_sync: 15:30 on weekdays
	s.cron.AddFunc("0 30 15 * * 1-5", s.jobFundSync)
	// close_update: 15:05 on weekdays
	s.cron.AddFunc("0 5 15 * * 1-5", s.jobCloseUpdate)

	s.cron.Start()
	logger.Log.Info("Scheduler started")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) jobStockMonitor() {
	now := time.Now()
	if !isTradeTime(now) {
		return
	}

	stockRepo := repository.NewStockRepo(s.db)
	recordRepo := repository.NewDailyRecordRepo(s.db)
	alertRepo := repository.NewAlertRepo(s.db)
	notificationConfigRepo := repository.NewNotificationConfigRepo(s.db)
	notificationService := service.NewNotificationService(notificationConfigRepo)

	stocks, err := stockRepo.GetActive()
	if err != nil {
		logger.Log.Error("jobStockMonitor: get stocks failed", zap.Error(err))
		return
	}
	stocks = filterStocksWithHoldings(stocks)
	if len(stocks) == 0 {
		return
	}

	codes := make([]string, 0, len(stocks))
	for _, st := range stocks {
		codes = append(codes, st.StockCode)
	}

	spots, err := datasource.FetchStockSpot(codes)
	if err != nil {
		logger.Log.Error("jobStockMonitor: fetch spot failed", zap.Error(err))
		return
	}

	spotMap := make(map[string]datasource.StockSpot)
	for _, sp := range spots {
		spotMap[sp.Code] = sp
	}

	for _, st := range stocks {
		spot := spotMap[st.StockCode]
		if spot.Code == "" {
			continue
		}

		changePercent := spot.ChangePercent
		if changePercent.IsZero() && spot.CurrentPrice.IsPositive() && spot.PrevClose.IsPositive() {
			changePercent = calculator.CalcStockChange(spot.CurrentPrice, spot.PrevClose)
		}

		record := &model.DailyRecord{
			TargetCode:    st.StockCode,
			TargetType:    model.TargetTypeStock,
			TradeDate:     time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
			OpenPrice:     &spot.OpenPrice,
			ChangePercent: changePercent,
		}
		if err := recordRepo.CreateOrUpdate(record); err != nil {
			logger.Log.Error("jobStockMonitor: save daily record failed", zap.Error(err))
		}

		alert := strategy.CheckSingleDayRule(alertRepo, st.StockCode, model.TargetTypeStock, changePercent, &st)
		if alert != nil {
			alert.CurrentPrice = &spot.CurrentPrice
			if err := alertRepo.Create(alert); err != nil {
				logger.Log.Error("jobStockMonitor: create alert failed", zap.Error(err))
			} else {
				cp := spot.CurrentPrice.InexactFloat64()
				notificationService.SendFeishuAlert(alert, st.StockName, &cp)
			}
		}

		// cumulative check after 15:00
		if now.Hour() >= 15 {
			cumAlert := strategy.CheckCumulativeRule(alertRepo, recordRepo, st.StockCode, model.TargetTypeStock, changePercent, &st)
			if cumAlert != nil {
				cumAlert.CurrentPrice = &spot.CurrentPrice
				if err := alertRepo.Create(cumAlert); err != nil {
					logger.Log.Error("jobStockMonitor: create cumulative alert failed", zap.Error(err))
				} else {
					cp := spot.CurrentPrice.InexactFloat64()
					notificationService.SendFeishuAlert(cumAlert, st.StockName, &cp)
				}
			}
		}
	}
}

func (s *Scheduler) jobFundEstimation() {
	fundRepo := repository.NewFundRepo(s.db)
	recordRepo := repository.NewDailyRecordRepo(s.db)
	alertRepo := repository.NewAlertRepo(s.db)
	notificationConfigRepo := repository.NewNotificationConfigRepo(s.db)
	notificationService := service.NewNotificationService(notificationConfigRepo)

	funds, err := fundRepo.GetActive()
	if err != nil {
		logger.Log.Error("jobFundEstimation: get funds failed", zap.Error(err))
		return
	}
	funds = filterFundsWithHoldings(funds)
	if len(funds) == 0 {
		return
	}

	now := time.Now()
	for _, f := range funds {
		est, _ := datasource.FetchFundEstimation(f.FundCode)
		changePercent := decimal.Zero
		currentNAV := decimal.Zero

		if est != nil {
			changePercent = est.EstChange
			currentNAV = est.EstNAV
		} else {
			// fallback: calculate from holdings
			holdingRepo := repository.NewHoldingRepo(s.db)
			holdings, _ := holdingRepo.GetByFundCode(f.FundCode)
			if len(holdings) > 0 {
				codes := make([]string, 0, len(holdings))
				for _, h := range holdings {
					codes = append(codes, h.StockCode)
				}
				spots, _ := datasource.FetchStockSpot(codes)
				stockChanges := make(map[string]decimal.Decimal)
				for _, sp := range spots {
					stockChanges[sp.Code] = sp.ChangePercent
				}
				var calcHoldings []struct {
					ChangePercent decimal.Decimal
					HoldRatio     decimal.Decimal
				}
				for _, h := range holdings {
					if cp, ok := stockChanges[h.StockCode]; ok {
						calcHoldings = append(calcHoldings, struct {
							ChangePercent decimal.Decimal
							HoldRatio     decimal.Decimal
						}{cp, h.HoldRatio})
					}
				}
				changePercent = calculator.CalcFundEstimatedChange(calcHoldings)
			}
		}

		record := &model.DailyRecord{
			TargetCode:    f.FundCode,
			TargetType:    model.TargetTypeFund,
			TradeDate:     time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
			ChangePercent: changePercent,
		}
		if currentNAV.IsPositive() {
			record.ClosePrice = &currentNAV
		}
		if err := recordRepo.CreateOrUpdate(record); err != nil {
			logger.Log.Error("jobFundEstimation: save daily record failed", zap.Error(err))
		}

		alert := strategy.CheckSingleDayRule(alertRepo, f.FundCode, model.TargetTypeFund, changePercent, &f)
		if alert != nil {
			if currentNAV.IsPositive() {
				alert.CurrentPrice = &currentNAV
			}
			if err := alertRepo.Create(alert); err != nil {
				logger.Log.Error("jobFundEstimation: create alert failed", zap.Error(err))
			} else {
				cp := currentNAV.InexactFloat64()
				notificationService.SendFeishuAlert(alert, f.FundName, &cp)
			}
		}
	}
}

func (s *Scheduler) jobFundSync() {
	fundRepo := repository.NewFundRepo(s.db)
	holdingRepo := repository.NewHoldingRepo(s.db)

	funds, err := fundRepo.GetAll()
	if err != nil {
		logger.Log.Error("jobFundSync: get funds failed", zap.Error(err))
		return
	}

	for _, f := range funds {
		data, err := datasource.FetchFundHoldings(f.FundCode)
		if err != nil {
			logger.Log.Warn("jobFundSync: fetch holdings failed", zap.String("fund", f.FundCode), zap.Error(err))
			continue
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

		if err := holdingRepo.Sync(f.FundCode, holdings); err != nil {
			logger.Log.Error("jobFundSync: sync holdings failed", zap.String("fund", f.FundCode), zap.Error(err))
		}
	}
}

func (s *Scheduler) jobCloseUpdate() {
	stockRepo := repository.NewStockRepo(s.db)
	recordRepo := repository.NewDailyRecordRepo(s.db)

	stocks, err := stockRepo.GetActive()
	if err != nil {
		logger.Log.Error("jobCloseUpdate: get stocks failed", zap.Error(err))
		return
	}

	codes := make([]string, 0, len(stocks))
	for _, st := range stocks {
		codes = append(codes, st.StockCode)
	}

	spots, err := datasource.FetchStockSpot(codes)
	if err != nil {
		logger.Log.Error("jobCloseUpdate: fetch spot failed", zap.Error(err))
		return
	}

	now := time.Now()
	tradeDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	for _, sp := range spots {
		if sp.CurrentPrice.IsPositive() {
			if err := recordRepo.UpdateClosePrice(sp.Code, model.TargetTypeStock, tradeDate, sp.CurrentPrice); err != nil {
				logger.Log.Error("jobCloseUpdate: update close price failed", zap.String("code", sp.Code), zap.Error(err))
			}
		}
	}
}

func filterStocksWithHoldings(stocks []model.Stock) []model.Stock {
	filtered := make([]model.Stock, 0, len(stocks))
	for _, s := range stocks {
		if s.HoldQuantity.IsPositive() {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func filterFundsWithHoldings(funds []model.Fund) []model.Fund {
	filtered := make([]model.Fund, 0, len(funds))
	for _, f := range funds {
		if f.HoldQuantity.IsPositive() {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func isTradeTime(t time.Time) bool {
	if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		return false
	}
	hour, min := t.Hour(), t.Minute()
	if (hour == 9 && min >= 30) || (hour == 10) || (hour == 11 && min <= 30) {
		return true
	}
	if hour >= 13 && hour < 15 {
		return true
	}
	return false
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	if t.IsZero() {
		return time.Now()
	}
	return t
}
