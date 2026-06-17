<script lang="ts">
  import { CheckCircleIcon, XCircleIcon, WarningCircleIcon, XIcon } from 'phosphor-svelte'
  import { toast } from '../../stores/toast'
</script>

{#if $toast.length > 0}
  <div class="fixed bottom-20 right-5 z-[200] flex flex-col gap-2 pointer-events-none">
    {#each $toast as t (t.id)}
      <div
        class="flex items-center gap-3 px-4 py-3 rounded-xl shadow-xl border backdrop-blur-sm pointer-events-auto max-w-xs text-sm font-medium
          animate-in slide-in-from-right-4 fade-in duration-200
          {t.type === 'success' ? 'bg-green-950/95 border-green-700/40 text-green-100' : ''}
          {t.type === 'error'   ? 'bg-red-950/95   border-red-700/40   text-red-100'   : ''}
          {t.type === 'warning' ? 'bg-amber-950/95 border-amber-700/40 text-amber-100' : ''}"
      >
        {#if t.type === 'success'}
          <CheckCircleIcon size={16} weight="fill" class="text-green-400 shrink-0" />
        {:else if t.type === 'error'}
          <XCircleIcon size={16} weight="fill" class="text-red-400 shrink-0" />
        {:else}
          <WarningCircleIcon size={16} weight="fill" class="text-amber-400 shrink-0" />
        {/if}
        <span class="flex-1">{t.message}</span>
        <button
          onclick={() => toast.remove(t.id)}
          class="shrink-0 p-0.5 opacity-60 hover:opacity-100 transition-opacity"
          aria-label="Dismiss"
        >
          <XIcon size={13} />
        </button>
      </div>
    {/each}
  </div>
{/if}
