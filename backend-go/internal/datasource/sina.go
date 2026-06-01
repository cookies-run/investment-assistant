package datasource

import (
	"io"
	"net/http"
	"stock-monitor/internal/model"
	"stock-monitor/pkg/logger"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type MarketItem struct {
	Symbol        string          `json:"symbol"`
	Name          string          `json:"name"`
	Price         decimal.Decimal `json:"price"`
	Change        decimal.Decimal `json:"change"`
	ChangePercent decimal.Decimal `json:"change_percent"`
	TradeDate     string          `json:"trade_date"`
	MA20          *float64        `json:"ma20,omitempty"`
}

type MarketCategory struct {
	Name  string       `json:"name"`
	Items []MarketItem `json:"items"`
}

type MarketDashboard struct {
	UpdateTime time.Time        `json:"update_time"`
	Categories []MarketCategory `json:"categories"`
}

func fetchSinaRaw(symbols []string) (map[string]string, error) {
	u := "https://hq.sinajs.cn/list=" + strings.Join(symbols, ",")
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Referer", "https://finance.sina.com.cn")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := string(body)
	results := make(map[string]string)
	for _, line := range strings.Split(data, ";") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, `="`) {
			continue
		}
		parts := strings.SplitN(line, `="`, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimPrefix(parts[0], "var hq_str_")
		val := strings.TrimSuffix(parts[1], `"`)
		results[key] = val
	}
	return results, nil
}

func parseAShareItem(symbol, name, val string) MarketItem {
	parts := strings.Split(val, ",")
	if len(parts) < 32 {
		return MarketItem{Name: name}
	}
	prev, _ := decimal.NewFromString(parts[2])
	current, _ := decimal.NewFromString(parts[3])
	change := current.Sub(prev)
	changePct := decimal.Zero
	if !prev.IsZero() {
		changePct = change.Div(prev).Mul(decimal.NewFromInt(100)).Round(2)
	}
	dt := parts[30]
	tm := parts[31]
	tradeDate := dt
	if tm != "" {
		tradeDate = dt + " " + tm
	}
	return MarketItem{
		Symbol:        symbol,
		Name:          name,
		Price:         current.Round(2),
		Change:        change.Round(2),
		ChangePercent: changePct,
		TradeDate:     tradeDate,
	}
}

func parseHKItem(symbol, name, val string) MarketItem {
	parts := strings.Split(val, ",")
	if len(parts) < 19 {
		return MarketItem{Name: name}
	}
	current, _ := decimal.NewFromString(parts[6])
	change, _ := decimal.NewFromString(parts[7])
	changePct, _ := decimal.NewFromString(parts[8])
	dt := strings.Replace(parts[17], "/", "-", -1)
	tm := parts[18]
	tradeDate := dt
	if tm != "" {
		tradeDate = dt + " " + tm
	}
	return MarketItem{
		Symbol:        symbol,
		Name:          name,
		Price:         current.Round(2),
		Change:        change.Round(2),
		ChangePercent: changePct.Round(2),
		TradeDate:     tradeDate,
	}
}

func parseUSItem(symbol, name, val string) MarketItem {
	parts := strings.Split(val, ",")
	if len(parts) < 5 {
		return MarketItem{Name: name}
	}
	current, _ := decimal.NewFromString(parts[1])
	changePct, _ := decimal.NewFromString(parts[2])
	change, _ := decimal.NewFromString(parts[4])
	return MarketItem{
		Symbol:        symbol,
		Name:          name,
		Price:         current.Round(2),
		Change:        change.Round(2),
		ChangePercent: changePct.Round(2),
		TradeDate:     time.Now().Format("2006-01-02 15:04:05"),
	}
}

func parseFuturesItem(symbol, name, val string) MarketItem {
	parts := strings.Split(val, ",")
	if len(parts) < 14 {
		return MarketItem{Name: name}
	}
	price, _ := decimal.NewFromString(parts[0])
	prev, _ := decimal.NewFromString(parts[7])
	change := price.Sub(prev)
	changePct := decimal.Zero
	if !prev.IsZero() {
		changePct = change.Div(prev).Mul(decimal.NewFromInt(100)).Round(2)
	}
	dt := strings.Replace(parts[12], "/", "-", -1)
	tm := parts[6]
	tradeDate := dt
	if tm != "" {
		tradeDate = dt + " " + tm
	}
	return MarketItem{
		Symbol:        symbol,
		Name:          name,
		Price:         price.Round(2),
		Change:        change.Round(2),
		ChangePercent: changePct,
		TradeDate:     tradeDate,
	}
}

func parseByType(symbol, name, sourceType, val string) MarketItem {
	switch sourceType {
	case "a_share":
		return parseAShareItem(symbol, name, val)
	case "hk":
		return parseHKItem(symbol, name, val)
	case "us":
		return parseUSItem(symbol, name, val)
	case "futures":
		return parseFuturesItem(symbol, name, val)
	default:
		return MarketItem{Name: name}
	}
}

// FetchIndexRealtime 拉取单个指数实时行情（复用 Sina 行情通道）。
// symbol 需为 Sina 支持的格式，如 sh000300、rt_hkHSTECH、gb_ixic。
func FetchIndexRealtime(symbol string) (MarketItem, error) {
	data, err := fetchSinaRaw([]string{symbol})
	if err != nil {
		return MarketItem{}, err
	}
	val, ok := data[symbol]
	if !ok || val == "" {
		return MarketItem{}, nil
	}
	// 根据 symbol 前缀推断 source_type
	sourceType := "a_share"
	if strings.HasPrefix(symbol, "rt_hk") {
		sourceType = "hk"
	} else if strings.HasPrefix(symbol, "gb_") {
		sourceType = "us"
	} else if strings.HasPrefix(symbol, "hf_") {
		sourceType = "futures"
	}
	return parseByType(symbol, "", sourceType, val), nil
}

// fxSymbolMap 将底层计价货币映射为 Sina 外汇行情符号。
// 注意：Sina 返回的是 base_currency/CNY 的涨跌幅，
// 函数内部会取负值以转换为 CNY/base_currency 的涨跌幅。
var fxSymbolMap = map[string]string{
	"USD": "fx_susdcnh",
	"HKD": "fx_shkdcny",
}

// FetchFXChange 获取人民币对目标货币的日内涨跌幅。
// 返回值 > 0 表示人民币升值（对跨境 QDII 净值产生压制）。
func FetchFXChange(currency string) (decimal.Decimal, error) {
	if currency == "" || currency == "CNY" {
		return decimal.Zero, nil
	}
	symbol, ok := fxSymbolMap[currency]
	if !ok {
		logger.Log.Warn("unsupported base currency for FX fetch", zap.String("currency", currency))
		return decimal.Zero, nil
	}
	data, err := fetchSinaRaw([]string{symbol})
	if err != nil {
		return decimal.Zero, err
	}
	val, ok := data[symbol]
	if !ok || val == "" {
		return decimal.Zero, nil
	}
	parts := strings.Split(val, ",")
	// Sina 外汇行情格式：时间, 买入, 卖出, 最新, ..., 昨收, 名称, 涨跌幅, 涨跌额, ...
	// 实测第 10 个字段（0-indexed）为涨跌幅（%）。
	if len(parts) < 11 {
		logger.Log.Warn("unexpected FX data format", zap.String("symbol", symbol), zap.String("raw", val))
		return decimal.Zero, nil
	}
	changePct, err := decimal.NewFromString(parts[10])
	if err != nil {
		logger.Log.Warn("failed to parse FX change percent", zap.String("symbol", symbol), zap.String("value", parts[10]))
		return decimal.Zero, nil
	}
	// Sina 给出的是 base_currency/CNY 的涨跌幅，取负值转换为 CNY/base_currency 的涨跌幅。
	return changePct.Neg().Round(4), nil
}

func FetchMarketDashboard(groups []model.MarketIndexGroup) (*MarketDashboard, error) {
	now := time.Now()

	if len(groups) == 0 {
		return &MarketDashboard{
			UpdateTime: now,
			Categories: []MarketCategory{},
		}, nil
	}

	type meta struct {
		Name       string
		SourceType string
		GroupID    uint
	}

	var allSymbols []string
	symbolMap := make(map[string]meta)

	for _, g := range groups {
		for _, item := range g.Items {
			allSymbols = append(allSymbols, item.Symbol)
			symbolMap[item.Symbol] = meta{
				Name:       item.Name,
				SourceType: item.SourceType,
				GroupID:    g.ID,
			}
		}
	}

	if len(allSymbols) == 0 {
		return &MarketDashboard{
			UpdateTime: now,
			Categories: []MarketCategory{},
		}, nil
	}

	data, err := fetchSinaRaw(allSymbols)
	if err != nil {
		logger.Log.Error("fetch sina raw failed", zap.Error(err))
		return nil, err
	}

	groupItems := make(map[uint][]MarketItem)
	for symbol, m := range symbolMap {
		val, ok := data[symbol]
		if !ok || val == "" {
			continue
		}
		item := parseByType(symbol, m.Name, m.SourceType, val)
		groupItems[m.GroupID] = append(groupItems[m.GroupID], item)
	}

	var categories []MarketCategory
	for _, g := range groups {
		if items, ok := groupItems[g.ID]; ok && len(items) > 0 {
			categories = append(categories, MarketCategory{
				Name:  g.Name,
				Items: items,
			})
		}
	}

	return &MarketDashboard{
		UpdateTime: now,
		Categories: categories,
	}, nil
}
