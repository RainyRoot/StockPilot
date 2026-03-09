<script lang="ts">
  import { page } from '$app/stores';
</script>

<div class="error-page">
  <div class="error-card">
    <h1 class="error-code">{$page.status}</h1>
    <p class="error-message">{$page.error?.message ?? 'Something went wrong'}</p>
    <a href="/" class="btn">Go to Dashboard</a>
  </div>
</div>

<style>
  .error-page {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
  }

  .error-card {
    text-align: center;
    padding: 3rem;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 8px;
    max-width: 400px;
  }

  .error-code {
    font-size: 4rem;
    font-weight: 800;
    color: #f85149;
    margin: 0 0 0.5rem;
  }

  .error-message {
    color: #8b949e;
    font-size: 1.1rem;
    margin: 0 0 2rem;
  }

  .btn {
    display: inline-block;
    padding: 0.6rem 1.5rem;
    background: #238636;
    color: #fff;
    text-decoration: none;
    border-radius: 6px;
    font-weight: 500;
    font-size: 0.9rem;
    transition: background 0.15s;
  }

  .btn:hover {
    background: #2ea043;
  }
</style>
