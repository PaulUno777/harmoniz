<script lang="ts">
  import { 
    FileAudioIcon, 
    CalendarIcon, 
    HashIcon, 
    FolderIcon, 
    TagIcon,
    FloppyDiskIcon
  } from 'phosphor-svelte'

  import type { Track } from '../../types'
  import { t } from '../../stores/i18n'
  import { formatFileSize } from '../../utils/format'

  interface Props {
    selectedTrack: Track | null
  }

  let { selectedTrack = $bindable(null) }: Props = $props()
</script>

<aside class="w-80 h-full bg-surface border-l border-border flex flex-col pt-8">
  <div class="px-6 mb-6">
    <h2 class="text-sm font-bold uppercase tracking-wider text-text-secondary opacity-70">{$t('details')}</h2>
  </div>

  {#if selectedTrack}
    <div class="flex-1 overflow-y-auto px-6 space-y-6">
      <div class="aspect-square w-full bg-background rounded-xl flex items-center justify-center border border-border group relative overflow-hidden">
        <FileAudioIcon size={64} weight="thin" class="text-text-secondary group-hover:scale-110 transition-transform duration-500" />
        <div class="absolute inset-0 bg-accent/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>
      </div>

      <div class="space-y-4">
        <div>
          <label for="track-title" class="text-[10px] font-bold uppercase text-text-secondary tracking-widest block mb-1">{$t('title')}</label>
          <input 
            id="track-title"
            type="text" 
            bind:value={selectedTrack.title}
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:ring-1 focus:ring-accent outline-none transition-all"
          />
        </div>

        <div>
          <label for="track-artist" class="text-[10px] font-bold uppercase text-text-secondary tracking-widest block mb-1">{$t('artist')}</label>
          <input 
            id="track-artist"
            type="text" 
            bind:value={selectedTrack.artist}
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:ring-1 focus:ring-accent outline-none transition-all"
          />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label for="track-year" class="text-[10px] font-bold uppercase text-text-secondary tracking-widest mb-1 flex items-center gap-1">
              <CalendarIcon size={12} /> {$t('year')}
            </label>
            <input 
              id="track-year"
              type="number" 
              bind:value={selectedTrack.year}
              class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:ring-1 focus:ring-accent outline-none transition-all"
            />
          </div>
          <div>
            <label class="text-[10px] font-bold uppercase text-text-secondary tracking-widest mb-1 flex items-center gap-1">
              <HashIcon size={12} /> {$t('size')}
            </label>
            <div class="px-3 py-2 text-sm text-text-secondary bg-background border border-border rounded-md opacity-50">
              {formatFileSize(selectedTrack.size)}
            </div>
          </div>
        </div>

        <div class="pt-2">
          <label class="text-[10px] font-bold uppercase text-text-secondary tracking-widest mb-1 flex items-center gap-1">
            <FolderIcon size={12} /> {$t('path')}
          </label>
          <div class="text-[11px] font-mono break-all text-text-secondary bg-background p-2 rounded border border-border leading-relaxed">
            {selectedTrack.path}
          </div>
        </div>
      </div>
    </div>

    <div class="p-6 border-t border-border mt-auto">
      <button class="w-full bg-accent hover:bg-accent/90 text-background font-bold py-2.5 rounded-lg flex items-center justify-center gap-2 transition-all active:scale-[0.98] shadow-lg shadow-accent/20">
        <FloppyDiskIcon size={18} weight="bold" />
        {$t('saveChanges')}
      </button>
    </div>
  {:else}
    <div class="flex-1 flex flex-col items-center justify-center text-text-secondary px-10 text-center gap-4 animate-pulse">
      <div class="w-16 h-16 bg-white/5 rounded-full flex items-center justify-center">
        <TagIcon size={32} weight="thin" />
      </div>
      <p class="text-xs font-medium leading-relaxed">{$t('selectTrackToEdit')}</p>
    </div>
  {/if}
</aside>
