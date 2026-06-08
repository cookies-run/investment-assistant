package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type TargetType string

const (
	TargetTypeStock TargetType = "stock"
	TargetTypeFund  TargetType = "fund"
)

type AlertType string

const (
	AlertTypeSingleDayProfit  AlertType = "single_day_profit"
	AlertTypeSingleDayLoss    AlertType = "single_day_loss"
	AlertTypeCumulativeProfit AlertType = "cumulative_profit"
	AlertTypeCumulativeLoss   AlertType = "cumulative_loss"
)

type NotifyStatus string

const (
	NotifyStatusPending NotifyStatus = "pending"
	NotifyStatusSent    NotifyStatus = "sent"
	NotifyStatusFailed  NotifyStatus = "failed"
)

type Stock struct {
	StockCode            string          `gorm:"primaryKey;size:10;index" json:"stock_code"`
	UserID               uint            `gorm:"not null;default:0;index:idx_stock_user" json:"user_id"`
	StockName            string          `gorm:"not null;size:50" json:"stock_name"`
	BuyPrice             decimal.Decimal `gorm:"not null;type:decimal(10,3);default:0" json:"buy_price"`
	HoldQuantity         decimal.Decimal `gorm:"not null;type:decimal(15,4);default:0" json:"hold_quantity"`
	DailyProfitLine      decimal.Decimal `gorm:"not null;type:decimal(5,2);default:3.0" json:"daily_profit_line"`
	DailyLossLine        decimal.Decimal `gorm:"not null;type:decimal(5,2);default:3.0" json:"daily_loss_line"`
	CumulativeProfitLine decimal.Decimal `gorm:"not null;type:decimal(5,2);default:8.0" json:"cumulative_profit_line"`
	CumulativeLossLine   decimal.Decimal `gorm:"not null;type:decimal(5,2);default:8.0" json:"cumulative_loss_line"`
	CumulativeDays       int             `gorm:"not null;default:5" json:"cumulative_days"`
	MonitorInterval      int             `gorm:"not null;default:60" json:"monitor_interval"`
	IsActive             bool            `gorm:"not null;default:true" json:"is_active"`
	CreatedAt            time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// 设计说明：Fund 模型承载 14:50 量化决策引擎所需的全部策略配置。
// LongTermProfitLine / LongTermLossLine 构成战略级风控红线，用于第 2 关长期止盈/熔断判定。
// CapitalScalePreset 为资金规模预设（SMALL/MEDIUM/LARGE），前置决定阈值联动与均线过滤开关。
type Fund struct {
	FundCode             string          `gorm:"primaryKey;size:10;index" json:"fund_code"`
	UserID               uint            `gorm:"not null;default:0;index:idx_fund_user" json:"user_id"`
	FundName             string          `gorm:"not null;size:100" json:"fund_name"`
	HoldQuantity         decimal.Decimal `gorm:"not null;type:decimal(15,4);default:0" json:"hold_quantity"`
	HoldCost             decimal.Decimal `gorm:"not null;type:decimal(10,3);default:0" json:"hold_cost"`
	DailyProfitLine      decimal.Decimal `gorm:"not null;type:decimal(5,2);default:2.0" json:"daily_profit_line"`
	DailyLossLine        decimal.Decimal `gorm:"not null;type:decimal(5,2);default:2.0" json:"daily_loss_line"`
	CumulativeProfitLine decimal.Decimal `gorm:"not null;type:decimal(5,2);default:8.0" json:"cumulative_profit_line"`
	CumulativeLossLine   decimal.Decimal `gorm:"not null;type:decimal(5,2);default:8.0" json:"cumulative_loss_line"`
	CumulativeDays       int             `gorm:"not null;default:5" json:"cumulative_days"`
	LongTermProfitLine   decimal.Decimal `gorm:"not null;type:decimal(5,2);default:20.0" json:"long_term_profit_line"`
	LongTermLossLine     decimal.Decimal `gorm:"not null;type:decimal(5,2);default:15.0" json:"long_term_loss_line"`
	CapitalScalePreset   string          `gorm:"not null;size:10;default:MEDIUM" json:"capital_scale_preset"`
	FundType             string          `gorm:"not null;size:10;default:ACTIVE" json:"fund_type"`
	RelatedIndexSymbol   string          `gorm:"size:30" json:"related_index_symbol"`
	BaseCurrency         string          `gorm:"not null;size:10;default:CNY" json:"base_currency"`
	IsActive             bool            `gorm:"not null;default:true" json:"is_active"`
	LastRebalancedAt     *time.Time      `gorm:"type:date" json:"last_rebalanced_at"`
	CreatedAt            time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

type FundHolding struct {
	ID         uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	FundCode   string          `gorm:"not null;size:10;index" json:"fund_code"`
	StockCode  string          `gorm:"not null;size:10" json:"stock_code"`
	StockName  string          `gorm:"not null;size:50" json:"stock_name"`
	HoldRatio  decimal.Decimal `gorm:"not null;type:decimal(5,2)" json:"hold_ratio"`
	ReportDate time.Time       `gorm:"not null;type:date" json:"report_date"`
	SyncedAt   time.Time       `gorm:"autoCreateTime" json:"synced_at"`
}

type FundLot struct {
	ID          uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	FundCode    string          `gorm:"not null;size:10;index" json:"fund_code"`
	Quantity    decimal.Decimal `gorm:"not null;type:decimal(15,4)" json:"quantity"`
	Cost        decimal.Decimal `gorm:"not null;type:decimal(10,3);default:0" json:"cost"`
	PurchasedAt time.Time       `gorm:"not null;type:date" json:"purchased_at"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

type DailyRecord struct {
	ID                uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetCode        string          `gorm:"not null;size:10;index" json:"target_code"`
	TargetType        TargetType      `gorm:"not null;size:10" json:"target_type"`
	TradeDate         time.Time       `gorm:"not null;type:date" json:"trade_date"`
	OpenPrice         *decimal.Decimal `gorm:"type:decimal(10,3)" json:"open_price"`
	ClosePrice        *decimal.Decimal `gorm:"type:decimal(10,3)" json:"close_price"`
	ChangePercent     decimal.Decimal `gorm:"not null;type:decimal(6,3)" json:"change_percent"`
	CumulativePercent decimal.Decimal `gorm:"not null;type:decimal(6,3);default:0" json:"cumulative_percent"`
	CreatedAt         time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

type AlertRecord struct {
	ID             uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetCode     string           `gorm:"not null;size:10;index" json:"target_code"`
	TargetName     string           `gorm:"not null;size:100" json:"target_name"`
	TargetType     TargetType       `gorm:"not null;size:10" json:"target_type"`
	AlertType      AlertType        `gorm:"not null;size:30" json:"alert_type"`
	TriggerValue   decimal.Decimal  `gorm:"not null;type:decimal(6,3)" json:"trigger_value"`
	ThresholdValue decimal.Decimal  `gorm:"not null;type:decimal(5,2)" json:"threshold_value"`
	CurrentPrice   *decimal.Decimal `gorm:"type:decimal(10,3)" json:"current_price"`
	NotifyStatus   NotifyStatus     `gorm:"not null;size:10;default:pending" json:"notify_status"`
	TriggeredAt    time.Time        `gorm:"autoCreateTime" json:"triggered_at"`
}

type MarketIndexGroup struct {
	ID        uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint              `gorm:"not null;default:0;index:idx_mig_user" json:"user_id"`
	Name      string            `gorm:"not null;size:50" json:"name"`
	SortOrder int               `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	Items     []MarketIndexItem `gorm:"foreignKey:GroupID;references:ID;constraint:OnDelete:CASCADE;" json:"items"`
}

type MarketIndexItem struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"not null;default:0;index:idx_mii_user" json:"user_id"`
	GroupID    uint      `gorm:"not null;index" json:"group_id"`
	Symbol     string    `gorm:"not null;size:30" json:"symbol"`
	Name       string    `gorm:"not null;size:50" json:"name"`
	SourceType string    `gorm:"not null;size:20" json:"source_type"`
	SortOrder  int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type StockGroup struct {
	ID        uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint             `gorm:"not null;default:0;index:idx_sg_user" json:"user_id"`
	Name      string           `gorm:"not null;size:50" json:"name"`
	SortOrder int              `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	Items     []StockGroupItem `gorm:"foreignKey:GroupID;references:ID;constraint:OnDelete:CASCADE;" json:"items"`
}

type StockGroupItem struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;default:0;index:idx_sgi_user" json:"user_id"`
	GroupID   uint      `gorm:"not null;index" json:"group_id"`
	StockCode string    `gorm:"not null;size:10" json:"stock_code"`
	StockName string    `gorm:"not null;size:50" json:"stock_name"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type FundGroup struct {
	ID        uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint            `gorm:"not null;default:0;index:idx_fg_user" json:"user_id"`
	Name      string          `gorm:"not null;size:50" json:"name"`
	SortOrder int             `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	Items     []FundGroupItem `gorm:"foreignKey:GroupID;references:ID;constraint:OnDelete:CASCADE;" json:"items"`
}

type FundGroupItem struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;default:0;index:idx_fgi_user" json:"user_id"`
	GroupID   uint      `gorm:"not null;index" json:"group_id"`
	FundCode  string    `gorm:"not null;size:10" json:"fund_code"`
	FundName  string    `gorm:"not null;size:100" json:"fund_name"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type NotificationConfig struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint      `gorm:"not null;default:0;index:idx_nc_user" json:"user_id"`
	FeishuWebhook string    `gorm:"size:500" json:"feishu_webhook"`
	EnableFeishu  bool      `gorm:"not null;default:false" json:"enable_feishu"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type DailyClose struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"not null;default:0;index:idx_dc_user" json:"user_id"`
	Symbol     string    `gorm:"not null;size:30;index:idx_symbol_date,unique" json:"symbol"`
	Date       string    `gorm:"not null;size:10;index:idx_symbol_date,unique" json:"date"`
	ClosePrice float64   `gorm:"not null" json:"close_price"`
	SourceType string    `gorm:"not null;size:20" json:"source_type"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UnionID   string    `gorm:"size:64;uniqueIndex" json:"union_id,omitempty"`
	OpenID    string    `gorm:"size:64;uniqueIndex" json:"open_id,omitempty"`
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
