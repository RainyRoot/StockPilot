<script lang="ts">
  import { searchStocks, getStockQuote, type StockSearchResult, type StockQuoteResponse } from '$lib/api/stocks';
  import { formatCents, formatPercent, formatVolume } from '$lib/utils/format';

  let query = $state('');
  let searchResults = $state<StockSearchResult[]>([]);
  let selectedStock = $state<StockQuoteResponse | null>(null);
  let loading = $state(false);
  let error = $state('');
  let searchTimeout: ReturnType<typeof setTimeout>;

  function handleSearch() {
    clearTimeout(searchTimeout);
    error = '';

    if (query.length < 1) {
      searchResults = [];
      return;
    }

    searchTimeout = setTimeout(async () => {
      try {
        loading = true;
        searchResults = await searchStocks(query);
      } catch (e) {
        error = e instanceof Error ? e.message : 'Search failed';
      } finally {
        loading = false;
      }
    }, 300);
  }

  async function selectStock(ticker: string) {
    try {
      loading = true;
      error = '';
      selectedStock = await getStockQuote(ticker);
      searchResults = [];
      query = '';
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load stock';
    } finally {
      loading = false;
    }
  }
</script>

<div class="dashboard">
  <h2>Stock Search</h2>

  <div class="search-box">
    <input
      type="text"
      placeholder="Search stocks (e.g., AAPL, SAP.DE)..."
      bind:value={query}
      oninput={handleSearch}
    />
    {#if loading}
      <span class="loading">Loading...</span>
    {/if}
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

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

  {#if selectedStock}
    {@const q = selectedStock.quote}
    {@const s = selectedStock.stock}
    <div class="stock-detail">
      <div class="stock-header">
        <h3>{s.ticker} — {s.name}</h3>
        <span class="exchange-badge">{s.exchange} · {s.currency}</span>
      </div>
      <div class="price-row">
        <span class="price">{formatCents(q.price_cents, q.currency)}</span>
        <span class="change" class:positive={q.change_percent >= 0} class:negative={q.change_percent < 0}>
          {formatPercent(q.change_percent)}
        </span>
      </div>
      <div class="details-grid">
        <div class="detail">
          <span class="label">Open</span>
          <span class="value">{formatCents(q.open_cents, q.currency)}</span>
        </div>
        <div class="detail">
          <span class="label">High</span>
          <span class="value">{formatCents(q.high_cents, q.currency)}</span>
        </div>
        <div class="detail">
          <span class="label">Low</span>
          <span class="value">{formatCents(q.low_cents, q.currency)}</span>
        </div>
        <div class="detail">
          <span class="label">Volume</span>
          <span class="value">{formatVolume(q.volume)}</span>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .dashboard {
    max-width: 800px;
  }

  h2 {
    color: #e1e4e8;
    margin-bottom: 1rem;
  }

  .search-box {
    position: relative;
    margin-bottom: 1rem;
  }

  .search-box input {
    width: 100%;
    padding: 0.75rem 1rem;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    color: #e1e4e8;
    font-size: 1rem;
    outline: none;
    box-sizing: border-box;
  }

  .search-box input:focus {
    border-color: #58a6ff;
  }

  .loading {
    position: absolute;
    right: 1rem;
    top: 50%;
    transform: translateY(-50%);
    color: #8b949e;
    font-size: 0.85rem;
  }

  .error {
    background: #3d1f1f;
    border: 1px solid #f85149;
    color: #f85149;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    font-size: 0.9rem;
  }

  .search-results {
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    overflow: hidden;
    margin-bottom: 1rem;
  }

  .result-item {
    display: flex;
    align-items: center;
    gap: 1rem;
    width: 100%;
    padding: 0.6rem 1rem;
    background: none;
    border: none;
    border-bottom: 1px solid #21262d;
    color: #e1e4e8;
    cursor: pointer;
    text-align: left;
    font-size: 0.9rem;
  }

  .result-item:hover {
    background: #1c2128;
  }

  .result-item:last-child {
    border-bottom: none;
  }

  .result-item .ticker {
    font-weight: 600;
    color: #58a6ff;
    min-width: 80px;
  }

  .result-item .name {
    flex: 1;
    color: #8b949e;
  }

  .result-item .exchange {
    color: #484f58;
    font-size: 0.8rem;
  }

  .stock-detail {
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 8px;
    padding: 1.5rem;
  }

  .stock-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .stock-header h3 {
    margin: 0;
    color: #e1e4e8;
  }

  .exchange-badge {
    color: #8b949e;
    font-size: 0.85rem;
  }

  .price-row {
    display: flex;
    align-items: baseline;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .price {
    font-size: 2rem;
    font-weight: 700;
    color: #e1e4e8;
  }

  .change {
    font-size: 1.1rem;
    font-weight: 600;
  }

  .positive {
    color: #3fb950;
  }

  .negative {
    color: #f85149;
  }

  .details-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
  }

  .detail .label {
    display: block;
    color: #8b949e;
    font-size: 0.8rem;
    margin-bottom: 0.25rem;
  }

  .detail .value {
    font-weight: 500;
  }
</style>
