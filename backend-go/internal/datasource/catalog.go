package datasource

type AvailableIndex struct {
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	SourceType string `json:"source_type"`
	Category   string `json:"category"`
}

func GetAvailableIndices() []AvailableIndex {
	return []AvailableIndex{
		// A股
		{Symbol: "sh000001", Name: "上证指数", SourceType: "a_share", Category: "A股"},
		{Symbol: "sz399001", Name: "深证成指", SourceType: "a_share", Category: "A股"},
		{Symbol: "sz399006", Name: "创业板指", SourceType: "a_share", Category: "A股"},
		{Symbol: "sh000688", Name: "科创50", SourceType: "a_share", Category: "A股"},
		{Symbol: "sh000016", Name: "上证50", SourceType: "a_share", Category: "A股"},
		{Symbol: "sh000905", Name: "中证500", SourceType: "a_share", Category: "A股"},
		{Symbol: "sh000300", Name: "沪深300", SourceType: "a_share", Category: "A股"},

		// 港股
		{Symbol: "rt_hkHSI", Name: "恒生指数", SourceType: "hk", Category: "港股"},
		{Symbol: "rt_hkHSTECH", Name: "恒生科技", SourceType: "hk", Category: "港股"},
		{Symbol: "rt_hkHSCEI", Name: "国企指数", SourceType: "hk", Category: "港股"},

		// 美股
		{Symbol: "gb_dji", Name: "道琼斯", SourceType: "us", Category: "美股"},
		{Symbol: "gb_ixic", Name: "纳斯达克", SourceType: "us", Category: "美股"},
		{Symbol: "gb_inx", Name: "标普500", SourceType: "us", Category: "美股"},

		// 日韩
		{Symbol: "hf_NK", Name: "日经225", SourceType: "futures", Category: "日韩"},

		// 商品
		{Symbol: "hf_GC", Name: "COMEX黄金", SourceType: "futures", Category: "商品"},
		{Symbol: "hf_CL", Name: "WTI原油", SourceType: "futures", Category: "商品"},
		{Symbol: "hf_SI", Name: "COMEX白银", SourceType: "futures", Category: "商品"},
	}
}
