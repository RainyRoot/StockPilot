<script lang="ts">
  let { width = '100%', height = '1rem', borderRadius = '4px' }:
    { width?: string; height?: string; borderRadius?: string } = $props();
</script>

<div class="skeleton" style="width:{width};height:{height};border-radius:{borderRadius}"></div>

<style>
  .skeleton {
    background: linear-gradient(90deg, #21262d 25%, #30363d 50%, #21262d 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
  }
  @keyframes shimmer {
    0% { background-position: 200% 0; }
    100% { background-position: -200% 0; }
  }
</style>
