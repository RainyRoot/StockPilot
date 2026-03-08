<script lang="ts">
  import type { Snippet } from 'svelte';

  let { children }: { children: Snippet } = $props();
</script>

<div class="app">
  <nav class="sidebar">
    <div class="logo">
      <h1>StockPilot</h1>
    </div>
    <ul>
      <li><a href="/">Dashboard</a></li>
      <li><a href="/watchlist">Watchlist</a></li>
      <li><a href="/portfolio">Portfolio</a></li>
      <li><a href="/valuation">Valuation</a></li>
      <li><a href="/allocation">Allocation</a></li>
    </ul>
  </nav>
  <main class="content">
    {@render children()}
  </main>
</div>

<style>
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #0f1117;
    color: #e1e4e8;
  }

  .app {
    display: flex;
    min-height: 100vh;
  }

  .sidebar {
    width: 220px;
    background: #161b22;
    border-right: 1px solid #30363d;
    padding: 1rem 0;
    flex-shrink: 0;
  }

  .logo {
    padding: 0 1.5rem 1rem;
    border-bottom: 1px solid #30363d;
  }

  .logo h1 {
    font-size: 1.25rem;
    color: #58a6ff;
    margin: 0;
  }

  .sidebar ul {
    list-style: none;
    padding: 0.5rem 0;
    margin: 0;
  }

  .sidebar li a {
    display: block;
    padding: 0.6rem 1.5rem;
    color: #8b949e;
    text-decoration: none;
    font-size: 0.9rem;
    transition: all 0.15s;
  }

  .sidebar li a:hover {
    color: #e1e4e8;
    background: #1c2128;
  }

  .content {
    flex: 1;
    padding: 2rem;
    overflow-y: auto;
  }
</style>
