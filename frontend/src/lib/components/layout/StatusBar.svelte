<script lang="ts">
  import { t } from '../../stores/i18n'
  import { WifiHighIcon, WifiSlashIcon } from 'phosphor-svelte'

  interface Props {
    connected: boolean | null
  }

  let { connected = null }: Props = $props()
</script>

<footer class="h-8 border-t border-border bg-surface px-4 flex items-center justify-between text-[10px] font-bold uppercase tracking-widest text-text-secondary select-none">
  <div class="flex items-center gap-4">
    <div class="flex items-center gap-2">
      {#if connected === true}
        <WifiHighIcon size={14} class="text-emerald-500" />
        <span class="text-emerald-500 opacity-80">{$t('connected')}</span>
      {:else if connected === false}
        <WifiSlashIcon size={14} class="text-rose-500" />
        <span class="text-rose-500 opacity-80">{$t('disconnected')}</span>
      {:else}
        <div class="w-2 h-2 rounded-full bg-amber-500 animate-pulse"></div>
        <span class="text-amber-500 opacity-80">Connecting...</span>
      {/if}
    </div>
    <div class="w-px h-3 bg-border"></div>
    <div class="flex items-center gap-2">
      <span class="opacity-50 text-[9px]">Harmonizr Engine v0.1.2</span>
    </div>
  </div>

  <div class="flex items-center gap-3">
    <div class="w-px h-3 bg-border"></div>
    <span class="opacity-40">Build 2026.02.16</span>
  </div>
</footer>
