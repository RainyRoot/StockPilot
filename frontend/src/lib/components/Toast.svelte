<script lang="ts" module>
  interface ToastMessage {
    id: number;
    type: 'success' | 'error' | 'warning' | 'info';
    message: string;
  }

  let toasts = $state<ToastMessage[]>([]);
  let nextId = 0;

  export function toast(type: ToastMessage['type'], message: string) {
    const id = nextId++;
    toasts = [...toasts, { id, type, message }];
    setTimeout(() => {
      toasts = toasts.filter(t => t.id !== id);
    }, 4000);
  }
</script>

{#if toasts.length > 0}
  <div class="toast-container">
    {#each toasts as t (t.id)}
      <div class="toast toast-{t.type}">
        <span class="toast-icon">
          {#if t.type === 'success'}&#10003;
          {:else if t.type === 'error'}&#10007;
          {:else if t.type === 'warning'}&#9888;
          {:else}&#8505;
          {/if}
        </span>
        <span class="toast-message">{t.message}</span>
        <button class="toast-close" onclick={() => { toasts = toasts.filter(x => x.id !== t.id); }}>&times;</button>
      </div>
    {/each}
  </div>
{/if}

<style>
  .toast-container {
    position: fixed;
    top: 1rem;
    right: 1rem;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-width: 380px;
  }

  .toast {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border-radius: 6px;
    font-size: 0.875rem;
    animation: slideIn 0.3s ease;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  }

  @keyframes slideIn {
    from { transform: translateX(100%); opacity: 0; }
    to { transform: translateX(0); opacity: 1; }
  }

  .toast-success { background: #1a3a2a; border: 1px solid #3fb950; color: #3fb950; }
  .toast-error { background: #3d1f1f; border: 1px solid #f85149; color: #f85149; }
  .toast-warning { background: #3d2e1f; border: 1px solid #d29922; color: #d29922; }
  .toast-info { background: #1f2d3d; border: 1px solid #58a6ff; color: #58a6ff; }

  .toast-icon { font-size: 1rem; flex-shrink: 0; }
  .toast-message { flex: 1; color: #e1e4e8; }
  .toast-close {
    background: none; border: none; color: #8b949e;
    cursor: pointer; font-size: 1.2rem; padding: 0;
    line-height: 1;
  }
  .toast-close:hover { color: #e1e4e8; }
</style>
