<script lang="ts">
  import { Chart, registerables } from 'chart.js';
  import type { PortfolioSnapshot } from '$lib/api/dashboard';
  import { formatCents } from '$lib/utils/format';

  Chart.register(...registerables);

  let { snapshots, currency = 'EUR' }: { snapshots: PortfolioSnapshot[]; currency?: string } = $props();

  let canvas: HTMLCanvasElement;
  let chart: Chart | null = null;

  $effect(() => {
    if (!canvas || snapshots.length === 0) return;

    if (chart) {
      chart.destroy();
    }

    const labels = snapshots.map((s) => s.date);
    const values = snapshots.map((s) => s.value_cents / 100);
    const costs = snapshots.map((s) => s.cost_cents / 100);

    chart = new Chart(canvas, {
      type: 'line',
      data: {
        labels,
        datasets: [
          {
            label: 'Market Value',
            data: values,
            borderColor: '#58a6ff',
            backgroundColor: 'rgba(88, 166, 255, 0.1)',
            fill: true,
            tension: 0.3,
            pointRadius: 0,
          },
          {
            label: 'Cost Basis',
            data: costs,
            borderColor: '#8b949e',
            borderDash: [5, 5],
            fill: false,
            tension: 0.3,
            pointRadius: 0,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            labels: { color: '#8b949e' },
          },
          tooltip: {
            callbacks: {
              label: (ctx) => `${ctx.dataset.label}: ${formatCents(Math.round(ctx.parsed.y * 100), currency)}`,
            },
          },
        },
        scales: {
          x: {
            ticks: { color: '#8b949e', maxTicksLimit: 10 },
            grid: { color: '#21262d' },
          },
          y: {
            ticks: {
              color: '#8b949e',
              callback: (v) => formatCents(Math.round(Number(v) * 100), currency),
            },
            grid: { color: '#21262d' },
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
