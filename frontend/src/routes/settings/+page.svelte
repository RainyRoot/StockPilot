<script lang="ts">
  import { getSettings, updateSettings, type UserSettings } from '$lib/api/settings';
  import { getPortfolios, type Portfolio } from '$lib/api/portfolios';
  import { downloadHoldingsCSV, downloadTradesCSV } from '$lib/api/export';
  import { toast } from '$lib/components/Toast.svelte';
  import Skeleton from '$lib/components/Skeleton.svelte';

  let settings = $state<UserSettings | null>(null);
  let portfolios = $state<Portfolio[]>([]);
  let loading = $state(true);
  let saving = $state(false);
  let error = $state('');
  let exportPortfolioId = $state<number | null>(null);

  // Form state
  let currency = $state('USD');
  let discountRate = $state(10);
  let growthRate = $state(5);
  let terminalGrowth = $state(2.5);

  async function load() {
    try {
      loading = true;
      const [s, p] = await Promise.all([getSettings(), getPortfolios()]);
      settings = s;
      portfolios = p;
      currency = s.currency;
      discountRate = s.discount_rate * 100;
      growthRate = s.growth_rate * 100;
      terminalGrowth = s.terminal_growth * 100;
      if (p.length > 0) exportPortfolioId = p[0].id;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load settings';
    } finally {
      loading = false;
    }
  }

  async function save() {
    try {
      saving = true;
      await updateSettings({
        currency,
        discount_rate: discountRate / 100,
        growth_rate: growthRate / 100,
        terminal_growth: terminalGrowth / 100,
        theme: 'dark',
      });
      toast('success', 'Settings saved successfully');
    } catch (e) {
      toast('error', e instanceof Error ? e.message : 'Failed to save settings');
    } finally {
      saving = false;
    }
  }

  async function exportHoldings() {
    if (!exportPortfolioId) return;
    try {
      await downloadHoldingsCSV(exportPortfolioId);
      toast('success', 'Holdings CSV downloaded');
    } catch {
      toast('error', 'Failed to export holdings');
    }
  }

  async function exportTrades() {
    if (!exportPortfolioId) return;
    try {
      await downloadTradesCSV(exportPortfolioId);
      toast('success', 'Trades CSV downloaded');
    } catch {
      toast('error', 'Failed to export trades');
    }
  }

  $effect(() => {
    load();
  });
</script>

<div class="settings-page">
  <h2>Settings</h2>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  {#if loading}
    <Skeleton height="20rem" />
  {:else}
    <div class="settings-grid">
      <!-- General Settings -->
      <section class="settings-section">
        <h3>General</h3>
        <div class="field">
          <label for="currency">Default Currency</label>
          <select id="currency" bind:value={currency}>
            <option value="USD">USD ($)</option>
            <option value="EUR">EUR (&#8364;)</option>
          </select>
        </div>
      </section>

      <!-- Valuation Parameters -->
      <section class="settings-section">
        <h3>Valuation Parameters</h3>
        <div class="field">
          <label for="discount">Discount Rate (%)</label>
          <input id="discount" type="number" min="1" max="30" step="0.5" bind:value={discountRate} />
          <span class="hint">Used in DCF calculations (1-30%)</span>
        </div>
        <div class="field">
          <label for="growth">Growth Rate (%)</label>
          <input id="growth" type="number" min="0" max="30" step="0.5" bind:value={growthRate} />
          <span class="hint">FCF growth assumption (0-30%)</span>
        </div>
        <div class="field">
          <label for="terminal">Terminal Growth Rate (%)</label>
          <input id="terminal" type="number" min="0" max="10" step="0.25" bind:value={terminalGrowth} />
          <span class="hint">Long-term growth rate (0-10%)</span>
        </div>
      </section>

      <!-- Export -->
      <section class="settings-section">
        <h3>Export Data</h3>
        {#if portfolios.length > 0}
          <div class="field">
            <label for="export-portfolio">Portfolio</label>
            <select id="export-portfolio" bind:value={exportPortfolioId}>
              {#each portfolios as p}
                <option value={p.id}>{p.name}</option>
              {/each}
            </select>
          </div>
          <div class="export-buttons">
            <button class="btn btn-secondary" onclick={exportHoldings}>Export Holdings CSV</button>
            <button class="btn btn-secondary" onclick={exportTrades}>Export Trades CSV</button>
          </div>
        {:else}
          <p class="empty-text">No portfolios to export</p>
        {/if}
      </section>
    </div>

    <div class="save-row">
      <button class="btn btn-primary" onclick={save} disabled={saving}>
        {saving ? 'Saving...' : 'Save Settings'}
      </button>
    </div>
  {/if}
</div>

<style>
  .settings-page { max-width: 700px; }
  .settings-page h2 { margin: 0 0 1.5rem; }
  .settings-page h3 { margin: 0 0 1rem; font-size: 1rem; color: #58a6ff; }

  .error {
    background: #3d1f1f; border: 1px solid #f85149; color: #f85149;
    padding: 0.5rem 1rem; border-radius: 6px; margin-bottom: 1rem;
  }

  .settings-grid {
    display: flex; flex-direction: column; gap: 1.5rem;
  }

  .settings-section {
    padding: 1.25rem; background: #161b22;
    border: 1px solid #30363d; border-radius: 6px;
  }

  .field {
    display: flex; flex-direction: column; gap: 0.25rem; margin-bottom: 1rem;
  }
  .field:last-child { margin-bottom: 0; }

  label {
    font-size: 0.85rem; color: #8b949e; font-weight: 500;
  }

  input, select {
    padding: 0.5rem 0.75rem; background: #0d1117; border: 1px solid #30363d;
    border-radius: 4px; color: #e1e4e8; font-size: 0.9rem;
    font-family: inherit;
  }
  input:focus, select:focus {
    outline: none; border-color: #58a6ff;
  }

  .hint { font-size: 0.75rem; color: #484f58; }

  .export-buttons {
    display: flex; gap: 0.75rem; flex-wrap: wrap;
  }

  .btn {
    padding: 0.5rem 1.25rem; border: none; border-radius: 6px;
    font-size: 0.875rem; cursor: pointer; font-family: inherit;
    font-weight: 500; transition: opacity 0.15s;
  }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .btn-primary { background: #238636; color: #fff; }
  .btn-primary:hover:not(:disabled) { background: #2ea043; }
  .btn-secondary { background: #21262d; color: #c9d1d9; border: 1px solid #30363d; }
  .btn-secondary:hover { background: #30363d; }

  .save-row { margin-top: 1.5rem; }

  .empty-text { color: #8b949e; font-size: 0.875rem; }

  @media (max-width: 768px) {
    .export-buttons { flex-direction: column; }
    input, select { width: 100%; }
  }
</style>
