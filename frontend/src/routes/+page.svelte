<script lang="ts">
  import {
    getDashboard,
    getPerformance,
    getHoldingsPL,
    type DashboardData,
    type PortfolioSnapshot,
    type HoldingPL,
  } from '$lib/api/dashboard';
  import {
    getHealthScore,
    type HealthScore,
  } from '$lib/api/health';
  import { formatCents, formatPercent } from '$lib/utils/format';
  import PerformanceChart from '$lib/components/PerformanceChart.svelte';
  import PLBarChart from '$lib/components/PLBarChart.svelte';
  import Skeleton from '$lib/components/Skeleton.svelte';

  let dashboard = $state<DashboardData | null>(null);
  let snapshots = $state<PortfolioSnapshot[]>([]);
  let holdingsPL = $state<HoldingPL[]>([]);
  let healthScore = $state<HealthScore | null>(null);
  let loading = $state(true);
  let healthLoading = $state(false);
  let error = $state('');
  let selectedPortfolioId = $state<number | null>(null);
  let selectedRange = $state('3m');
  let selectedCurrency = $state('EUR');

  const ranges = ['1m', '3m', '6m', 'ytd', '1y', 'all'];

  function gradeColor(grade: string): string {
    switch (grade) {
      case 'A': return '#3fb950';
      case 'B': return '#56d364';
      case 'C': return '#d29922';
      case 'D': return '#db6d28';
      default: return '#f85149';
    }
  }

  function severityColor(severity: string): string {
    switch (severity) {
      case 'critical': return '#f85149';
      case 'warning': return '#d29922';
      default: return '#58a6ff';
    }
  }

  function actionColor(action: string): string {
    switch (action) {
      case 'BUY': return '#3fb950';
      case 'HOLD': return '#58a6ff';
      default: return '#d29922';
    }
  }

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
      healthLoading = true;
      const [perf, pl] = await Promise.all([
        getPerformance(selectedPortfolioId, selectedRange),
        getHoldingsPL(selectedPortfolioId),
      ]);
      snapshots = perf;
      holdingsPL = pl;

      getHealthScore(selectedPortfolioId).then(h => {
        healthScore = h;
      }).catch(() => {
        healthScore = null;
      }).finally(() => {
        healthLoading = false;
      });
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load charts';
      healthLoading = false;
    }
  }

  async function selectPortfolio(id: number, currency: string) {
    selectedPortfolioId = id;
    selectedCurrency = currency;
    healthScore = null;
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
    <div class="skeleton-grid">
      <Skeleton height="5rem" />
      <Skeleton height="5rem" />
    </div>
    <Skeleton height="12rem" />
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

    <!-- Health Score Widget -->
    {#if healthLoading}
      <div class="health-row">
        <div class="card health-card"><Skeleton height="8rem" /></div>
        <div class="card alerts-card"><Skeleton height="8rem" /></div>
      </div>
    {:else if healthScore}
      <div class="health-row">
        <div class="card health-card">
          <span class="card-label">Portfolio Health</span>
          <div class="health-body">
            <svg viewBox="0 0 100 100" class="score-ring">
              <circle cx="50" cy="50" r="40" fill="none" stroke="#30363d" stroke-width="8" />
              <circle
                cx="50" cy="50" r="40" fill="none"
                stroke={gradeColor(healthScore.grade)}
                stroke-width="8"
                stroke-linecap="round"
                stroke-dasharray="{healthScore.overall_score * 2.51} 251"
                transform="rotate(-90 50 50)"
              />
              <text x="50" y="45" text-anchor="middle" fill="#e1e4e8" font-size="22" font-weight="700">{healthScore.overall_score}</text>
              <text x="50" y="62" text-anchor="middle" fill={gradeColor(healthScore.grade)} font-size="14" font-weight="600">{healthScore.grade}</text>
            </svg>
            <div class="components">
              {#each healthScore.components as c}
                <div class="component">
                  <div class="comp-header">
                    <span class="comp-name">{c.name}</span>
                    <span class="comp-score">{c.score}</span>
                  </div>
                  <div class="comp-bar-bg">
                    <div class="comp-bar" style="width:{c.score}%;background:{c.score >= 70 ? '#3fb950' : c.score >= 40 ? '#d29922' : '#f85149'}"></div>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        </div>

        <div class="card alerts-card">
          <span class="card-label">Alerts ({healthScore.alerts.length})</span>
          <div class="alerts-list">
            {#if healthScore.alerts.length === 0}
              <p class="empty-text">No alerts</p>
            {:else}
              {#each healthScore.alerts as alert}
                <div class="alert-item" style="border-left: 3px solid {severityColor(alert.severity)}">
                  <div class="alert-header">
                    <span class="alert-dot" style="background:{severityColor(alert.severity)}"></span>
                    <span class="alert-title">{alert.title}</span>
                    {#if alert.ticker}
                      <span class="alert-ticker">{alert.ticker}</span>
                    {/if}
                  </div>
                  <p class="alert-msg">{alert.message}</p>
                </div>
              {/each}
            {/if}
          </div>
        </div>
      </div>

      {#if healthScore.top_picks.length > 0}
        <div class="top-picks-section">
          <h3>Top Picks</h3>
          <div class="picks-grid">
            {#each healthScore.top_picks as pick}
              <div class="pick-card">
                <div class="pick-header">
                  <span class="pick-ticker">{pick.ticker}</span>
                  <span class="pick-action" style="background:{actionColor(pick.action)}">{pick.action}</span>
                </div>
                <p class="pick-name">{pick.name}</p>
                <p class="pick-reason">{pick.reason}</p>
                <div class="pick-stats">
                  <span>MOS: {(pick.margin_of_safety * 100).toFixed(0)}%</span>
                  <span>Quality: {pick.quality_score}/100</span>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    {/if}

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

  .skeleton-grid {
    display: flex; gap: 1rem; margin-bottom: 1.5rem;
  }
  .skeleton-grid > :global(*) { flex: 1; }

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

  .health-row {
    display: flex; gap: 1rem; margin-bottom: 1.5rem;
  }
  .health-card { flex: 1; }
  .alerts-card { flex: 1; max-height: 300px; overflow-y: auto; }

  .health-body {
    display: flex; align-items: center; gap: 1.5rem;
  }
  .score-ring { width: 110px; height: 110px; flex-shrink: 0; }

  .components { flex: 1; display: flex; flex-direction: column; gap: 0.5rem; }
  .comp-header { display: flex; justify-content: space-between; font-size: 0.8rem; margin-bottom: 0.2rem; }
  .comp-name { color: #8b949e; }
  .comp-score { color: #e1e4e8; font-weight: 600; }
  .comp-bar-bg { height: 4px; background: #21262d; border-radius: 2px; }
  .comp-bar { height: 100%; border-radius: 2px; transition: width 0.3s; }

  .alerts-list { display: flex; flex-direction: column; gap: 0.5rem; }
  .alert-item {
    padding: 0.5rem 0.75rem; background: #0d1117; border-radius: 4px;
  }
  .alert-header { display: flex; align-items: center; gap: 0.5rem; }
  .alert-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
  .alert-title { font-size: 0.85rem; font-weight: 600; }
  .alert-ticker { font-size: 0.75rem; color: #8b949e; margin-left: auto; }
  .alert-msg { font-size: 0.8rem; color: #8b949e; margin: 0.25rem 0 0; }

  .top-picks-section { margin-bottom: 1.5rem; }
  .picks-grid {
    display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 0.75rem;
  }
  .pick-card {
    padding: 1rem; background: #161b22; border: 1px solid #30363d; border-radius: 6px;
  }
  .pick-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.25rem; }
  .pick-ticker { font-weight: 700; font-size: 1rem; }
  .pick-action {
    font-size: 0.65rem; font-weight: 700; padding: 0.15rem 0.5rem;
    border-radius: 3px; color: #0d1117; text-transform: uppercase;
  }
  .pick-name { font-size: 0.8rem; color: #8b949e; margin: 0 0 0.25rem; }
  .pick-reason { font-size: 0.8rem; color: #c9d1d9; margin: 0 0 0.5rem; }
  .pick-stats { display: flex; gap: 1rem; font-size: 0.75rem; color: #8b949e; }

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

  @media (max-width: 768px) {
    .overview-cards { flex-direction: column; }
    .health-row { flex-direction: column; }
    .health-body { flex-direction: column; }
    .skeleton-grid { flex-direction: column; }
  }
</style>
