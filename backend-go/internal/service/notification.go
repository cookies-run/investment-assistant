package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"stock-monitor/internal/model"
	"stock-monitor/internal/repository"
	"stock-monitor/pkg/logger"

	"go.uber.org/zap"
)

type NotificationService struct {
	repo *repository.NotificationConfigRepo
}

func NewNotificationService(repo *repository.NotificationConfigRepo) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) GetConfig() (*model.NotificationConfig, error) {
	return s.repo.Get()
}

func (s *NotificationService) SaveConfig(cfg *model.NotificationConfig) error {
	return s.repo.Save(cfg)
}

func (s *NotificationService) SendFeishuAlert(alert *model.AlertRecord, targetName string, currentPrice *float64) {
	cfg, err := s.repo.Get()
	if err != nil {
		logger.Log.Warn("get notification config failed", zap.Error(err))
		return
	}
	if !cfg.EnableFeishu || cfg.FeishuWebhook == "" {
		return
	}

	alertTypeText := map[model.AlertType]string{
		model.AlertTypeSingleDayProfit:  "单日止盈",
		model.AlertTypeSingleDayLoss:    "单日止损",
		model.AlertTypeCumulativeProfit: "累计止盈",
		model.AlertTypeCumulativeLoss:   "累计止损",
	}

	alertTypeName := alertTypeText[alert.AlertType]
	if alertTypeName == "" {
		alertTypeName = string(alert.AlertType)
	}

	targetTypeText := "股票"
	if alert.TargetType == model.TargetTypeFund {
		targetTypeText = "基金"
	}

	priceStr := "未知"
	if currentPrice != nil {
		priceStr = fmt.Sprintf("%.3f", *currentPrice)
	}

	color := "red"
	if alert.AlertType == model.AlertTypeSingleDayLoss || alert.AlertType == model.AlertTypeCumulativeLoss {
		color = "green"
	}

	card := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"config": map[string]interface{}{"wide_screen_mode": true},
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": "🔔 投资助手策略提醒",
				},
				"template": color,
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{"is_short": true, "text": map[string]interface{}{"tag": "lark_md", "content": fmt.Sprintf("**标的类型**\n%s", targetTypeText)}},
						{"is_short": true, "text": map[string]interface{}{"tag": "lark_md", "content": fmt.Sprintf("**代码**\n%s", alert.TargetCode)}},
					},
				},
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{"is_short": true, "text": map[string]interface{}{"tag": "lark_md", "content": fmt.Sprintf("**名称**\n%s", targetName)}},
						{"is_short": true, "text": map[string]interface{}{"tag": "lark_md", "content": fmt.Sprintf("**触发类型**\n%s", alertTypeName)}},
					},
				},
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{"is_short": true, "text": map[string]interface{}{"tag": "lark_md", "content": fmt.Sprintf("**触发值**\n%.2f%%", alert.TriggerValue.InexactFloat64())}},
						{"is_short": true, "text": map[string]interface{}{"tag": "lark_md", "content": fmt.Sprintf("**当前价**\n%s", priceStr)}},
					},
				},
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": fmt.Sprintf("阈值: %.2f%% | 触发时间: %s", alert.ThresholdValue.InexactFloat64(), time.Now().Format("15:04:05")),
					},
				},
			},
		},
	}

	body, _ := json.Marshal(card)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(cfg.FeishuWebhook, "application/json", bytes.NewReader(body))
	if err != nil {
		logger.Log.Warn("send feishu alert failed", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Warn("feishu webhook returned non-200", zap.Int("status", resp.StatusCode))
	}
}
