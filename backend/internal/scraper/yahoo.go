package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/pkg/money"
	"golang.org/x/time/rate"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
}

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

type YahooClient struct {
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	crumb       string
	cookies     []*http.Cookie
	crumbMu     sync.RWMutex
	cache       map[string]cacheEntry
	cacheMu     sync.RWMutex
}

func NewYahooClient() *YahooClient {
	return &YahooClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: rate.NewLimiter(rate.Every(500*time.Millisecond), 3),
		cache:       make(map[string]cacheEntry),
	}
}

func (c *YahooClient) doRequest(ctx context.Context, reqURL string) ([]byte, error) {
	// Check cache first (30s TTL — avoids duplicate calls for same data)
	c.cacheMu.RLock()
	if entry, ok := c.cache[reqURL]; ok && time.Now().Before(entry.expiresAt) {
		c.cacheMu.RUnlock()
		return entry.data, nil
	}
	c.cacheMu.RUnlock()

	data, err := c.doRequestWithRetry(ctx, reqURL, 0)
	if err != nil {
		return nil, err
	}

	c.cacheMu.Lock()
	c.cache[reqURL] = cacheEntry{data: data, expiresAt: time.Now().Add(30 * time.Second)}
	c.cacheMu.Unlock()

	return data, nil
}

func (c *YahooClient) doRequestWithRetry(ctx context.Context, reqURL string, attempt int) ([]byte, error) {
	if attempt > 3 {
		return nil, fmt.Errorf("yahoo request failed after %d retries", attempt)
	}

	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	// Ensure we have a crumb before first request
	c.crumbMu.RLock()
	hasCrumb := c.crumb != ""
	c.crumbMu.RUnlock()
	if !hasCrumb {
		if err := c.refreshCrumb(ctx); err != nil {
			return nil, fmt.Errorf("initial crumb fetch: %w", err)
		}
	}

	// Append crumb to URL if not already present
	c.crumbMu.RLock()
	finalURL := reqURL
	if c.crumb != "" && !strings.Contains(reqURL, "crumb=") {
		if strings.Contains(reqURL, "?") {
			finalURL += "&crumb=" + url.QueryEscape(c.crumb)
		} else {
			finalURL += "?crumb=" + url.QueryEscape(c.crumb)
		}
	}
	c.crumbMu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", finalURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
	req.Header.Set("Accept", "application/json")

	c.crumbMu.RLock()
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}
	c.crumbMu.RUnlock()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		log.Printf("[yahoo] %d on attempt %d, refreshing crumb: %s", resp.StatusCode, attempt, reqURL)
		if err := c.refreshCrumb(ctx); err != nil {
			return nil, fmt.Errorf("refresh crumb after %d: %w", resp.StatusCode, err)
		}
		return c.doRequestWithRetry(ctx, reqURL, attempt+1)
	}

	if resp.StatusCode == 429 {
		backoff := time.Duration(3*(attempt+1)) * time.Second
		log.Printf("[yahoo] 429 rate limited on attempt %d, backing off %v: %s", attempt, backoff, reqURL)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		return c.doRequestWithRetry(ctx, reqURL, attempt+1)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("yahoo returned %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *YahooClient) refreshCrumb(ctx context.Context) error {
	c.crumbMu.Lock()
	defer c.crumbMu.Unlock()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://fc.yahoo.com", nil)
	req.Header.Set("User-Agent", userAgents[0])
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch cookies: %w", err)
	}
	resp.Body.Close()
	c.cookies = resp.Cookies()

	crumbReq, _ := http.NewRequestWithContext(ctx, "GET", "https://query2.finance.yahoo.com/v1/test/getcrumb", nil)
	crumbReq.Header.Set("User-Agent", userAgents[0])
	for _, cookie := range c.cookies {
		crumbReq.AddCookie(cookie)
	}
	crumbResp, err := c.httpClient.Do(crumbReq)
	if err != nil {
		return fmt.Errorf("fetch crumb: %w", err)
	}
	defer crumbResp.Body.Close()
	crumbBytes, _ := io.ReadAll(crumbResp.Body)
	c.crumb = string(crumbBytes)
	log.Printf("[yahoo] refreshed crumb: %q (cookies: %d)", c.crumb, len(c.cookies))

	return nil
}

func (c *YahooClient) GetQuote(ctx context.Context, ticker string) (*domain.Quote, error) {
	quotes, err := c.GetBatchQuotes(ctx, []string{ticker})
	if err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quote found for %s", ticker)
	}
	return &quotes[0], nil
}

func (c *YahooClient) GetBatchQuotes(ctx context.Context, tickers []string) ([]domain.Quote, error) {
	symbols := strings.Join(tickers, ",")
	reqURL := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s", url.QueryEscape(symbols))

	body, err := c.doRequest(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch quotes: %w", err)
	}

	var resp yahooQuoteResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse quote response: %w", err)
	}

	quotes := make([]domain.Quote, 0, len(resp.QuoteResponse.Result))
	for _, r := range resp.QuoteResponse.Result {
		name := r.ShortName
		if r.LongName != "" {
			name = r.LongName
		}
		quotes = append(quotes, domain.Quote{
			Ticker:        r.Symbol,
			Name:          name,
			PriceCents:    money.FloatToCents(r.RegularMarketPrice),
			OpenCents:     money.FloatToCents(r.RegularMarketOpen),
			HighCents:     money.FloatToCents(r.RegularMarketDayHigh),
			LowCents:      money.FloatToCents(r.RegularMarketDayLow),
			Volume:        r.RegularMarketVolume,
			ChangePercent: r.RegularMarketChangePercent / 100.0,
			Currency:      r.Currency,
			Exchange:      r.FullExchangeName,
			Sector:        r.Sector,
			Industry:      r.Industry,
			FetchedAt:     time.Now(),
		})
	}

	return quotes, nil
}

func (c *YahooClient) SearchStock(ctx context.Context, query string) ([]domain.StockSearchResult, error) {
	reqURL := fmt.Sprintf("https://query2.finance.yahoo.com/v1/finance/search?q=%s&quotesCount=10&newsCount=0", url.QueryEscape(query))

	body, err := c.doRequest(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("search stocks: %w", err)
	}

	var resp yahooSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	results := make([]domain.StockSearchResult, 0, len(resp.Quotes))
	for _, q := range resp.Quotes {
		name := q.ShortName
		if q.LongName != "" {
			name = q.LongName
		}
		results = append(results, domain.StockSearchResult{
			Ticker:   q.Symbol,
			Name:     name,
			Exchange: q.Exchange,
			Type:     q.QuoteType,
		})
	}

	return results, nil
}

func (c *YahooClient) GetHistory(ctx context.Context, ticker string, rangeStr string, interval string) ([]domain.PricePoint, error) {
	if interval == "" {
		interval = "1d"
	}
	if rangeStr == "" {
		rangeStr = "1y"
	}

	reqURL := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=%s&interval=%s",
		url.PathEscape(ticker), url.QueryEscape(rangeStr), url.QueryEscape(interval))

	body, err := c.doRequest(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch history: %w", err)
	}

	var resp yahooChartResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse chart response: %w", err)
	}

	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart data for %s", ticker)
	}

	result := resp.Chart.Result[0]
	quotes := result.Indicators.Quote
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quote indicators for %s", ticker)
	}

	q := quotes[0]
	points := make([]domain.PricePoint, 0, len(result.Timestamps))

	for i, ts := range result.Timestamps {
		if i >= len(q.Open) || i >= len(q.Close) {
			break
		}

		adjClose := q.Close[i]
		if len(result.Indicators.AdjClose) > 0 && i < len(result.Indicators.AdjClose[0].AdjClose) {
			adjClose = result.Indicators.AdjClose[0].AdjClose[i]
		}

		var vol int64
		if i < len(q.Volume) {
			vol = q.Volume[i]
		}

		t := time.Unix(ts, 0)
		points = append(points, domain.PricePoint{
			Date:          t.Format("2006-01-02"),
			OpenCents:     money.FloatToCents(q.Open[i]),
			HighCents:     money.FloatToCents(q.High[i]),
			LowCents:      money.FloatToCents(q.Low[i]),
			CloseCents:    money.FloatToCents(q.Close[i]),
			AdjCloseCents: money.FloatToCents(adjClose),
			Volume:        vol,
		})
	}

	return points, nil
}

func (c *YahooClient) GetFundamentals(ctx context.Context, ticker string) (*domain.Fundamentals, error) {
	modules := "financialData,defaultKeyStatistics,incomeStatementHistory,cashflowStatementHistory,balanceSheetHistory"
	reqURL := fmt.Sprintf(
		"https://query2.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=%s",
		url.PathEscape(ticker), url.QueryEscape(modules),
	)

	body, err := c.doRequest(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch fundamentals for %s: %w", ticker, err)
	}

	var resp yahooQuoteSummaryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse quoteSummary for %s: %w", ticker, err)
	}

	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no fundamentals data for %s", ticker)
	}

	r := resp.QuoteSummary.Result[0]
	fd := r.FinancialData
	ks := r.DefaultKeyStatistics

	// Net income: try financialData first, fall back to incomeStatementHistory
	netIncome := fd.NetIncome.Raw
	if netIncome == 0 && len(r.IncomeStatementHistory.IncomeStatementHistory) > 0 {
		netIncome = r.IncomeStatementHistory.IncomeStatementHistory[0].NetIncome.Raw
	}

	// Free cash flow: try financialData first, fall back to cashflowStatementHistory
	fcf := fd.FreeCashflow.Raw
	if fcf == 0 && len(r.CashflowStatementHistory.CashflowStatements) > 0 {
		cf := r.CashflowStatementHistory.CashflowStatements[0]
		fcf = cf.TotalCashFromOperatingActivities.Raw + cf.CapitalExpenditures.Raw // capex is negative
	}

	// Extract historical net income (most recent first) for earnings growth
	var netIncomeHistory []int64
	for _, stmt := range r.IncomeStatementHistory.IncomeStatementHistory {
		netIncomeHistory = append(netIncomeHistory, money.FloatToCents(stmt.NetIncome.Raw))
	}

	// Extract historical revenue (most recent first)
	var revenueHistory []int64
	for _, stmt := range r.IncomeStatementHistory.IncomeStatementHistory {
		revenueHistory = append(revenueHistory, money.FloatToCents(stmt.TotalRevenue.Raw))
	}

	// Extract historical FCF (most recent first)
	var fcfHistory []int64
	for _, stmt := range r.CashflowStatementHistory.CashflowStatements {
		fcfVal := stmt.FreeCashFlow.Raw
		if fcfVal == 0 {
			fcfVal = stmt.TotalCashFromOperatingActivities.Raw + stmt.CapitalExpenditures.Raw
		}
		fcfHistory = append(fcfHistory, money.FloatToCents(fcfVal))
	}

	// Extract historical long-term debt (most recent first)
	var debtHistory []int64
	for _, stmt := range r.BalanceSheetHistory.BalanceSheetStatements {
		debtHistory = append(debtHistory, money.FloatToCents(stmt.LongTermDebt.Raw))
	}

	// Extract historical operating margin (most recent first)
	var marginHistory []float64
	for _, stmt := range r.IncomeStatementHistory.IncomeStatementHistory {
		if stmt.TotalRevenue.Raw > 0 {
			margin := stmt.OperatingIncome.Raw / stmt.TotalRevenue.Raw
			marginHistory = append(marginHistory, margin)
		}
	}

	return &domain.Fundamentals{
		FiscalYear:        time.Now().Year(),
		RevenueCents:      money.FloatToCents(fd.TotalRevenue.Raw),
		NetIncomeCents:    money.FloatToCents(netIncome),
		EPSCents:          money.FloatToCents(ks.TrailingEps.Raw),
		BookValueCents:    money.FloatToCents(ks.BookValue.Raw),
		FreeCashFlowCents: money.FloatToCents(fcf),
		SharesOutstanding: int64(ks.SharesOutstanding.Raw),
		DebtCents:         money.FloatToCents(fd.TotalDebt.Raw),
		CashCents:         money.FloatToCents(fd.TotalCash.Raw),
		ROE:               fd.ReturnOnEquity.Raw,
		ProfitMargin:      fd.ProfitMargins.Raw,
		NetIncomeHistory:  netIncomeHistory,
		RevenueHistory:    revenueHistory,
		FCFHistory:        fcfHistory,
		DebtHistory:       debtHistory,
		MarginHistory:     marginHistory,
	}, nil
}

func (c *YahooClient) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	ticker := fmt.Sprintf("%s%s=X", from, to)
	quote, err := c.GetQuote(ctx, ticker)
	if err != nil {
		return 0, fmt.Errorf("fetch exchange rate %s/%s: %w", from, to, err)
	}
	return money.CentsToFloat(quote.PriceCents), nil
}
