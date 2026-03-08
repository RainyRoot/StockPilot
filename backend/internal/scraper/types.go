package scraper

// Yahoo Finance v7 quote response
type yahooQuoteResponse struct {
	QuoteResponse struct {
		Result []yahooQuoteResult `json:"result"`
		Error  interface{}        `json:"error"`
	} `json:"quoteResponse"`
}

type yahooQuoteResult struct {
	Symbol                     string  `json:"symbol"`
	ShortName                  string  `json:"shortName"`
	LongName                   string  `json:"longName"`
	Exchange                   string  `json:"exchange"`
	FullExchangeName           string  `json:"fullExchangeName"`
	Currency                   string  `json:"currency"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketOpen          float64 `json:"regularMarketOpen"`
	RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
	RegularMarketVolume        int64   `json:"regularMarketVolume"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	MarketCap                  int64   `json:"marketCap"`
	Sector                     string  `json:"sector"`
	Industry                   string  `json:"industry"`
	QuoteType                  string  `json:"quoteType"`
}

// Yahoo Finance v1 search response
type yahooSearchResponse struct {
	Quotes []yahooSearchQuote `json:"quotes"`
}

type yahooSearchQuote struct {
	Symbol    string `json:"symbol"`
	ShortName string `json:"shortname"`
	LongName  string `json:"longname"`
	Exchange  string `json:"exchange"`
	QuoteType string `json:"quoteType"`
}

// Yahoo Finance v8 chart response
type yahooChartResponse struct {
	Chart struct {
		Result []yahooChartResult `json:"result"`
		Error  interface{}        `json:"error"`
	} `json:"chart"`
}

type yahooChartResult struct {
	Timestamps []int64 `json:"timestamp"`
	Indicators struct {
		Quote []struct {
			Open   []float64 `json:"open"`
			High   []float64 `json:"high"`
			Low    []float64 `json:"low"`
			Close  []float64 `json:"close"`
			Volume []int64   `json:"volume"`
		} `json:"quote"`
		AdjClose []struct {
			AdjClose []float64 `json:"adjclose"`
		} `json:"adjclose"`
	} `json:"indicators"`
}
