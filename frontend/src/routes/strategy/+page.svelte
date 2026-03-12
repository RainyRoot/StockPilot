<script lang="ts">
  import { getPortfolios, getStrategyOverview, type Portfolio, type StrategyOverview } from '$lib/api/portfolios';
  import { formatCents, formatPercent } from '$lib/utils/format';

  let portfolios = $state<Portfolio[]>([]);
  let selectedId = $state<number | null>(null);
  let overview = $state<StrategyOverview | null>(null);
  let loading = $state(false);
  let error = $state('');

  async function init() {
    try {
      portfolios = await getPortfolios();
      if (portfolios.length > 0) {
        selectedId = portfolios[0].id;
        await loadStrategy();
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load portfolios';
    }
  }

  async function loadStrategy() {
    if (!selectedId) return;
    try {
      loading = true;
      error = '';
      overview = await getStrategyOverview(selectedId);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load strategy';
      overview = null;
    } finally {
      loading = false;
    }
  }

  function plColor(value: number): string {
    return value >= 0 ? '#3fb950' : '#f85149';
  }

  function formatLargeCents(cents: number): string {
    const val = Math.abs(cents) / 100;
    const sign = cents < 0 ? '-' : '';
    if (val >= 1_000_000) return `${sign}$${(val / 1_000_000).toFixed(2)}M`;
    if (val >= 1_000) return `${sign}$${(val / 1_000).toFixed(1)}K`;
    return formatCents(cents);
  }

  function bucketColor(bucket: string): string {
    const colors: Record<string, string> = {
      'ETF': '#58a6ff',
      'Stocks': '#3fb950',
      'Trading': '#f0883e',
      'Untagged': '#484f58',
    };
    return colors[bucket] ?? '#8b949e';
  }

  init();
</script>

<div class="strategy-page">
  <div class="header">
    <h2>Strategy Spread</h2>
    {#if portfolios.length > 1}
      <select
        value={selectedId}
        onchange={(e) => { selectedId = Number((e.target as HTMLSelectElement).value); loadStrategy(); }}
      >
        {#each portfolios as p}
          <option value={p.id}>{p.name}</option>
        {/each}
      </select>
    {/if}
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  {#if loading}
    <p class="loading-text">Loading strategy overview...</p>
  {:else if overview}
    <!-- Totals -->
    <div class="totals">
      <div class="card">
        <div class="card-label">Total Invested</div>
        <div class="card-value">{formatLargeCents(overview.total_invested_cents)}</div>
      </div>
      <div class="card">
        <div class="card-label">Current Value</div>
        <div class="card-value">{formatLargeCents(overview.total_current_cents)}</div>
      </div>
      <div class="card">
        <div class="card-label">Total P/L</div>
        <div class="card-value" style="color: {plColor(overview.total_pl_cents)}">
          {formatLargeCents(overview.total_pl_cents)} ({formatPercent(overview.total_pl_percent)})
        </div>
      </div>
    </div>

    <!-- Spread Bar -->
    {#if overview.buckets.length > 0}
      <div class="section">
        <h3>Spread</h3>
        <div class="spread-bar">
          {#each overview.buckets as bucket}
            <div
              class="spread-segment"
              style="width: {Math.max(bucket.actual_percent * 100, 2)}%; background: {bucketColor(bucket.bucket)}"
              title="{bucket.bucket}: {(bucket.actual_percent * 100).toFixed(1)}%"
            >
              {#if bucket.actual_percent > 0.08}
                <span class="spread-label">{bucket.bucket} {(bucket.actual_percent * 100).toFixed(0)}%</span>
              {/if}
            </div>
          {/each}
        </div>
        <div class="spread-legend">
          {#each overview.buckets as bucket}
            <div class="legend-item">
              <span class="legend-dot" style="background: {bucketColor(bucket.bucket)}"></span>
              <span>{bucket.bucket}: {(bucket.actual_percent * 100).toFixed(1)}%</span>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Bucket Cards -->
    <div class="section">
      <h3>Performance by Bucket</h3>
      {#if overview.buckets.length === 0}
        <p class="empty-text">No holdings yet. Add trades with a bucket tag (ETF, Stocks, Trading) to see your strategy spread.</p>
      {:else}
        <div class="bucket-grid">
          {#each overview.buckets as bucket}
            <div class="bucket-card">
              <div class="bucket-header">
                <span class="bucket-name" style="color: {bucketColor(bucket.bucket)}">{bucket.bucket}</span>
                <span class="bucket-pct" style="color: {plColor(bucket.pl_cents)}">{formatPercent(bucket.pl_percent)}</span>
              </div>
              <div class="bucket-stats">
                <div class="stat-row">
                  <span class="stat-label">Invested</span>
                  <span class="stat-value">{formatLargeCents(bucket.invested_cents)}</span>
                </div>
                <div class="stat-row">
                  <span class="stat-label">Current</span>
                  <span class="stat-value">{formatLargeCents(bucket.current_cents)}</span>
                </div>
                <div class="stat-row">
                  <span class="stat-label">P/L</span>
                  <span class="stat-value" style="color: {plColor(bucket.pl_cents)}">
                    {formatLargeCents(bucket.pl_cents)}
                  </span>
                </div>
                <div class="stat-row">
                  <span class="stat-label">Share of Portfolio</span>
                  <span class="stat-value">{(bucket.actual_percent * 100).toFixed(1)}%</span>
                </div>
                <div class="stat-row">
                  <span class="stat-label">Holdings</span>
                  <span class="stat-value">{bucket.holding_count}</span>
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Comparison Table -->
    {#if overview.buckets.length > 1}
      <div class="section">
        <h3>Comparison</h3>
        <table class="compare-table">
          <thead>
            <tr>
              <th>Bucket</th>
              <th class="num">Invested</th>
              <th class="num">Current</th>
              <th class="num">P/L</th>
              <th class="num">Return</th>
              <th class="num">Share</th>
            </tr>
          </thead>
          <tbody>
            {#each overview.buckets as bucket}
              <tr>
                <td style="color: {bucketColor(bucket.bucket)}; font-weight: 600">{bucket.bucket}</td>
                <td class="num">{formatLargeCents(bucket.invested_cents)}</td>
                <td class="num">{formatLargeCents(bucket.current_cents)}</td>
                <td class="num" style="color: {plColor(bucket.pl_cents)}">{formatLargeCents(bucket.pl_cents)}</td>
                <td class="num" style="color: {plColor(bucket.pl_cents)}">{formatPercent(bucket.pl_percent)}</td>
                <td class="num">{(bucket.actual_percent * 100).toFixed(1)}%</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}

    <div class="hint">
      Tip: When adding trades, set the bucket to "ETF", "Stocks", or "Trading" to categorize them.
    </div>
  {:else}
    <p class="empty-text">Select a portfolio to see your strategy spread.</p>
  {/if}
</div>

<style>
  .strategy-page { max-width: 1000px; }
  .header { display: flex; align-items: center; gap: 1rem; margin-bottom: 1.5rem; }
  .header h2 { margin: 0; }
  .header select {
    padding: 0.4rem 0.75rem; background: #161b22; border: 1px solid #30363d;
    border-radius: 6px; color: #e1e4e8; font-size: 0.85rem;
  }

  .error {
    background: #3d1f1f; border: 1px solid #f85149; color: #f85149;
    padding: 0.5rem 1rem; border-radius: 6px; margin-bottom: 1rem; font-size: 0.9rem;
  }

  .totals { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; margin-bottom: 1.5rem; }
  .card {
    background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 1.25rem;
  }
  .card-label { font-size: 0.8rem; color: #8b949e; text-transform: uppercase; margin-bottom: 0.5rem; }
  .card-value { font-size: 1.4rem; font-weight: 600; color: #e1e4e8; }

  .section { margin-bottom: 1.5rem; }
  .section h3 { margin: 0 0 0.75rem 0; color: #e1e4e8; font-size: 1.1rem; }

  .spread-bar {
    display: flex; height: 32px; border-radius: 6px; overflow: hidden;
    background: #21262d; margin-bottom: 0.5rem;
  }
  .spread-segment {
    display: flex; align-items: center; justify-content: center;
    transition: width 0.3s ease; min-width: 4px;
  }
  .spread-label { font-size: 0.7rem; font-weight: 600; color: #fff; white-space: nowrap; }
  .spread-legend { display: flex; gap: 1rem; flex-wrap: wrap; }
  .legend-item { display: flex; align-items: center; gap: 0.35rem; font-size: 0.8rem; color: #8b949e; }
  .legend-dot { width: 10px; height: 10px; border-radius: 50%; }

  .bucket-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1rem; }
  .bucket-card {
    background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 1.25rem;
  }
  .bucket-header {
    display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem;
  }
  .bucket-name { font-size: 1.1rem; font-weight: 700; }
  .bucket-pct { font-size: 1.1rem; font-weight: 600; }
  .bucket-stats { display: flex; flex-direction: column; gap: 0.35rem; }
  .stat-row { display: flex; justify-content: space-between; font-size: 0.85rem; }
  .stat-label { color: #8b949e; }
  .stat-value { color: #e1e4e8; font-weight: 500; }

  .compare-table {
    width: 100%; border-collapse: collapse; background: #161b22;
    border: 1px solid #30363d; border-radius: 6px; overflow: hidden;
  }
  .compare-table th {
    padding: 0.6rem 1rem; text-align: left; color: #8b949e; font-size: 0.8rem;
    font-weight: 600; text-transform: uppercase; border-bottom: 1px solid #30363d;
  }
  .compare-table th.num, .compare-table td.num { text-align: right; }
  .compare-table td {
    padding: 0.6rem 1rem; border-bottom: 1px solid #21262d; font-size: 0.9rem;
  }
  .compare-table tr:last-child td { border-bottom: none; }
  .compare-table tr:hover { background: #1c2128; }

  .hint {
    color: #484f58; font-size: 0.8rem; font-style: italic;
    background: #161b22; border: 1px solid #30363d; border-radius: 6px;
    padding: 0.75rem 1rem;
  }

  .loading-text, .empty-text { color: #8b949e; text-align: center; padding: 3rem; }

  @media (max-width: 768px) {
    .totals { grid-template-columns: 1fr; }
    .bucket-grid { grid-template-columns: 1fr; }
  }
</style>
