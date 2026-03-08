<script lang="ts">
  import { searchStocks, type StockSearchResult } from '$lib/api/stocks';
  import { getValuation, getFundamentals, type ValuationResult, type Fundamentals } from '$lib/api/valuation';
  import { formatCents, formatPercent } from '$lib/utils/format';

  let query = $state('');
  let searchResults = $state<StockSearchResult[]>([]);
  let valuation = $state<ValuationResult | null>(null);
  let fundamentals = $state<Fundamentals | null>(null);
  let loading = $state(false);
  let error = $state('');
  let searchTimeout: ReturnType<typeof setTimeout>;

  function handleSearch() {
    clearTimeout(searchTimeout);
    if (query.length < 1) { searchResults = []; return; }
    searchTimeout = setTimeout(async () => {
      try { searchResults = await searchStocks(query); } catch {}
    }, 300);
  }

  async function selectStock(ticker: string) {
    try {
      loading = true; error = '';
      searchResults = []; query = ticker;
      const [val, fund] = await Promise.all([
        getValuation(ticker),
        getFundamentals(ticker),
      ]);
      valuation = val;
      fundamentals = fund;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load valuation';
      valuation = null;
      fundamentals = null;
    } finally {
      loading = false;
    }
  }

  function verdictColor(verdict: string): string {
    switch (verdict) {
      case 'UNDERVALUED': return '#3fb950';
      case 'FAIRLY_VALUED_LOW': return '#58a6ff';
      case 'FAIR_VALUE': return '#d29922';
      case 'OVERVALUED': return '#f85149';
      default: return '#8b949e';
    }
  }

  function verdictLabel(verdict: string): string {
    switch (verdict) {
      case 'UNDERVALUED': return 'Undervalued';
      case 'FAIRLY_VALUED_LOW': return 'Slightly Undervalued';
      case 'FAIR_VALUE': return 'Fair Value';
      case 'OVERVALUED': return 'Overvalued';
      default: return verdict;
    }
  }

  function qualityColor(score: number): string {
    if (score >= 70) return '#3fb950';
    if (score >= 50) return '#d29922';
    return '#f85149';
  }

  function formatLargeCents(cents: number): string {
    const val = Math.abs(cents) / 100;
    const sign = cents < 0 ? '-' : '';
    if (val >= 1_000_000_000) return `${sign}$${(val / 1_000_000_000).toFixed(2)}B`;
    if (val >= 1_000_000) return `${sign}$${(val / 1_000_000).toFixed(2)}M`;
    if (val >= 1_000) return `${sign}$${(val / 1_000).toFixed(2)}K`;
    return formatCents(cents);
  }
</script>

<div class="valuation-page">
  <div class="header">
    <h2>Buffett Valuation</h2>
  </div>

  <!-- Search -->
  <div class="search-box">
    <input
      type="text"
      placeholder="Search stock to analyze (e.g., AAPL, MSFT)..."
      bind:value={query}
      oninput={handleSearch}
      onkeydown={(e) => { if (e.key === 'Enter' && query.length > 0) selectStock(query.toUpperCase()); }}
    />
    {#if searchResults.length > 0}
      <div class="search-results">
        {#each searchResults as result}
          <button class="result-item" onclick={() => selectStock(result.ticker)}>
            <span class="ticker">{result.ticker}</span>
            <span class="name">{result.name}</span>
            <span class="exchange">{result.exchange}</span>
          </button>
        {/each}
      </div>
    {/if}
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  {#if loading}
    <p class="loading-text">Analyzing fundamentals...</p>
  {:else if valuation}
    <!-- Overview Cards -->
    <div class="cards">
      <div class="card">
        <div class="card-label">Current Price</div>
        <div class="card-value">{formatCents(valuation.current_price_cents)}</div>
      </div>
      <div class="card">
        <div class="card-label">Intrinsic Value</div>
        <div class="card-value">{formatCents(valuation.intrinsic_value_cents)}</div>
      </div>
      <div class="card">
        <div class="card-label">Margin of Safety</div>
        <div class="card-value" style="color: {verdictColor(valuation.verdict)}">
          {formatPercent(valuation.margin_of_safety)}
        </div>
      </div>
      <div class="card">
        <div class="card-label">Verdict</div>
        <div class="verdict-badge" style="background: {verdictColor(valuation.verdict)}20; color: {verdictColor(valuation.verdict)}; border: 1px solid {verdictColor(valuation.verdict)}40">
          {verdictLabel(valuation.verdict)}
        </div>
      </div>
    </div>

    <!-- Quality Score -->
    <div class="section">
      <h3>Quality Score</h3>
      <div class="quality-bar-container">
        <div class="quality-bar" style="width: {valuation.quality_score}%; background: {qualityColor(valuation.quality_score)}"></div>
      </div>
      <div class="quality-label" style="color: {qualityColor(valuation.quality_score)}">
        {valuation.quality_score} / 100
      </div>
    </div>

    <!-- Valuation Methods -->
    <div class="section">
      <h3>Valuation Methods</h3>
      <table class="methods-table">
        <thead>
          <tr>
            <th>Method</th>
            <th class="num">Fair Value</th>
            <th class="num">Weight</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          {#each valuation.methods as method}
            <tr>
              <td class="method-name">{method.name}</td>
              <td class="num">{formatCents(method.value_cents)}</td>
              <td class="num">{(method.weight * 100).toFixed(0)}%</td>
              <td class="method-desc">{method.description ?? ''}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Fundamentals -->
    {#if fundamentals}
      <div class="section">
        <h3>Fundamentals</h3>
        <div class="fundamentals-grid">
          <div class="fund-item">
            <span class="fund-label">Revenue</span>
            <span class="fund-value">{formatLargeCents(fundamentals.revenue_cents)}</span>
          </div>
          <div class="fund-item">
            <span class="fund-label">Net Income</span>
            <span class="fund-value">{formatLargeCents(fundamentals.net_income_cents)}</span>
          </div>
          <div class="fund-item">
            <span class="fund-label">EPS</span>
            <span class="fund-value">{formatCents(fundamentals.eps_cents)}</span>
          </div>
          <div class="fund-item">
            <span class="fund-label">Book Value / Share</span>
            <span class="fund-value">{formatCents(fundamentals.book_value_cents)}</span>
          </div>
          <div class="fund-item">
            <span class="fund-label">Free Cash Flow</span>
            <span class="fund-value">{formatLargeCents(fundamentals.free_cash_flow_cents)}</span>
          </div>
          <div class="fund-item">
            <span class="fund-label">Total Debt</span>
            <span class="fund-value">{formatLargeCents(fundamentals.debt_cents)}</span>
          </div>
          <div class="fund-item">
            <span class="fund-label">Total Cash</span>
            <span class="fund-value">{formatLargeCents(fundamentals.cash_cents)}</span>
          </div>
          <div class="fund-item">
            <span class="fund-label">Shares Outstanding</span>
            <span class="fund-value">{fundamentals.shares_outstanding > 0 ? (fundamentals.shares_outstanding / 1_000_000_000).toFixed(2) + 'B' : '—'}</span>
          </div>
          <div class="fund-item">
            <span class="fund-label">ROE</span>
            <span class="fund-value" class:positive={fundamentals.roe > 0.15} class:warning={fundamentals.roe > 0 && fundamentals.roe <= 0.15}>
              {(fundamentals.roe * 100).toFixed(2)}%
            </span>
          </div>
          <div class="fund-item">
            <span class="fund-label">Profit Margin</span>
            <span class="fund-value" class:positive={fundamentals.profit_margin > 0.20} class:warning={fundamentals.profit_margin > 0 && fundamentals.profit_margin <= 0.20}>
              {(fundamentals.profit_margin * 100).toFixed(2)}%
            </span>
          </div>
        </div>
      </div>
    {/if}
  {:else}
    <p class="empty-text">Search for a stock to see its Buffett-style valuation analysis.</p>
  {/if}
</div>

<style>
  .valuation-page {
    max-width: 1000px;
  }

  .header {
    margin-bottom: 1rem;
  }

  .header h2 {
    margin: 0;
  }

  .search-box {
    position: relative;
    margin-bottom: 1.5rem;
  }

  .search-box input {
    width: 100%;
    padding: 0.7rem 1rem;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    color: #e1e4e8;
    font-size: 0.95rem;
    outline: none;
    box-sizing: border-box;
  }

  .search-box input:focus { border-color: #58a6ff; }

  .search-results {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 0 0 6px 6px;
    border-top: none;
    max-height: 240px;
    overflow-y: auto;
    z-index: 10;
  }

  .result-item {
    display: flex;
    gap: 0.75rem;
    width: 100%;
    padding: 0.5rem 1rem;
    background: none;
    border: none;
    border-bottom: 1px solid #21262d;
    color: #e1e4e8;
    cursor: pointer;
    text-align: left;
    font-size: 0.85rem;
  }

  .result-item:hover { background: #1c2128; }
  .result-item:last-child { border-bottom: none; }
  .result-item .ticker { font-weight: 600; color: #58a6ff; min-width: 70px; }
  .result-item .name { color: #8b949e; flex: 1; }
  .result-item .exchange { color: #484f58; font-size: 0.8rem; }

  .error {
    background: #3d1f1f;
    border: 1px solid #f85149;
    color: #f85149;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    font-size: 0.9rem;
  }

  .cards {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .card {
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 8px;
    padding: 1.25rem;
  }

  .card-label {
    font-size: 0.8rem;
    color: #8b949e;
    text-transform: uppercase;
    margin-bottom: 0.5rem;
  }

  .card-value {
    font-size: 1.5rem;
    font-weight: 600;
    color: #e1e4e8;
  }

  .verdict-badge {
    display: inline-block;
    padding: 0.35rem 0.75rem;
    border-radius: 20px;
    font-size: 0.85rem;
    font-weight: 600;
    margin-top: 0.25rem;
  }

  .section {
    margin-bottom: 1.5rem;
  }

  .section h3 {
    margin: 0 0 0.75rem 0;
    color: #e1e4e8;
    font-size: 1.1rem;
  }

  .quality-bar-container {
    width: 100%;
    height: 12px;
    background: #21262d;
    border-radius: 6px;
    overflow: hidden;
    margin-bottom: 0.5rem;
  }

  .quality-bar {
    height: 100%;
    border-radius: 6px;
    transition: width 0.5s ease;
  }

  .quality-label {
    font-size: 0.9rem;
    font-weight: 600;
  }

  .methods-table {
    width: 100%;
    border-collapse: collapse;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    overflow: hidden;
  }

  .methods-table th {
    padding: 0.6rem 1rem;
    text-align: left;
    color: #8b949e;
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: uppercase;
    border-bottom: 1px solid #30363d;
  }

  .methods-table th.num, .methods-table td.num { text-align: right; }

  .methods-table td {
    padding: 0.6rem 1rem;
    border-bottom: 1px solid #21262d;
    font-size: 0.9rem;
  }

  .methods-table tr:last-child td { border-bottom: none; }
  .methods-table tr:hover { background: #1c2128; }

  .method-name { font-weight: 600; color: #58a6ff; }
  .method-desc { color: #8b949e; font-size: 0.8rem; }

  .fundamentals-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 0.75rem;
  }

  .fund-item {
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    padding: 0.75rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .fund-label {
    font-size: 0.75rem;
    color: #8b949e;
    text-transform: uppercase;
  }

  .fund-value {
    font-size: 1.05rem;
    font-weight: 600;
    color: #e1e4e8;
  }

  .positive { color: #3fb950; }
  .warning { color: #d29922; }

  .loading-text, .empty-text {
    color: #8b949e;
    text-align: center;
    padding: 3rem;
  }

  @media (max-width: 768px) {
    .cards { grid-template-columns: repeat(2, 1fr); }
  }
</style>
