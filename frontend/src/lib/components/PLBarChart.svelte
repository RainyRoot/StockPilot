<script lang="ts">
  import { Chart, registerables } from 'chart.js';
  import type { HoldingPL } from '$lib/api/dashboard';
  import { formatCents } from '$lib/utils/format';

  Chart.register(...registerables);

  let { holdings, currency = 'EUR' }: { holdings: HoldingPL[]; currency?: string } = $props();

  let canvas: HTMLCanvasElement;
  let chart: Chart | null = null;

  $effect(() => {
    if (!canvas || holdings.length === 0) return;

    if (chart) {
      chart.destroy();
    }

    const sorted = [...holdings].sort((a, b) => b.unrealized_pl_cents - a.unrealized_pl_cents);
    const labels = sorted.map((h) => h.ticker);
    const values = sorted.map((h) => h.unrealized_pl_cents / 100);
    const colors = sorted.map((h) => (h.unrealized_pl_cents >= 0 ? '#3fb950' : '#f85149'));

    chart = new Chart(canvas, {
      type: 'bar',
      data: {
        labels,
        datasets: [
          {
            label: 'Unrealized P&L',
            data: values,
            backgroundColor: colors,
            borderRadius: 4,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        indexAxis: 'y',
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (ctx) => formatCents(Math.round(ctx.parsed.x * 100), currency),
            },
          },
        },
        scales: {
          x: {
            ticks: {
              color: '#8b949e',
              callback: (v) => formatCents(Math.round(Number(v) * 100), currency),
            },
            grid: { color: '#21262d' },
          },
          y: {
            ticks: { color: '#e1e4e8' },
            grid: { display: false },
          },
        },
      },
    });

    return () => {
      if (chart) {
        chart.destroy();
        chart = null;
      }
    };
  });
</script>

<div class="chart-container">
  <canvas bind:this={canvas}></canvas>
</div>

<style>
  .chart-container {
    height: 300px;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    padding: 1rem;
  }
</style>
