<script lang="ts">
  import { MusicNotesIcon, TrashIcon } from 'phosphor-svelte'

  interface Props {
    tracked: number
    cleaned?: number
    hasLibrary: boolean
  }

  let { tracked = 0, cleaned = 0, hasLibrary = false }: Props = $props()
</script>

<footer class="h-8 border-t border-border bg-surface px-4 flex items-center justify-between text-[10px] font-bold uppercase tracking-widest text-text-secondary select-none">
  <div class="flex items-center gap-4">
    {#if hasLibrary}
      <div class="flex items-center gap-2">
        <MusicNotesIcon size={12} class="text-accent shrink-0" />
        <span class="text-[9px] normal-case opacity-90">
          <span class="text-accent font-mono">{tracked}</span>
          <span class="opacity-70 ml-1">tracked</span>
        </span>
      </div>
      <div class="w-px h-3 bg-border"></div>
      <div class="flex items-center gap-2">
        <TrashIcon size={12} class="text-text-muted shrink-0" />
        <span class="text-[9px] normal-case opacity-90">
          <span class="font-mono">{cleaned}</span>
          <span class="opacity-70 ml-1">cleaned</span>
        </span>
      </div>
    {:else}
      <span class="opacity-50 text-[9px] normal-case">No library selected</span>
    {/if}
  </div>

  <div class="flex items-center gap-3 shrink-0">
    <span class="opacity-40 text-[9px]">Harmoniz v0.1.0</span>
  </div>
</footer>
