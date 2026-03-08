<script lang="ts">
  import {
    getDashboard,
    getPerformance,
    getHoldingsPL,
    type DashboardData,
    type PortfolioSnapshot,
    type HoldingPL,
  } from '$lib/api/dashboard';
  import { formatCents, formatPercent } from '$lib/utils/format';
  import PerformanceChart from '$lib/components/PerformanceChart.svelte';
  import PLBarChart from '$lib/components/PLBarChart.svelte';

  let dashboard = $state<DashboardData | null>(null);
  let snapshots = $state<PortfolioSnapshot[]>([]);
  let holdingsPL = $state<HoldingPL[]>([]);
  let loading = $state(true);
  let error = $state('');
  let selectedPortfolioId = $state<number | null>(null);
  let selectedRange = $state('3m');
  let selectedCurrency = $state('EUR');

  const ranges = ['1m', '3m', '6m', 'ytd', '1y', 'all'];

  async function loadDashboard() {
    try {
      loading = true;
      dashboard = await getDashboard();
      if (dashboard.summaries.length > 0) {
        selectedPortfolioId = dashboard.summaries[0].portfolio_id;
        selectedCurrency = dashboard.summaries[0].currency;
        await loadCharts();
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load dashboard';
    } finally {
      loading = false;
    }
  }

  async function loadCharts() {
    if (!selectedPortfolioId) return;
    try {
      const [perf, pl] = await Promise.all([
        getPerformance(selectedPortfolioId, selectedRange),
        getHoldingsPL(selectedPortfolioId),
      ]);
      snapshots = perf;
      holdingsPL = pl;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load charts';
    }
  }

  async function selectPortfolio(id: number, currency: string) {
    selectedPortfolioId = id;
    selectedCurrency = currency;
    await loadCharts();
  }

  async function changeRange(range: string) {
    selectedRange = range;
    await loadCharts();
  }

  $effect(() => {
    loadDashboard();
  });
</script>

<div class="dashboard">
  <h2>Dashboard</h2>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  {#if loading}
    <p class="loading-text">Loading dashboard...</p>
  {:else if dashboard}
    <!-- Total Overview -->
    <div class="overview-cards">
      <div class="card">
        <span class="card-label">Total Value</span>
        <span class="card-value">{formatCents(dashboard.total_value_cents, selectedCurrency)}</span>
      </div>
      <div class="card">
        <span class="card-label">Total P&L</span>
        <span class="card-value" class:positive={dashboard.total_pl_cents >= 0} class:negative={dashboard.total_pl_cents < 0}>
          {formatCents(dashboard.total_pl_cents, selectedCurrency)} ({formatPercent(dashboard.total_pl_percent)})
        </span>
      </div>
    </div>

    <!-- Portfolio Cards -->
    {#if dashboard.summaries.length > 0}
      <div class="portfolio-cards">
        {#each dashboard.summaries as s}
          <button
            class="portfolio-card"
            class:active={selectedPortfolioId === s.portfolio_id}
            onclick={() => selectPortfolio(s.portfolio_id, s.currency)}
          >
            <div class="pc-header">
              <span class="pc-name">{s.portfolio_name}</span>
              <span class="pc-currency">{s.currency}</span>
            </div>
            <div class="pc-value">{formatCents(s.total_value_cents, s.currency)}</div>
            <div class="pc-pl" class:positive={s.unrealized_pl_cents >= 0} class:negative={s.unrealized_pl_cents < 0}>
              {formatCents(s.unrealized_pl_cents, s.currency)} ({formatPercent(s.unrealized_pl_percent)})
            </div>
            <div class="pc-day" class:positive={s.day_change_cents >= 0} class:negative={s.day_change_cents < 0}>
              Today: {formatCents(s.day_change_cents, s.currency)} ({formatPercent(s.day_change_percent)})
            </div>
            <div class="pc-count">{s.holdings_count} holdings</div>
          </button>
        {/each}
      </div>
    {:else}
      <p class="empty-text">No portfolios yet. Create one on the <a href="/portfolio">Portfolio</a> page.</p>
    {/if}

    <!-- Charts Section -->
    {#if selectedPortfolioId}
      <div class="charts-section">
        <div class="chart-header">
          <h3>Performance</h3>
          <div class="range-buttons">
            {#each ranges as r}
              <button
                class="range-btn"
                class:active={selectedRange === r}
                onclick={() => changeRange(r)}
              >
                {r.toUpperCase()}
              </button>
            {/each}
          </div>
        </div>
        <PerformanceChart {snapshots} currency={selectedCurrency} />
      </div>

      {#if holdingsPL.length > 0}
        <div class="charts-section">
          <h3>P&L by Stock</h3>
          <PLBarChart holdings={holdingsPL} currency={selectedCurrency} />
        </div>
      {/if}
    {/if}
  {/if}
</div>

<style>
  .dashboard { max-width: 1100px; }
  .dashboard h2 { margin: 0 0 1.5rem; }
  .dashboard h3 { margin: 0 0 1rem; font-size: 1rem; }

  .error {
    background: #3d1f1f; border: 1px solid #f85149; color: #f85149;
    padding: 0.5rem 1rem; border-radius: 6px; margin-bottom: 1rem;
  }

  .overview-cards {
    display: flex; gap: 1rem; margin-bottom: 1.5rem;
  }
  .card {
    flex: 1; padding: 1.25rem; background: #161b22;
    border: 1px solid #30363d; border-radius: 6px;
    display: flex; flex-direction: column; gap: 0.5rem;
  }
  .card-label { font-size: 0.75rem; color: #8b949e; text-transform: uppercase; }
  .card-value { font-size: 1.5rem; font-weight: 700; }

  .portfolio-cards {
    display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem; margin-bottom: 1.5rem;
  }
  .portfolio-card {
    padding: 1rem; background: #161b22; border: 1px solid #30363d;
    border-radius: 6px; cursor: pointer; text-align: left;
    color: #e1e4e8; font-family: inherit; font-size: inherit;
    transition: border-color 0.15s;
  }
  .portfolio-card:hover { border-color: #484f58; }
  .portfolio-card.active { border-color: #58a6ff; }
  .pc-header { display: flex; justify-content: space-between; margin-bottom: 0.5rem; }
  .pc-name { font-weight: 600; }
  .pc-currency { font-size: 0.8rem; color: #8b949e; }
  .pc-value { font-size: 1.25rem; font-weight: 700; margin-bottom: 0.25rem; }
  .pc-pl { font-size: 0.9rem; margin-bottom: 0.25rem; }
  .pc-day { font-size: 0.8rem; color: #8b949e; margin-bottom: 0.25rem; }
  .pc-count { font-size: 0.75rem; color: #484f58; }

  .positive { color: #3fb950; }
  .negative { color: #f85149; }

  .charts-section { margin-bottom: 1.5rem; }
  .chart-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1rem; }
  .range-buttons { display: flex; gap: 0.25rem; }
  .range-btn {
    padding: 0.3rem 0.6rem; background: #21262d; border: 1px solid #30363d;
    border-radius: 4px; color: #8b949e; cursor: pointer; font-size: 0.75rem;
  }
  .range-btn.active { background: #30363d; color: #e1e4e8; border-color: #58a6ff; }
  .range-btn:hover { color: #e1e4e8; }

  .loading-text, .empty-text { color: #8b949e; text-align: center; padding: 2rem; }
  .empty-text a { color: #58a6ff; }
</style>
