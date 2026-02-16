<script lang="ts">
  import { FileAudioIcon } from "phosphor-svelte";
  import { formatFileSize } from "../../utils/format";
  import { t as translateStore } from "../../stores/i18n";
  import { get } from 'svelte/store';
  import type { Track } from "../../types";

  interface Props {
    tracks: Track[];
    selectedTrack: Track | null;
    isLoadingTracks: boolean;
    currentLibraryPath: string;
    onSelectTrack: (track: Track) => void;
  }

  let {
    tracks,
    selectedTrack,
    isLoadingTracks,
    currentLibraryPath,
    onSelectTrack,
  }: Props = $props();
  let t = $state(get(translateStore));
  $effect(() => {
    t = get(translateStore);
  });
</script>

<div class="h-full min-w-0 overflow-y-auto overflow-x-hidden p-8 custom-scrollbar">
  {#if isLoadingTracks}
    <div class="flex items-center justify-center h-full min-w-0">
      <div class="text-center min-w-0 px-4 max-w-full">
        <div class="text-text-secondary mb-2">{t('scanningLibrary')}</div>
        <div class="text-[10px] text-text-muted truncate" title={currentLibraryPath}>
          {currentLibraryPath}
        </div>
      </div>
    </div>
  {:else if tracks.length === 0}
    <div class="flex items-center justify-center h-full min-w-0">
      <div class="text-center min-w-0 px-4">
        <div class="text-text-secondary mb-2 truncate">{t('noTracksFound')}</div>
        <div class="text-[10px] text-text-muted truncate">
          {t('tryScanningFolder')}
        </div>
      </div>
    </div>
  {:else}
    <div class="grid gap-2 max-w-5xl mx-auto min-w-0">
      {#each tracks as track}
        <button
          type="button"
          onclick={() => onSelectTrack(track)}
          class="flex items-center gap-4 p-4 rounded-xl border border-transparent hover:border-border hover:bg-surface/30 transition-all text-left group min-w-0
                 {selectedTrack?.path === track.path
            ? 'bg-surface border-border ring-1 ring-accent/20'
            : ''}"
        >
          <div
            class="w-11 h-11 shrink-0 bg-surface rounded-lg flex items-center justify-center border border-border group-hover:border-accent/30 transition-all group-hover:shadow-lg group-hover:shadow-accent/5"
          >
            <FileAudioIcon
              size={20}
              weight="duotone"
              class="text-text-muted group-hover:text-accent transition-colors"
            />
          </div>
          <div class="flex-1 min-w-0 overflow-hidden">
            <div
              class="font-bold text-text-primary truncate group-hover:text-accent transition-colors"
              title={track.title}
            >
              {track.title}
            </div>
            <div class="text-xs text-text-secondary truncate mt-0.5">
              {track.artist || t('unknownArtist')}
              <span class="mx-1.5 opacity-30">•</span>
              {track.album || t('unknownAlbum')}
            </div>
          </div>
          <div
            class="text-[11px] font-mono font-bold text-text-muted opacity-50 bg-white/5 px-2 py-1 rounded group-hover:opacity-100 transition-all shrink-0"
          >
            {formatFileSize(track.size)}
          </div>
        </button>
      {/each}
    </div>
  {/if}
</div>
