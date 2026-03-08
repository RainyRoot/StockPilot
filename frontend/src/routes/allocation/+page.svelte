<script lang="ts">
  import { getPortfolios, type Portfolio } from '$lib/api/portfolios';
  import {
    getTargets, setTargets, getActual, getComparison, getRebalance,
    type AllocationTarget, type AllocationActual, type AllocationComparison, type RebalanceAction
  } from '$lib/api/allocation';
  import { formatCents } from '$lib/utils/format';

  let portfolios = $state<Portfolio[]>([]);
  let selectedPortfolioId = $state<number | null>(null);
  let targets = $state<AllocationTarget[]>([]);
  let actual = $state<AllocationActual[]>([]);
  let comparison = $state<AllocationComparison[]>([]);
  let rebalanceActions = $state<RebalanceAction[]>([]);
  let loading = $state(false);
  let error = $state('');
  let editTargets = $state<{ category: string; target_percent: number }[]>([]);
  let editMode = $state(false);
  let saving = $state(false);

  const SECTOR_PRESETS = [
    'Technology', 'Healthcare', 'Financial Services', 'Consumer Cyclical',
    'Communication Services', 'Industrials', 'Consumer Defensive',
    'Energy', 'Utilities', 'Real Estate', 'Basic Materials', 'Other'
  ];

  const COLORS = [
    '#58a6ff', '#3fb950', '#d29922', '#f85149', '#bc8cff',
    '#79c0ff', '#56d364', '#e3b341', '#ff7b72', '#d2a8ff',
    '#39d353', '#a5d6ff'
  ];

  let allCategories = $derived(() => {
    const cats = new Set<string>();
    for (const t of targets) cats.add(t.category);
    for (const a of actual) cats.add(a.category);
    for (const c of comparison) cats.add(c.category);
    return [...cats].sort();
  });

  function categoryColor(category: string): string {
    const cats = allCategories();
    const idx = cats.indexOf(category);
    return COLORS[idx >= 0 ? idx % COLORS.length : 0];
  }

  let targetSum = $derived(() => {
    return editTargets.reduce((s, t) => s + t.target_percent, 0);
  });

  const circumference = 2 * Math.PI * 60;

  function pieSegments(data: { category: string; percent: number }[]) {
    let offset = 0;
    return data.map(d => {
      const dashLength = (d.percent / 100) * circumference;
      const seg = { ...d, dashLength, offset: -offset, color: categoryColor(d.category) };
      offset += dashLength;
      return seg;
    });
  }

  let targetSegments = $derived(() => {
    return pieSegments(targets.map(t => ({ category: t.category, percent: t.target_percent })));
  });

  let actualSegments = $derived(() => {
    return pieSegments(actual.map(a => ({ category: a.category, percent: a.actual_percent })));
  });

  async function loadPortfolios() {
    try {
      portfolios = await getPortfolios();
      if (portfolios.length > 0 && selectedPortfolioId === null) {
        selectedPortfolioId = portfolios[0].id;
      }
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : 'Failed to load portfolios';
    }
  }

  async function loadAllocationData() {
    if (!selectedPortfolioId) return;
    loading = true;
    error = '';
    try {
      const [t, a, c, r] = await Promise.all([
        getTargets(selectedPortfolioId),
        getActual(selectedPortfolioId),
        getComparison(selectedPortfolioId),
        getRebalance(selectedPortfolioId),
      ]);
      targets = t;
      actual = a;
      comparison = c;
      rebalanceActions = r;
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : 'Failed to load allocation data';
    } finally {
      loading = false;
    }
  }

  function startEdit() {
    if (targets.length > 0) {
      editTargets = targets.map(t => ({ category: t.category, target_percent: t.target_percent }));
    } else {
      editTargets = [{ category: 'Technology', target_percent: 0 }];
    }
    editMode = true;
  }

  function addCategory() {
    const used = new Set(editTargets.map(t => t.category));
    const next = SECTOR_PRESETS.find(s => !used.has(s)) || '';
    editTargets = [...editTargets, { category: next, target_percent: 0 }];
  }

  function removeCategory(idx: number) {
    editTargets = editTargets.filter((_, i) => i !== idx);
  }

  async function saveTargets() {
    if (!selectedPortfolioId) return;
    saving = true;
    error = '';
    try {
      await setTargets(selectedPortfolioId, editTargets);
      editMode = false;
      await loadAllocationData();
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : 'Failed to save targets';
    } finally {
      saving = false;
    }
  }

  function diffClass(pct: number): string {
    const abs = Math.abs(pct);
    if (abs <= 2) return 'diff-ok';
    if (abs <= 5) return 'diff-warn';
    return 'diff-danger';
  }

  $effect(() => {
    loadPortfolios();
  });

  $effect(() => {
    if (selectedPortfolioId) {
      loadAllocationData();
    }
  });
</script>

<div class="allocation-page">
  <h1>Allocation & Rebalancing</h1>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  <div class="controls">
    <label for="portfolio-select">Portfolio</label>
    <select id="portfolio-select" bind:value={selectedPortfolioId} onchange={() => { editMode = false; }}>
      {#each portfolios as p}
        <option value={p.id}>{p.name} ({p.currency})</option>
      {/each}
    </select>
  </div>

  {#if loading}
    <div class="loading">Loading allocation data...</div>
  {:else if selectedPortfolioId}

    <!-- Target Editor -->
    <section class="section">
      <div class="section-header">
        <h2>Target Allocation</h2>
        {#if !editMode}
          <button class="btn btn-primary" onclick={startEdit}>Edit Targets</button>
        {/if}
      </div>

      {#if editMode}
        <div class="target-editor">
          {#each editTargets as target, idx}
            <div class="target-row">
              <select bind:value={target.category}>
                {#each SECTOR_PRESETS as preset}
                  <option value={preset}>{preset}</option>
                {/each}
              </select>
              <div class="pct-input">
                <input type="number" bind:value={target.target_percent} min="0" max="100" step="0.5" />
                <span>%</span>
              </div>
              <button class="btn btn-danger btn-sm" onclick={() => removeCategory(idx)}>Remove</button>
            </div>
          {/each}
          <div class="target-actions">
            <button class="btn btn-secondary" onclick={addCategory}>+ Add Category</button>
            <span class="sum-display" class:sum-valid={Math.abs(targetSum() - 100) < 0.01} class:sum-invalid={Math.abs(targetSum() - 100) >= 0.01}>
              Sum: {targetSum().toFixed(1)}%
            </span>
            <div class="action-buttons">
              <button class="btn btn-secondary" onclick={() => { editMode = false; }}>Cancel</button>
              <button class="btn btn-primary" onclick={saveTargets} disabled={saving || Math.abs(targetSum() - 100) >= 0.01}>
                {saving ? 'Saving...' : 'Save Targets'}
              </button>
            </div>
          </div>
        </div>
      {:else if targets.length === 0}
        <p class="muted">No target allocation set. Click "Edit Targets" to define your allocation goals.</p>
      {/if}
    </section>

    <!-- Pie Charts -->
    {#if targets.length > 0 || actual.length > 0}
      <section class="section">
        <h2>Allocation Overview</h2>
        <div class="charts-grid">
          {#if targets.length > 0}
            <div class="chart-card">
              <h3>Target</h3>
              <svg viewBox="0 0 160 160" class="pie-chart">
                {#each targetSegments() as seg}
                  <circle
                    cx="80" cy="80" r="60"
                    fill="none"
                    stroke={seg.color}
                    stroke-width="40"
                    stroke-dasharray="{seg.dashLength} {circumference - seg.dashLength}"
                    stroke-dashoffset={seg.offset}
                    transform="rotate(-90 80 80)"
                  />
                {/each}
              </svg>
              <div class="legend">
                {#each targets as t}
                  <div class="legend-item">
                    <span class="legend-dot" style="background:{categoryColor(t.category)}"></span>
                    <span class="legend-label">{t.category}</span>
                    <span class="legend-value">{t.target_percent.toFixed(1)}%</span>
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          {#if actual.length > 0}
            <div class="chart-card">
              <h3>Actual</h3>
              <svg viewBox="0 0 160 160" class="pie-chart">
                {#each actualSegments() as seg}
                  <circle
                    cx="80" cy="80" r="60"
                    fill="none"
                    stroke={seg.color}
                    stroke-width="40"
                    stroke-dasharray="{seg.dashLength} {circumference - seg.dashLength}"
                    stroke-dashoffset={seg.offset}
                    transform="rotate(-90 80 80)"
                  />
                {/each}
              </svg>
              <div class="legend">
                {#each actual as a}
                  <div class="legend-item">
                    <span class="legend-dot" style="background:{categoryColor(a.category)}"></span>
                    <span class="legend-label">{a.category}</span>
                    <span class="legend-value">{a.actual_percent.toFixed(1)}% ({formatCents(a.value_cents)})</span>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </section>
    {/if}

    <!-- Comparison Table -->
    {#if comparison.length > 0}
      <section class="section">
        <h2>Target vs Actual</h2>
        <table class="data-table">
          <thead>
            <tr>
              <th>Category</th>
              <th>Target %</th>
              <th>Actual %</th>
              <th>Difference %</th>
              <th>Difference $</th>
            </tr>
          </thead>
          <tbody>
            {#each comparison as c}
              <tr>
                <td>
                  <span class="legend-dot" style="background:{categoryColor(c.category)}"></span>
                  {c.category}
                </td>
                <td>{c.target_percent.toFixed(1)}%</td>
                <td>{c.actual_percent.toFixed(1)}%</td>
                <td class={diffClass(c.diff_percent)}>
                  {c.diff_percent > 0 ? '+' : ''}{c.diff_percent.toFixed(1)}%
                </td>
                <td class={diffClass(c.diff_percent)}>
                  {c.diff_cents > 0 ? '+' : ''}{formatCents(c.diff_cents)}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </section>
    {/if}

    <!-- Rebalancing Suggestions -->
    {#if rebalanceActions.length > 0}
      <section class="section">
        <h2>Rebalancing Suggestions</h2>
        <table class="data-table">
          <thead>
            <tr>
              <th>Ticker</th>
              <th>Action</th>
              <th>Quantity</th>
              <th>Value</th>
              <th>Reason</th>
            </tr>
          </thead>
          <tbody>
            {#each rebalanceActions as ra}
              <tr>
                <td class="ticker">{ra.ticker}</td>
                <td>
                  <span class="badge" class:badge-buy={ra.action === 'BUY'} class:badge-sell={ra.action === 'SELL'}>
                    {ra.action}
                  </span>
                </td>
                <td>{ra.quantity}</td>
                <td>{formatCents(ra.value_cents)}</td>
                <td class="reason">{ra.reason}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </section>
    {:else if targets.length > 0 && actual.length > 0}
      <section class="section">
        <h2>Rebalancing Suggestions</h2>
        <p class="muted">Your portfolio is well balanced. No rebalancing needed.</p>
      </section>
    {/if}
  {/if}
</div>

<style>
  .allocation-page {
    max-width: 1100px;
    margin: 0 auto;
  }

  h1 {
    font-size: 1.5rem;
    margin-bottom: 1.5rem;
    color: #e1e4e8;
  }

  h2 {
    font-size: 1.15rem;
    color: #e1e4e8;
    margin: 0;
  }

  h3 {
    font-size: 1rem;
    color: #8b949e;
    margin: 0 0 0.75rem;
    text-align: center;
  }

  .error-banner {
    background: #f8514922;
    border: 1px solid #f85149;
    color: #f85149;
    padding: 0.75rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
  }

  .controls {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 1.5rem;
  }

  .controls label {
    color: #8b949e;
    font-size: 0.9rem;
  }

  .controls select {
    background: #161b22;
    border: 1px solid #30363d;
    color: #e1e4e8;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    font-size: 0.9rem;
  }

  .loading {
    color: #8b949e;
    padding: 2rem;
    text-align: center;
  }

  .muted {
    color: #8b949e;
    font-size: 0.9rem;
  }

  .section {
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 8px;
    padding: 1.25rem;
    margin-bottom: 1.5rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  /* Buttons */
  .btn {
    padding: 0.45rem 1rem;
    border: 1px solid #30363d;
    border-radius: 6px;
    font-size: 0.85rem;
    cursor: pointer;
    transition: all 0.15s;
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .btn-primary {
    background: #238636;
    border-color: #238636;
    color: #fff;
  }
  .btn-primary:hover:not(:disabled) {
    background: #2ea043;
  }
  .btn-secondary {
    background: #21262d;
    color: #c9d1d9;
  }
  .btn-secondary:hover {
    background: #30363d;
  }
  .btn-danger {
    background: transparent;
    border-color: #f85149;
    color: #f85149;
  }
  .btn-danger:hover {
    background: #f8514922;
  }
  .btn-sm {
    padding: 0.25rem 0.5rem;
    font-size: 0.8rem;
  }

  /* Target Editor */
  .target-editor {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .target-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .target-row select {
    flex: 1;
    background: #0d1117;
    border: 1px solid #30363d;
    color: #e1e4e8;
    padding: 0.4rem 0.5rem;
    border-radius: 6px;
    font-size: 0.85rem;
  }
  .pct-input {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }
  .pct-input input {
    width: 70px;
    background: #0d1117;
    border: 1px solid #30363d;
    color: #e1e4e8;
    padding: 0.4rem 0.5rem;
    border-radius: 6px;
    font-size: 0.85rem;
    text-align: right;
  }
  .pct-input span {
    color: #8b949e;
    font-size: 0.85rem;
  }
  .target-actions {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-top: 0.75rem;
    flex-wrap: wrap;
  }
  .sum-display {
    font-size: 0.9rem;
    font-weight: 600;
  }
  .sum-valid {
    color: #3fb950;
  }
  .sum-invalid {
    color: #f85149;
  }
  .action-buttons {
    display: flex;
    gap: 0.5rem;
    margin-left: auto;
  }

  /* Charts */
  .charts-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
    margin-top: 1rem;
  }
  .chart-card {
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .pie-chart {
    width: 180px;
    height: 180px;
    margin-bottom: 1rem;
  }
  .legend {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    width: 100%;
  }
  .legend-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.82rem;
    color: #c9d1d9;
  }
  .legend-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .legend-label {
    flex: 1;
  }
  .legend-value {
    color: #8b949e;
  }

  /* Tables */
  .data-table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 1rem;
    font-size: 0.85rem;
  }
  .data-table th {
    text-align: left;
    padding: 0.6rem 0.75rem;
    color: #8b949e;
    font-weight: 600;
    border-bottom: 1px solid #30363d;
  }
  .data-table td {
    padding: 0.6rem 0.75rem;
    color: #c9d1d9;
    border-bottom: 1px solid #21262d;
  }
  .data-table tbody tr:hover {
    background: #1c2128;
  }
  .ticker {
    font-weight: 600;
    color: #58a6ff;
  }
  .reason {
    color: #8b949e;
    font-size: 0.82rem;
  }

  .diff-ok {
    color: #3fb950;
  }
  .diff-warn {
    color: #d29922;
  }
  .diff-danger {
    color: #f85149;
  }

  /* Badges */
  .badge {
    display: inline-block;
    padding: 0.15rem 0.5rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
  }
  .badge-buy {
    background: #238636;
    color: #fff;
  }
  .badge-sell {
    background: #da3633;
    color: #fff;
  }
</style>
