<script lang="ts">
  import {
    getWatchlists,
    getWatchlist,
    createWatchlist,
    deleteWatchlist,
    addStockToWatchlist,
    removeStockFromWatchlist,
    type Watchlist,
  } from '$lib/api/watchlists';
  import { searchStocks, type StockSearchResult } from '$lib/api/stocks';
  import { formatCents, formatPercent, formatVolume } from '$lib/utils/format';

  let watchlists = $state<Watchlist[]>([]);
  let activeWatchlist = $state<Watchlist | null>(null);
  let loading = $state(false);
  let error = $state('');

  // Add stock search
  let addQuery = $state('');
  let addResults = $state<StockSearchResult[]>([]);
  let addSearchTimeout: ReturnType<typeof setTimeout>;
  let showAddSearch = $state(false);

  // New watchlist
  let newWatchlistName = $state('');
  let showNewForm = $state(false);

  // Sort
  type SortKey = 'ticker' | 'name' | 'price' | 'change' | 'volume';
  let sortKey = $state<SortKey>('ticker');
  let sortAsc = $state(true);

  async function loadWatchlists() {
    try {
      loading = true;
      watchlists = await getWatchlists();
      if (watchlists.length > 0 && !activeWatchlist) {
        await selectWatchlist(watchlists[0].id);
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load watchlists';
    } finally {
      loading = false;
    }
  }

  async function selectWatchlist(id: number) {
    try {
      loading = true;
      error = '';
      activeWatchlist = await getWatchlist(id);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load watchlist';
    } finally {
      loading = false;
    }
  }

  async function handleCreateWatchlist() {
    if (!newWatchlistName.trim()) return;
    try {
      const wl = await createWatchlist(newWatchlistName.trim());
      watchlists = [...watchlists, wl];
      newWatchlistName = '';
      showNewForm = false;
      await selectWatchlist(wl.id);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create watchlist';
    }
  }

  async function handleDeleteWatchlist(id: number) {
    try {
      await deleteWatchlist(id);
      watchlists = watchlists.filter((w) => w.id !== id);
      if (activeWatchlist?.id === id) {
        activeWatchlist = null;
        if (watchlists.length > 0) {
          await selectWatchlist(watchlists[0].id);
        }
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete watchlist';
    }
  }

  function handleAddSearch() {
    clearTimeout(addSearchTimeout);
    if (addQuery.length < 1) {
      addResults = [];
      return;
    }
    addSearchTimeout = setTimeout(async () => {
      try {
        addResults = await searchStocks(addQuery);
      } catch (e) {
        error = e instanceof Error ? e.message : 'Search failed';
      }
    }, 300);
  }

  async function handleAddStock(ticker: string) {
    if (!activeWatchlist) return;
    try {
      await addStockToWatchlist(activeWatchlist.id, ticker);
      addQuery = '';
      addResults = [];
      showAddSearch = false;
      await selectWatchlist(activeWatchlist.id);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to add stock';
    }
  }

  async function handleRemoveStock(ticker: string) {
    if (!activeWatchlist) return;
    try {
      await removeStockFromWatchlist(activeWatchlist.id, ticker);
      await selectWatchlist(activeWatchlist.id);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to remove stock';
    }
  }

  function sortItems() {
    if (!activeWatchlist?.items) return [];
    return [...activeWatchlist.items].sort((a, b) => {
      let cmp = 0;
      switch (sortKey) {
        case 'ticker':
          cmp = (a.stock?.ticker ?? '').localeCompare(b.stock?.ticker ?? '');
          break;
        case 'name':
          cmp = (a.stock?.name ?? '').localeCompare(b.stock?.name ?? '');
          break;
        case 'price':
          cmp = (a.quote?.price_cents ?? 0) - (b.quote?.price_cents ?? 0);
          break;
        case 'change':
          cmp = (a.quote?.change_percent ?? 0) - (b.quote?.change_percent ?? 0);
          break;
        case 'volume':
          cmp = (a.quote?.volume ?? 0) - (b.quote?.volume ?? 0);
          break;
      }
      return sortAsc ? cmp : -cmp;
    });
  }

  function toggleSort(key: SortKey) {
    if (sortKey === key) {
      sortAsc = !sortAsc;
    } else {
      sortKey = key;
      sortAsc = true;
    }
  }

  function sortIndicator(key: SortKey): string {
    if (sortKey !== key) return '';
    return sortAsc ? ' ▲' : ' ▼';
  }

  // Load on mount
  $effect(() => {
    loadWatchlists();
  });

  let sortedItems = $derived(sortItems());
</script>

<div class="watchlist-page">
  <div class="header">
    <h2>Watchlist</h2>
    <button class="btn-add" onclick={() => { showAddSearch = !showAddSearch; showNewForm = false; }}>
      + Add Stock
    </button>
    <button class="btn-new" onclick={() => { showNewForm = !showNewForm; showAddSearch = false; }}>
      + New List
    </button>
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  <!-- Watchlist Tabs -->
  {#if watchlists.length > 0}
    <div class="tabs">
      {#each watchlists as wl}
        <button
          class="tab"
          class:active={activeWatchlist?.id === wl.id}
          onclick={() => selectWatchlist(wl.id)}
        >
          {wl.name}
          <button class="tab-delete" onclick={(e) => { e.stopPropagation(); handleDeleteWatchlist(wl.id); }}>×</button>
        </button>
      {/each}
    </div>
  {/if}

  <!-- New Watchlist Form -->
  {#if showNewForm}
    <div class="add-form">
      <input
        type="text"
        placeholder="Watchlist name..."
        bind:value={newWatchlistName}
        onkeydown={(e) => e.key === 'Enter' && handleCreateWatchlist()}
      />
      <button onclick={handleCreateWatchlist}>Create</button>
    </div>
  {/if}

  <!-- Add Stock Search -->
  {#if showAddSearch && activeWatchlist}
    <div class="add-form">
      <input
        type="text"
        placeholder="Search stock to add (e.g., AAPL)..."
        bind:value={addQuery}
        oninput={handleAddSearch}
      />
      {#if addResults.length > 0}
        <div class="search-results">
          {#each addResults as result}
            <button class="result-item" onclick={() => handleAddStock(result.ticker)}>
              <span class="ticker">{result.ticker}</span>
              <span class="name">{result.name}</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <!-- Stock Table -->
  {#if loading}
    <p class="loading-text">Loading...</p>
  {:else if activeWatchlist && sortedItems.length > 0}
    <table class="stock-table">
      <thead>
        <tr>
          <th onclick={() => toggleSort('ticker')}>Ticker{sortIndicator('ticker')}</th>
          <th onclick={() => toggleSort('name')}>Name{sortIndicator('name')}</th>
          <th class="num" onclick={() => toggleSort('price')}>Price{sortIndicator('price')}</th>
          <th class="num" onclick={() => toggleSort('change')}>Change{sortIndicator('change')}</th>
          <th class="num" onclick={() => toggleSort('volume')}>Volume{sortIndicator('volume')}</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each sortedItems as item}
          <tr>
            <td class="ticker-cell">{item.stock?.ticker ?? '—'}</td>
            <td class="name-cell">{item.stock?.name ?? '—'}</td>
            <td class="num">
              {item.quote ? formatCents(item.quote.price_cents, item.quote.currency) : '—'}
            </td>
            <td class="num">
              {#if item.quote}
                <span class:positive={item.quote.change_percent >= 0} class:negative={item.quote.change_percent < 0}>
                  {formatPercent(item.quote.change_percent)}
                </span>
              {:else}
                —
              {/if}
            </td>
            <td class="num">{item.quote ? formatVolume(item.quote.volume) : '—'}</td>
            <td>
              <button class="btn-remove" onclick={() => handleRemoveStock(item.stock?.ticker ?? '')}>
                Remove
              </button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else if activeWatchlist}
    <p class="empty-text">No stocks in this watchlist. Click "+ Add Stock" to get started.</p>
  {:else}
    <p class="empty-text">No watchlists yet. Click "+ New List" to create one.</p>
  {/if}
</div>

<style>
  .watchlist-page {
    max-width: 1000px;
  }

  .header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .header h2 {
    margin: 0;
    flex: 1;
  }

  .btn-add, .btn-new {
    padding: 0.5rem 1rem;
    background: #238636;
    color: #fff;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.85rem;
  }

  .btn-new {
    background: #30363d;
  }

  .btn-add:hover { background: #2ea043; }
  .btn-new:hover { background: #484f58; }

  .error {
    background: #3d1f1f;
    border: 1px solid #f85149;
    color: #f85149;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    font-size: 0.9rem;
  }

  .tabs {
    display: flex;
    gap: 0.25rem;
    margin-bottom: 1rem;
    border-bottom: 1px solid #30363d;
    padding-bottom: 0;
  }

  .tab {
    padding: 0.5rem 1rem;
    background: none;
    border: 1px solid transparent;
    border-bottom: 2px solid transparent;
    color: #8b949e;
    cursor: pointer;
    font-size: 0.9rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .tab.active {
    color: #e1e4e8;
    border-bottom-color: #58a6ff;
  }

  .tab:hover { color: #e1e4e8; }

  .tab-delete {
    font-size: 0.75rem;
    color: #484f58;
    cursor: pointer;
    background: none;
    border: none;
    padding: 0;
    line-height: 1;
  }
  .tab-delete:hover { color: #f85149; }

  .add-form {
    margin-bottom: 1rem;
    position: relative;
  }

  .add-form input {
    width: 100%;
    padding: 0.6rem 1rem;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    color: #e1e4e8;
    font-size: 0.9rem;
    outline: none;
    box-sizing: border-box;
  }

  .add-form input:focus { border-color: #58a6ff; }

  .add-form button {
    margin-top: 0.5rem;
    padding: 0.4rem 1rem;
    background: #238636;
    color: #fff;
    border: none;
    border-radius: 6px;
    cursor: pointer;
  }

  .search-results {
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 0 0 6px 6px;
    border-top: none;
    max-height: 200px;
    overflow-y: auto;
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
  .result-item .name { color: #8b949e; }

  .stock-table {
    width: 100%;
    border-collapse: collapse;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    overflow: hidden;
  }

  .stock-table th {
    padding: 0.6rem 1rem;
    text-align: left;
    color: #8b949e;
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: uppercase;
    border-bottom: 1px solid #30363d;
    cursor: pointer;
    user-select: none;
  }

  .stock-table th:hover { color: #e1e4e8; }
  .stock-table th.num, .stock-table td.num { text-align: right; }

  .stock-table td {
    padding: 0.6rem 1rem;
    border-bottom: 1px solid #21262d;
    font-size: 0.9rem;
  }

  .stock-table tr:last-child td { border-bottom: none; }
  .stock-table tr:hover { background: #1c2128; }

  .ticker-cell { font-weight: 600; color: #58a6ff; }
  .name-cell { color: #8b949e; }

  .positive { color: #3fb950; }
  .negative { color: #f85149; }

  .btn-remove {
    padding: 0.25rem 0.5rem;
    background: none;
    border: 1px solid #30363d;
    color: #8b949e;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.8rem;
  }

  .btn-remove:hover {
    border-color: #f85149;
    color: #f85149;
  }

  .loading-text, .empty-text {
    color: #8b949e;
    text-align: center;
    padding: 2rem;
  }
</style>
