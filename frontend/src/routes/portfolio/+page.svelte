<script lang="ts">
  import {
    getPortfolios,
    getPortfolio,
    createPortfolio,
    deletePortfolio,
    addTrade,
    deleteTrade,
    listTrades,
    getHoldings,
    type Portfolio,
    type PortfolioStrategy,
    type Trade,
    type Holding,
  } from '$lib/api/portfolios';
  import { searchStocks, type StockSearchResult } from '$lib/api/stocks';
  import { formatCents, formatPercent } from '$lib/utils/format';

  let portfolios = $state<Portfolio[]>([]);
  let activePortfolio = $state<Portfolio | null>(null);
  let holdings = $state<Holding[]>([]);
  let trades = $state<Trade[]>([]);
  let tradesTotal = $state(0);
  let loading = $state(false);
  let error = $state('');
  let activeTab = $state<'holdings' | 'trades'>('holdings');

  // New portfolio form
  let showNewForm = $state(false);
  let newName = $state('');
  let newCurrency = $state('EUR');
  let newStrategy = $state<PortfolioStrategy>('VALUE');

  // New trade form
  let showTradeForm = $state(false);
  let tradeQuery = $state('');
  let tradeResults = $state<StockSearchResult[]>([]);
  let tradeSearchTimeout: ReturnType<typeof setTimeout>;
  let selectedTicker = $state('');
  let tradeType = $state<'BUY' | 'SELL'>('BUY');
  let tradeAmount = $state('');
  let tradePrice = $state('');
  let tradeFee = $state('1.00');
  let tradeCurrency = $state('USD');
  let tradeDate = $state(new Date().toISOString().split('T')[0]);
  let tradeNotes = $state('');

  async function loadPortfolios() {
    try {
      loading = true;
      portfolios = await getPortfolios();
      if (portfolios.length > 0 && !activePortfolio) {
        await selectPortfolio(portfolios[0].id);
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load portfolios';
    } finally {
      loading = false;
    }
  }

  async function selectPortfolio(id: number) {
    try {
      loading = true;
      error = '';
      activePortfolio = await getPortfolio(id);
      await loadHoldings();
      await loadTrades();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load portfolio';
    } finally {
      loading = false;
    }
  }

  async function loadHoldings() {
    if (!activePortfolio) return;
    try {
      holdings = await getHoldings(activePortfolio.id);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load holdings';
    }
  }

  async function loadTrades() {
    if (!activePortfolio) return;
    try {
      const resp = await listTrades(activePortfolio.id);
      trades = resp.trades;
      tradesTotal = resp.total;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load trades';
    }
  }

  async function handleCreatePortfolio() {
    if (!newName.trim()) return;
    try {
      const p = await createPortfolio(newName.trim(), '', newCurrency, newStrategy);
      portfolios = [...portfolios, p];
      newName = '';
      showNewForm = false;
      await selectPortfolio(p.id);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create portfolio';
    }
  }

  async function handleDeletePortfolio(id: number) {
    try {
      await deletePortfolio(id);
      portfolios = portfolios.filter((p) => p.id !== id);
      if (activePortfolio?.id === id) {
        activePortfolio = null;
        holdings = [];
        trades = [];
        if (portfolios.length > 0) {
          await selectPortfolio(portfolios[0].id);
        }
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete portfolio';
    }
  }

  function handleTradeSearch() {
    clearTimeout(tradeSearchTimeout);
    if (tradeQuery.length < 1) {
      tradeResults = [];
      return;
    }
    tradeSearchTimeout = setTimeout(async () => {
      try {
        tradeResults = await searchStocks(tradeQuery);
      } catch (e) {
        error = e instanceof Error ? e.message : 'Search failed';
      }
    }, 300);
  }

  function selectTradeStock(ticker: string) {
    selectedTicker = ticker;
    tradeQuery = ticker;
    tradeResults = [];
  }

  let computedQuantity = $derived(() => {
    const amount = parseFloat(tradeAmount);
    const price = parseFloat(tradePrice);
    if (isNaN(amount) || amount <= 0 || isNaN(price) || price <= 0) return 0;
    return amount / price;
  });

  async function handleSubmitTrade() {
    if (!activePortfolio || !selectedTicker) return;
    const amount = parseFloat(tradeAmount);
    const price = parseFloat(tradePrice);
    const fee = parseFloat(tradeFee);
    if (isNaN(amount) || amount <= 0 || isNaN(price) || price <= 0) {
      error = 'Amount and price per share must be positive numbers';
      return;
    }
    const qty = amount / price;
    try {
      const executedAt = new Date(tradeDate).toISOString();
      await addTrade(
        activePortfolio.id,
        selectedTicker,
        tradeType,
        qty,
        Math.round(price * 100),
        Math.round((isNaN(fee) ? 1 : fee) * 100),
        tradeCurrency,
        executedAt,
        tradeNotes,
      );
      // Reset form
      showTradeForm = false;
      selectedTicker = '';
      tradeQuery = '';
      tradeAmount = '';
      tradePrice = '';
      tradeFee = '1.00';
      tradeNotes = '';
      // Reload
      await loadHoldings();
      await loadTrades();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to add trade';
    }
  }

  async function handleDeleteTrade(tradeId: number) {
    if (!activePortfolio) return;
    try {
      await deleteTrade(activePortfolio.id, tradeId);
      await loadHoldings();
      await loadTrades();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete trade';
    }
  }

  // Compute totals
  let totalValue = $derived(holdings.reduce((sum, h) => sum + h.market_value_cents, 0));
  let totalPL = $derived(holdings.reduce((sum, h) => sum + h.unrealized_pl_cents, 0));
  let totalCost = $derived(holdings.reduce((sum, h) => sum + Math.round(h.avg_cost_cents * h.quantity), 0));
  let totalPLPct = $derived(totalCost > 0 ? totalPL / totalCost : 0);

  $effect(() => {
    loadPortfolios();
  });
</script>

<div class="portfolio-page">
  <div class="header">
    <h2>Portfolio</h2>
    <button class="btn-trade" onclick={() => { showTradeForm = !showTradeForm; showNewForm = false; }}>
      + New Trade
    </button>
    <button class="btn-new" onclick={() => { showNewForm = !showNewForm; showTradeForm = false; }}>
      + New Portfolio
    </button>
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  <!-- Portfolio Tabs -->
  {#if portfolios.length > 0}
    <div class="tabs">
      {#each portfolios as p}
        <button
          class="tab"
          class:active={activePortfolio?.id === p.id}
          onclick={() => selectPortfolio(p.id)}
        >
          {p.name} ({p.strategy || 'VALUE'})
          <span class="tab-delete" role="button" tabindex="0" onclick={(e) => { e.stopPropagation(); handleDeletePortfolio(p.id); }} onkeydown={(e) => { if (e.key === 'Enter') { e.stopPropagation(); handleDeletePortfolio(p.id); } }}>×</span>
        </button>
      {/each}
    </div>
  {/if}

  <!-- New Portfolio Form -->
  {#if showNewForm}
    <div class="add-form">
      <div class="form-row">
        <input type="text" placeholder="Portfolio name..." bind:value={newName}
          onkeydown={(e) => e.key === 'Enter' && handleCreatePortfolio()} />
        <select bind:value={newCurrency}>
          <option value="EUR">EUR</option>
          <option value="USD">USD</option>
        </select>
        <select bind:value={newStrategy}>
          <option value="VALUE">Value</option>
          <option value="TRADING">Trading</option>
          <option value="INDEX">Index/ETF</option>
        </select>
        <button onclick={handleCreatePortfolio}>Create</button>
      </div>
    </div>
  {/if}

  <!-- New Trade Form -->
  {#if showTradeForm && activePortfolio}
    <div class="trade-form">
      <h3>Record Trade</h3>
      <div class="form-row">
        <div class="form-group">
          <label for="trade-stock">Stock</label>
          <input id="trade-stock" type="text" placeholder="Search ticker..." bind:value={tradeQuery}
            oninput={handleTradeSearch} />
          {#if tradeResults.length > 0}
            <div class="search-results">
              {#each tradeResults as result}
                <button class="result-item" onclick={() => selectTradeStock(result.ticker)}>
                  <span class="ticker">{result.ticker}</span>
                  <span class="name">{result.name}</span>
                </button>
              {/each}
            </div>
          {/if}
        </div>
        <div class="form-group">
          <label for="trade-type">Type</label>
          <select id="trade-type" bind:value={tradeType}>
            <option value="BUY">BUY</option>
            <option value="SELL">SELL</option>
          </select>
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label for="trade-price">Price per share</label>
          <input id="trade-price" type="number" step="0.01" placeholder="175.00" bind:value={tradePrice} />
        </div>
        <div class="form-group">
          <label for="trade-amount">Amount ({tradeCurrency})</label>
          <input id="trade-amount" type="number" step="0.01" placeholder="200.00" bind:value={tradeAmount} />
        </div>
        {#if computedQuantity() > 0}
          <div class="form-group computed-qty">
            <label>Shares</label>
            <span class="computed-value">{computedQuantity().toFixed(6)}</span>
          </div>
        {/if}
        <div class="form-group">
          <label for="trade-fee">Fee</label>
          <input id="trade-fee" type="number" step="0.01" placeholder="1.00" bind:value={tradeFee} />
        </div>
        <div class="form-group">
          <label for="trade-currency">Currency</label>
          <select id="trade-currency" bind:value={tradeCurrency}>
            <option value="USD">USD</option>
            <option value="EUR">EUR</option>
          </select>
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label for="trade-date">Date</label>
          <input id="trade-date" type="date" bind:value={tradeDate} />
        </div>
        <div class="form-group flex-1">
          <label for="trade-notes">Notes</label>
          <input id="trade-notes" type="text" placeholder="Optional notes..." bind:value={tradeNotes} />
        </div>
      </div>
      <div class="form-actions">
        <button class="btn-submit" onclick={handleSubmitTrade}>
          {tradeType === 'BUY' ? 'Record Buy' : 'Record Sell'}
        </button>
        <button class="btn-cancel" onclick={() => { showTradeForm = false; }}>Cancel</button>
      </div>
    </div>
  {/if}

  {#if activePortfolio}
    <!-- Summary -->
    {#if holdings.length > 0}
      <div class="summary">
        <div class="summary-item">
          <span class="label">Market Value</span>
          <span class="value">{formatCents(totalValue, activePortfolio.currency)}</span>
        </div>
        <div class="summary-item">
          <span class="label">Unrealized P&L</span>
          <span class="value" class:positive={totalPL >= 0} class:negative={totalPL < 0}>
            {formatCents(totalPL, activePortfolio.currency)} ({formatPercent(totalPLPct)})
          </span>
        </div>
      </div>
    {/if}

    <!-- Sub-tabs: Holdings / Trades -->
    <div class="sub-tabs">
      <button class="sub-tab" class:active={activeTab === 'holdings'} onclick={() => activeTab = 'holdings'}>
        Holdings
      </button>
      <button class="sub-tab" class:active={activeTab === 'trades'} onclick={() => activeTab = 'trades'}>
        Trades ({tradesTotal})
      </button>
    </div>

    {#if loading}
      <p class="loading-text">Loading...</p>
    {:else if activeTab === 'holdings'}
      {#if holdings.length > 0}
        <table class="data-table">
          <thead>
            <tr>
              <th>Ticker</th>
              <th>Name</th>
              <th class="num">Qty</th>
              <th class="num">Avg Cost</th>
              <th class="num">Price</th>
              <th class="num">Value</th>
              <th class="num">P&L</th>
            </tr>
          </thead>
          <tbody>
            {#each holdings as h}
              <tr>
                <td class="ticker-cell">{h.stock.ticker}</td>
                <td class="name-cell">{h.stock.name}</td>
                <td class="num">{h.quantity.toFixed(4)}</td>
                <td class="num">{formatCents(h.avg_cost_cents, activePortfolio.currency)}</td>
                <td class="num">{formatCents(h.current_price_cents, activePortfolio.currency)}</td>
                <td class="num">{formatCents(h.market_value_cents, activePortfolio.currency)}</td>
                <td class="num">
                  <span class:positive={h.unrealized_pl_cents >= 0} class:negative={h.unrealized_pl_cents < 0}>
                    {formatCents(h.unrealized_pl_cents, activePortfolio.currency)}
                    ({formatPercent(h.unrealized_pl_percent)})
                  </span>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {:else}
        <p class="empty-text">No holdings yet. Record a trade to get started.</p>
      {/if}
    {:else}
      {#if trades.length > 0}
        <table class="data-table">
          <thead>
            <tr>
              <th>Date</th>
              <th>Type</th>
              <th>Ticker</th>
              <th class="num">Qty</th>
              <th class="num">Price</th>
              <th class="num">Fee</th>
              <th class="num">Total</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each trades as t}
              <tr>
                <td>{new Date(t.executed_at).toLocaleDateString()}</td>
                <td>
                  <span class="trade-type" class:buy={t.trade_type === 'BUY'} class:sell={t.trade_type === 'SELL'}>
                    {t.trade_type}
                  </span>
                </td>
                <td class="ticker-cell">{t.ticker}</td>
                <td class="num">{t.quantity.toFixed(4)}</td>
                <td class="num">{formatCents(t.price_cents, t.currency)}</td>
                <td class="num">{formatCents(t.fee_cents, t.currency)}</td>
                <td class="num">{formatCents(Math.round(t.quantity * t.price_cents) + t.fee_cents, t.currency)}</td>
                <td>
                  <button class="btn-remove" onclick={() => handleDeleteTrade(t.id)}>Delete</button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {:else}
        <p class="empty-text">No trades recorded yet.</p>
      {/if}
    {/if}
  {:else}
    <p class="empty-text">No portfolios yet. Click "+ New Portfolio" to create one.</p>
  {/if}
</div>

<style>
  .portfolio-page { max-width: 1100px; }

  .header { display: flex; align-items: center; gap: 1rem; margin-bottom: 1rem; }
  .header h2 { margin: 0; flex: 1; }

  .btn-trade, .btn-new {
    padding: 0.5rem 1rem; background: #238636; color: #fff;
    border: none; border-radius: 6px; cursor: pointer; font-size: 0.85rem;
  }
  .btn-new { background: #30363d; }
  .btn-trade:hover { background: #2ea043; }
  .btn-new:hover { background: #484f58; }

  .error {
    background: #3d1f1f; border: 1px solid #f85149; color: #f85149;
    padding: 0.5rem 1rem; border-radius: 6px; margin-bottom: 1rem; font-size: 0.9rem;
  }

  .tabs { display: flex; gap: 0.25rem; margin-bottom: 1rem; border-bottom: 1px solid #30363d; }
  .tab {
    padding: 0.5rem 1rem; background: none; border: 1px solid transparent;
    border-bottom: 2px solid transparent; color: #8b949e; cursor: pointer;
    font-size: 0.9rem; display: flex; align-items: center; gap: 0.5rem;
  }
  .tab.active { color: #e1e4e8; border-bottom-color: #58a6ff; }
  .tab:hover { color: #e1e4e8; }
  .tab-delete {
    font-size: 0.75rem; color: #484f58; cursor: pointer;
    background: none; border: none; padding: 0; line-height: 1;
  }
  .tab-delete:hover { color: #f85149; }

  .add-form { margin-bottom: 1rem; }
  .add-form input, .add-form select {
    padding: 0.6rem 1rem; background: #161b22; border: 1px solid #30363d;
    border-radius: 6px; color: #e1e4e8; font-size: 0.9rem; outline: none;
  }
  .add-form input:focus, .add-form select:focus { border-color: #58a6ff; }
  .add-form button {
    padding: 0.4rem 1rem; background: #238636; color: #fff;
    border: none; border-radius: 6px; cursor: pointer;
  }

  .form-row { display: flex; gap: 0.75rem; margin-bottom: 0.75rem; align-items: flex-end; }
  .form-group { display: flex; flex-direction: column; gap: 0.25rem; }
  .form-group.flex-1 { flex: 1; }
  .form-group label { font-size: 0.75rem; color: #8b949e; text-transform: uppercase; }
  .form-group input, .form-group select {
    padding: 0.5rem 0.75rem; background: #161b22; border: 1px solid #30363d;
    border-radius: 6px; color: #e1e4e8; font-size: 0.9rem; outline: none;
  }
  .form-group input:focus, .form-group select:focus { border-color: #58a6ff; }

  .trade-form {
    background: #161b22; border: 1px solid #30363d; border-radius: 6px;
    padding: 1.25rem; margin-bottom: 1rem;
  }
  .trade-form h3 { margin: 0 0 1rem; font-size: 1rem; }
  .form-actions { display: flex; gap: 0.5rem; }
  .btn-submit {
    padding: 0.5rem 1.5rem; background: #238636; color: #fff;
    border: none; border-radius: 6px; cursor: pointer;
  }
  .btn-submit:hover { background: #2ea043; }
  .btn-cancel {
    padding: 0.5rem 1rem; background: #30363d; color: #e1e4e8;
    border: none; border-radius: 6px; cursor: pointer;
  }
  .btn-cancel:hover { background: #484f58; }

  .search-results {
    background: #161b22; border: 1px solid #30363d; border-radius: 0 0 6px 6px;
    border-top: none; max-height: 150px; overflow-y: auto;
  }
  .result-item {
    display: flex; gap: 0.75rem; width: 100%; padding: 0.4rem 0.75rem;
    background: none; border: none; border-bottom: 1px solid #21262d;
    color: #e1e4e8; cursor: pointer; text-align: left; font-size: 0.85rem;
  }
  .result-item:hover { background: #1c2128; }
  .result-item:last-child { border-bottom: none; }
  .result-item .ticker { font-weight: 600; color: #58a6ff; min-width: 60px; }
  .result-item .name { color: #8b949e; }

  .summary {
    display: flex; gap: 2rem; padding: 1rem; background: #161b22;
    border: 1px solid #30363d; border-radius: 6px; margin-bottom: 1rem;
  }
  .summary-item { display: flex; flex-direction: column; gap: 0.25rem; }
  .summary-item .label { font-size: 0.75rem; color: #8b949e; text-transform: uppercase; }
  .summary-item .value { font-size: 1.25rem; font-weight: 600; }

  .sub-tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
  .sub-tab {
    padding: 0.4rem 1rem; background: #21262d; border: 1px solid #30363d;
    border-radius: 6px; color: #8b949e; cursor: pointer; font-size: 0.85rem;
  }
  .sub-tab.active { background: #30363d; color: #e1e4e8; border-color: #58a6ff; }
  .sub-tab:hover { color: #e1e4e8; }

  .data-table {
    width: 100%; border-collapse: collapse; background: #161b22;
    border: 1px solid #30363d; border-radius: 6px; overflow: hidden;
  }
  .data-table th {
    padding: 0.6rem 1rem; text-align: left; color: #8b949e;
    font-size: 0.8rem; font-weight: 600; text-transform: uppercase;
    border-bottom: 1px solid #30363d;
  }
  .data-table th.num, .data-table td.num { text-align: right; }
  .data-table td { padding: 0.6rem 1rem; border-bottom: 1px solid #21262d; font-size: 0.9rem; }
  .data-table tr:last-child td { border-bottom: none; }
  .data-table tr:hover { background: #1c2128; }

  .ticker-cell { font-weight: 600; color: #58a6ff; }
  .name-cell { color: #8b949e; }
  .positive { color: #3fb950; }
  .negative { color: #f85149; }

  .trade-type { font-weight: 600; font-size: 0.8rem; padding: 0.1rem 0.4rem; border-radius: 3px; }
  .trade-type.buy { background: #0d3321; color: #3fb950; }
  .trade-type.sell { background: #3d1f1f; color: #f85149; }

  .btn-remove {
    padding: 0.25rem 0.5rem; background: none; border: 1px solid #30363d;
    color: #8b949e; border-radius: 4px; cursor: pointer; font-size: 0.8rem;
  }
  .btn-remove:hover { border-color: #f85149; color: #f85149; }

  .computed-qty { justify-content: center; }
  .computed-value {
    font-size: 0.95rem; font-weight: 600; color: #58a6ff;
    padding: 0.5rem 0; line-height: 1.4;
  }

  .loading-text, .empty-text { color: #8b949e; text-align: center; padding: 2rem; }

  @media (max-width: 768px) {
    .form-row { flex-direction: column; }
    .summary { flex-direction: column; gap: 1rem; }
    .data-table { display: block; overflow-x: auto; }
    .header { flex-direction: column; align-items: flex-start; }
    .form-group input, .form-group select { width: 100%; }
  }
</style>
