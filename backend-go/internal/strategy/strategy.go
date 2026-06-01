package strategy

import (
	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"stock-monitor/pkg/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func CheckSingleDayRule(repo *repository.AlertRepo, targetCode string, targetType model.TargetType, changePercent decimal.Decimal, config interface{}) *model.AlertRecord {
	var dailyProfitLine, dailyLossLine decimal.Decimal
	switch c := config.(type) {
	case *model.Stock:
		dailyProfitLine = c.DailyProfitLine
		dailyLossLine = c.DailyLossLine
	case *model.Fund:
		dailyProfitLine = c.DailyProfitLine
		dailyLossLine = c.DailyLossLine
	default:
		return nil
	}

	var alertType model.AlertType
	var triggerValue decimal.Decimal

	if changePercent.GreaterThanOrEqual(dailyProfitLine) {
		alertType = model.AlertTypeSingleDayProfit
		triggerValue = changePercent
	} else if changePercent.LessThanOrEqual(dailyLossLine.Neg()) {
		alertType = model.AlertTypeSingleDayLoss
		triggerValue = changePercent
	} else {
		return nil
	}

	hasAlert, err := repo.HasAlertToday(targetCode, targetType, alertType)
	if err != nil {
		logger.Log.Error("check alert today failed", zap.Error(err))
		return nil
	}
	if hasAlert {
		return nil
	}

	return &model.AlertRecord{
		TargetCode:     targetCode,
		TargetType:     targetType,
		AlertType:      alertType,
		TriggerValue:   triggerValue,
		ThresholdValue: dailyProfitLine,
	}
}

func CheckCumulativeRule(repo *repository.AlertRepo, recordRepo *repository.DailyRecordRepo, targetCode string, targetType model.TargetType, currentChange decimal.Decimal, config interface{}) *model.AlertRecord {
	var cumulativeDays int
	var cumulativeProfitLine, cumulativeLossLine decimal.Decimal
	switch c := config.(type) {
	case *model.Stock:
		cumulativeDays = c.CumulativeDays
		cumulativeProfitLine = c.CumulativeProfitLine
		cumulativeLossLine = c.CumulativeLossLine
	case *model.Fund:
		cumulativeDays = c.CumulativeDays
		cumulativeProfitLine = c.CumulativeProfitLine
		cumulativeLossLine = c.CumulativeLossLine
	default:
		return nil
	}

	records, err := recordRepo.GetRecent(targetCode, targetType, cumulativeDays)
	if err != nil {
		logger.Log.Error("get recent daily records failed", zap.Error(err))
		return nil
	}

	var changes []decimal.Decimal
	if currentChange.IsZero() || (len(records) > 0 && records[0].TradeDate.Format("2006-01-02") != "") {
		// prepend current change if today's record not found
		// simplified: always prepend currentChange
	}
	changes = append(changes, currentChange)
	for _, r := range records {
		changes = append(changes, r.ChangePercent)
	}

	// limit to cumulativeDays
	if len(changes) > cumulativeDays {
		changes = changes[:cumulativeDays]
	}

	sum := decimal.Zero
	for _, c := range changes {
		sum = sum.Add(c)
	}

	var alertType model.AlertType
	var triggerValue decimal.Decimal

	if sum.GreaterThanOrEqual(cumulativeProfitLine) {
		alertType = model.AlertTypeCumulativeProfit
		triggerValue = sum
	} else if sum.LessThanOrEqual(cumulativeLossLine.Neg()) {
		alertType = model.AlertTypeCumulativeLoss
		triggerValue = sum
	} else {
		return nil
	}

	hasAlert, err := repo.HasAlertToday(targetCode, targetType, alertType)
	if err != nil {
		logger.Log.Error("check alert today failed", zap.Error(err))
		return nil
	}
	if hasAlert {
		return nil
	}

	return &model.AlertRecord{
		TargetCode:     targetCode,
		TargetType:     targetType,
		AlertType:      alertType,
		TriggerValue:   triggerValue,
		ThresholdValue: cumulativeProfitLine,
	}
}
