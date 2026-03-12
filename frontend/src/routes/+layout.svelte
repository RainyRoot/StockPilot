<script lang="ts">
  import type { Snippet } from 'svelte';
  import Toast from '$lib/components/Toast.svelte';

  let { children }: { children: Snippet } = $props();
  let menuOpen = $state(false);
</script>

<div class="app">
  <nav class="sidebar" class:open={menuOpen}>
    <div class="logo">
      <h1>StockPilot</h1>
      <button class="hamburger" onclick={() => menuOpen = !menuOpen} aria-label="Toggle menu">
        {#if menuOpen}&#10005;{:else}&#9776;{/if}
      </button>
    </div>
    <ul>
      <li><a href="/" onclick={() => menuOpen = false}>Dashboard</a></li>
      <li><a href="/watchlist" onclick={() => menuOpen = false}>Watchlist</a></li>
      <li><a href="/portfolio" onclick={() => menuOpen = false}>Portfolio</a></li>
      <li><a href="/valuation" onclick={() => menuOpen = false}>Valuation</a></li>
      <li><a href="/strategy" onclick={() => menuOpen = false}>Strategy</a></li>
      <li><a href="/allocation" onclick={() => menuOpen = false}>Allocation</a></li>
      <li><a href="/settings" onclick={() => menuOpen = false}>Settings</a></li>
    </ul>
  </nav>
  <main class="content">
    {@render children()}
  </main>
</div>

<Toast />

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
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .logo h1 {
    font-size: 1.25rem;
    color: #58a6ff;
    margin: 0;
  }

  .hamburger {
    display: none;
    background: none;
    border: none;
    color: #e1e4e8;
    font-size: 1.4rem;
    cursor: pointer;
    padding: 0;
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

  @media (max-width: 768px) {
    .hamburger {
      display: block;
    }

    .sidebar {
      position: fixed;
      top: 0;
      left: 0;
      bottom: 0;
      z-index: 100;
      transform: translateX(-100%);
      transition: transform 0.2s ease;
    }

    .sidebar.open {
      transform: translateX(0);
    }

    .logo {
      border-bottom: 1px solid #30363d;
    }

    .content {
      padding: 1rem;
      width: 100%;
    }

    .app {
      flex-direction: column;
    }
  }
</style>
