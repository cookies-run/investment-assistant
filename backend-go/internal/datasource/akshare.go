package datasource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"stock-monitor/pkg/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type LevelInfo struct {
	Price  decimal.Decimal `json:"price"`
	Volume int64           `json:"volume"`
}

type StockSpot struct {
	Code                  string          `json:"code"`
	Name                  string          `json:"name"`
	CurrentPrice          decimal.Decimal `json:"current_price"`
	PrevClose             decimal.Decimal `json:"prev_close"`
	OpenPrice             decimal.Decimal `json:"open_price"`
	High                  decimal.Decimal `json:"high"`
	Low                   decimal.Decimal `json:"low"`
	ChangePercent         decimal.Decimal `json:"change_percent"`
	Volume                int64           `json:"volume"`
	Turnover              decimal.Decimal `json:"turnover"`
	BuyVolume             int64           `json:"buy_volume"`
	SellVolume            int64           `json:"sell_volume"`
	BuySellDiff           int64           `json:"buy_sell_diff"`
	BuyLevels             []LevelInfo     `json:"buy_levels"`
	SellLevels            []LevelInfo     `json:"sell_levels"`
	Trends                []TrendPoint    `json:"trends"`
	ActiveBuyVol          int64           `json:"active_buy_vol"`            // 外盘（主动买入）
	ActiveSellVol         int64           `json:"active_sell_vol"`           // 内盘（主动卖出）
	LargeOrderNetInflow   decimal.Decimal `json:"large_order_net_inflow"`    // 大单净流入（需L2，暂预留）
	MinuteTurnoverRate    decimal.Decimal `json:"minute_turnover_rate"`      // 当前分钟换手率%
	MinutePriceVolatility decimal.Decimal `json:"minute_price_volatility"`   // 当前分钟震幅%
}

type TrendPoint struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume float64 `json:"volume"`
}

type MinutePoint struct {
	Time     string  `json:"time"`
	Price    float64 `json:"price"`
	Volume   int64   `json:"volume"`
	Turnover float64 `json:"turnover"`
}

type FundEstimation struct {
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	EstNAV    decimal.Decimal `json:"est_nav"`
	EstChange decimal.Decimal `json:"est_change"`
}

type FundHolding struct {
	FundCode   string          `json:"fund_code"`
	StockCode  string          `json:"stock_code"`
	StockName  string          `json:"stock_name"`
	HoldRatio  decimal.Decimal `json:"hold_ratio"`
	ReportDate string          `json:"report_date"`
}

type SearchResult struct {
	Code          string           `json:"code"`
	Name          string           `json:"name"`
	CurrentPrice  *decimal.Decimal `json:"current_price,omitempty"`
	ChangePercent *decimal.Decimal `json:"change_percent,omitempty"`
}

func prefixCode(code string) string {
	code = strings.TrimSpace(code)
	if len(code) == 8 && (code[:2] == "sh" || code[:2] == "sz" || code[:2] == "bj") {
		return code
	}
	if strings.HasPrefix(code, "6") {
		return "sh" + code
	}
	if strings.HasPrefix(code, "8") || strings.HasPrefix(code, "4") {
		return "bj" + code
	}
	return "sz" + code
}

func parseInt64(s string) int64 {
	var v int64
	fmt.Sscanf(s, "%d", &v)
	return v
}

func pureCode(symbol string) string {
	if len(symbol) == 8 && (symbol[:2] == "sh" || symbol[:2] == "sz" || symbol[:2] == "bj") {
		return symbol[2:]
	}
	return symbol
}

func FetchStockSpot(codes []string) ([]StockSpot, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	prefixed := make([]string, 0, len(codes))
	for _, c := range codes {
		prefixed = append(prefixed, prefixCode(c))
	}

	raw, err := fetchSinaRaw(prefixed)
	if err != nil {
		logger.Log.Warn("fetch stock spot failed", zap.Error(err))
		return nil, err
	}

	var spots []StockSpot
	for _, code := range codes {
		symbol := prefixCode(code)
		val, ok := raw[symbol]
		if !ok || val == "" {
			continue
		}
		parts := strings.Split(val, ",")
		if len(parts) < 32 {
			continue
		}
		name := parts[0]
		openPrice, _ := decimal.NewFromString(parts[1])
		prevClose, _ := decimal.NewFromString(parts[2])
		current, _ := decimal.NewFromString(parts[3])
		high, _ := decimal.NewFromString(parts[4])
		low, _ := decimal.NewFromString(parts[5])
		changePct := decimal.Zero
		if prevClose.IsPositive() {
			changePct = current.Sub(prevClose).Div(prevClose).Mul(decimal.NewFromInt(100)).Round(2)
		}

		volume := parseInt64(parts[8])
		turnover, _ := decimal.NewFromString(parts[9])

		buyVol := parseInt64(parts[10]) + parseInt64(parts[12]) + parseInt64(parts[14]) + parseInt64(parts[16]) + parseInt64(parts[18])
		sellVol := parseInt64(parts[20]) + parseInt64(parts[22]) + parseInt64(parts[24]) + parseInt64(parts[26]) + parseInt64(parts[28])

		var buyLevels []LevelInfo
		var sellLevels []LevelInfo
		for i := 0; i < 5; i++ {
			bp, _ := decimal.NewFromString(parts[11+i*2])
			bv := parseInt64(parts[10+i*2])
			buyLevels = append(buyLevels, LevelInfo{Price: bp, Volume: bv})
			sp, _ := decimal.NewFromString(parts[21+i*2])
			sv := parseInt64(parts[20+i*2])
			sellLevels = append(sellLevels, LevelInfo{Price: sp, Volume: sv})
		}

		spots = append(spots, StockSpot{
			Code:          code,
			Name:          name,
			CurrentPrice:  current,
			PrevClose:     prevClose,
			OpenPrice:     openPrice,
			High:          high,
			Low:           low,
			ChangePercent: changePct,
			Volume:        volume,
			Turnover:      turnover,
			BuyVolume:     buyVol,
			SellVolume:    sellVol,
			BuySellDiff:   buyVol - sellVol,
			BuyLevels:     buyLevels,
			SellLevels:    sellLevels,
		})
	}
	return spots, nil
}

func eastmoneySecid(code string) string {
	code = pureCode(code)
	if strings.HasPrefix(code, "6") {
		return "1." + code
	}
	if strings.HasPrefix(code, "8") || strings.HasPrefix(code, "4") {
		return "" // 北交所暂不支持东方财富该接口
	}
	return "0." + code
}

// fetchEastmoneyActiveVolumes 从东方财富获取外盘/内盘（主动买入/卖出）。
func fetchEastmoneyActiveVolumes(code string) (activeBuy, activeSell int64, floatShares int64, err error) {
	secid := eastmoneySecid(code)
	if secid == "" {
		return 0, 0, 0, nil
	}
	u := fmt.Sprintf("https://push2.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f49,f50,f84,f85", secid)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, 0, err
	}

	var result struct {
		Data struct {
			F43 float64 `json:"f43"`
			F49 float64 `json:"f49"`
			F50 float64 `json:"f50"`
			F84 float64 `json:"f84"`
			F85 float64 `json:"f85"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, 0, err
	}
	// f49/f50 单位为"手"，转为"股"
	activeBuy = int64(result.Data.F49 * 100)
	activeSell = int64(result.Data.F50 * 100)
	floatShares = int64(result.Data.F85) // 流通股本（股）

	// 数据合理性校验：东财接口偶发返回异常小的内盘，导致 AR 爆炸
	if activeSell < 1000 {
		return 0, 0, floatShares, fmt.Errorf("abnormal active sell volume: %d (too small)", activeSell)
	}
	if activeSell > 0 && activeBuy/activeSell > 100 {
		return 0, 0, floatShares, fmt.Errorf("abnormal ratio active_buy/active_sell: %d/%d", activeBuy, activeSell)
	}

	return activeBuy, activeSell, floatShares, nil
}

// FetchStockSpotEnhanced 为详情页获取增强盘口数据（含外盘/内盘、分钟换手/震幅）。
func FetchStockSpotEnhanced(code string) (*StockSpot, error) {
	spots, err := FetchStockSpot([]string{code})
	if err != nil || len(spots) == 0 {
		return nil, err
	}
	spot := &spots[0]

	// 补充外盘/内盘 + 流通股本
	activeBuy, activeSell, floatShares, err := fetchEastmoneyActiveVolumes(code)
	if err == nil {
		spot.ActiveBuyVol = activeBuy
		spot.ActiveSellVol = activeSell
	} else {
		logger.Log.Warn("fetch eastmoney active volumes failed", zap.String("code", code), zap.Error(err))
	}

	// 计算分钟换手率和分钟震幅
	minutePoints, err := FetchStockMinute(code)
	if err == nil && len(minutePoints) >= 2 {
		latest := minutePoints[len(minutePoints)-1]
		prev := minutePoints[len(minutePoints)-2]
		if floatShares > 0 {
			// 分钟换手率 = 该分钟成交量 / 流通股本 * 100
			spot.MinuteTurnoverRate = decimal.NewFromInt(latest.Volume).Div(decimal.NewFromInt(floatShares)).Mul(decimal.NewFromInt(100)).Round(4)
		}
		if prev.Price > 0 {
			// 分钟震幅 = |当前分钟价 - 上一分钟价| / 上一分钟价 * 100
			volatility := (latest.Price - prev.Price) / prev.Price * 100
			spot.MinutePriceVolatility = decimal.NewFromFloat(volatility).Abs().Round(4)
		}
	}

	return spot, nil
}

func FetchFundEstimation(fundCode string) (*FundEstimation, error) {
	url := fmt.Sprintf("https://fundgz.1234567.com.cn/js/%s.js", fundCode)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	text := string(body)
	// jsonpgz({...});
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("invalid response format")
	}

	var data struct {
		FundCode string `json:"fundcode"`
		Name     string `json:"name"`
		DWJZ     string `json:"dwjz"`
		GSZ      string `json:"gsz"`
		GSZZL    string `json:"gszzl"`
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &data); err != nil {
		return nil, err
	}

	estNav, _ := decimal.NewFromString(data.GSZ)
	estChange, _ := decimal.NewFromString(data.GSZZL)
	return &FundEstimation{
		Code:      data.FundCode,
		Name:      data.Name,
		EstNAV:    estNav,
		EstChange: estChange,
	}, nil
}

func FetchFundHoldings(fundCode string) ([]FundHolding, error) {
	url := fmt.Sprintf("https://fundf10.eastmoney.com/FundArchivesDatas.aspx?type=jjcc&code=%s&topline=10", fundCode)
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Referer", "https://fundf10.eastmoney.com/")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	text := string(body)
	// Extract report date
	var reportDate string
	if m := regexp.MustCompile(`截止至：<font[^>]*>(\d{4}-\d{2}-\d{2})</font>`).FindStringSubmatch(text); len(m) > 1 {
		reportDate = m[1]
	} else if m := regexp.MustCompile(`(\d{4}年\d季度)`).FindStringSubmatch(text); len(m) > 1 {
		reportDate = m[1]
	}

	// Extract table rows
	// Format: <tr><td>1</td><td><a href='...'>300308</a></td><td class='tol'><a href='...'>中际旭创</a></td>...<td class='tor'>4.31%</td>...</tr>
	var holdings []FundHolding
	rowRe := regexp.MustCompile(`<tr>\s*<td>\d+</td>\s*<td>\s*<a[^>]*>\s*(\d{6})\s*</a>\s*</td>\s*<td[^>]*>\s*<a[^>]*>\s*([^<]+)\s*</a>\s*</td>.*?<td[^>]*>\s*([\d.]+)%\s*</td>`)
	matches := rowRe.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		ratio, _ := decimal.NewFromString(m[3])
		holdings = append(holdings, FundHolding{
			FundCode:   fundCode,
			StockCode:  strings.TrimSpace(m[1]),
			StockName:  strings.TrimSpace(m[2]),
			HoldRatio:  ratio,
			ReportDate: reportDate,
		})
	}

	return holdings, nil
}

func tencentSymbol(code string) string {
	if len(code) == 8 && (code[:2] == "sh" || code[:2] == "sz" || code[:2] == "bj") {
		code = code[2:]
	}
	if strings.HasPrefix(code, "6") {
		return "sh" + code
	}
	if strings.HasPrefix(code, "8") || strings.HasPrefix(code, "4") {
		return "bj" + code
	}
	return "sz" + code
}

func FetchStockTrends(code string) ([]TrendPoint, error) {
	symbol := tencentSymbol(code)
	url := fmt.Sprintf("https://web.ifzq.gtimg.cn/appstock/app/fqkline/get?param=%s,day,,,30,qfq", symbol)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int `json:"code"`
		Data map[string]struct {
			Qfqday [][]string `json:"qfqday"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	stockData, ok := result.Data[symbol]
	if !ok || len(stockData.Qfqday) == 0 {
		return nil, nil
	}

	var trends []TrendPoint
	for _, item := range stockData.Qfqday {
		if len(item) < 6 {
			continue
		}
		open, _ := strconv.ParseFloat(item[1], 64)
		close, _ := strconv.ParseFloat(item[2], 64)
		high, _ := strconv.ParseFloat(item[3], 64)
		low, _ := strconv.ParseFloat(item[4], 64)
		vol, _ := strconv.ParseFloat(item[5], 64)
		trends = append(trends, TrendPoint{
			Date:   item[0],
			Open:   open,
			Close:  close,
			High:   high,
			Low:    low,
			Volume: vol,
		})
	}
	return trends, nil
}

func FetchStockMinute(code string) ([]MinutePoint, error) {
	symbol := tencentSymbol(code)
	url := fmt.Sprintf("https://web.ifzq.gtimg.cn/appstock/app/minute/query?code=%s", symbol)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int `json:"code"`
		Data map[string]struct {
			Inner struct {
				Data []string `json:"data"`
			} `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	stockData, ok := result.Data[symbol]
	if !ok || len(stockData.Inner.Data) == 0 {
		return nil, nil
	}

	var points []MinutePoint
	for _, item := range stockData.Inner.Data {
		parts := strings.Split(item, " ")
		if len(parts) < 2 {
			continue
		}
		price, _ := strconv.ParseFloat(parts[1], 64)
		vol := int64(0)
		turnover := float64(0)
		if len(parts) >= 3 {
			vol, _ = strconv.ParseInt(parts[2], 10, 64)
		}
		if len(parts) >= 4 {
			turnover, _ = strconv.ParseFloat(parts[3], 64)
		}
		points = append(points, MinutePoint{
			Time:     parts[0],
			Price:    price,
			Volume:   vol,
			Turnover: turnover,
		})
	}
	return points, nil
}

func SearchFunds(keyword string) ([]SearchResult, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, nil
	}
	url := fmt.Sprintf("https://searchapi.eastmoney.com/api/suggest/get?input=%s&type=14&count=20", url.QueryEscape(keyword))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		QuotationCodeTable struct {
			Data []struct {
				Code             string `json:"Code"`
				Name             string `json:"Name"`
				Classify         string `json:"Classify"`
				SecurityTypeName string `json:"SecurityTypeName"`
			} `json:"Data"`
		} `json:"QuotationCodeTable"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, item := range result.QuotationCodeTable.Data {
		if item.SecurityTypeName != "基金" {
			continue
		}
		results = append(results, SearchResult{
			Code: item.Code,
			Name: item.Name,
		})
		if len(results) >= 20 {
			break
		}
	}
	return results, nil
}

func SearchStocks(keyword string) ([]SearchResult, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, nil
	}
	url := fmt.Sprintf("https://searchapi.eastmoney.com/api/suggest/get?input=%s&type=14&count=20", url.QueryEscape(keyword))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		QuotationCodeTable struct {
			Data []struct {
				Code     string `json:"Code"`
				Name     string `json:"Name"`
				Classify string `json:"Classify"`
			} `json:"Data"`
		} `json:"QuotationCodeTable"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, item := range result.QuotationCodeTable.Data {
		if item.Classify != "AStock" {
			continue
		}
		results = append(results, SearchResult{
			Code: item.Code,
			Name: item.Name,
		})
		if len(results) >= 20 {
			break
		}
	}
	return results, nil
}
