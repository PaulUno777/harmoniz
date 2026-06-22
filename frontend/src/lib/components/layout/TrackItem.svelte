<script lang="ts">
  import { FileAudioIcon } from "phosphor-svelte";
  import { formatFileSize } from "../../utils/format";
  import { subscribeI18n, type TranslateFn } from "../../utils/i18nSubscribe";
  import type { Track } from "../../types";

  interface Props {
    track: Track;
    isSelected: boolean;
    onSelect: (track: Track) => void;
  }

  let { track, isSelected, onSelect }: Props = $props();

  let tr = $state<TranslateFn>((key) => String(key));
  subscribeI18n((fn) => {
    tr = fn;
  });
</script>

<button
  type="button"
  onclick={() => onSelect(track)}
  class="w-full flex items-center gap-4 p-4 rounded-xl border border-transparent hover:border-border hover:bg-surface/30 transition-all text-left group min-w-0
         {isSelected
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
      class="font-bold text-text-primary truncate break-all group-hover:text-accent transition-colors"
      title={track.title}
    >
      {track.title}
    </div>
    <div class="text-xs text-text-secondary truncate break-all mt-0.5">
      {track.artist || tr("unknownArtist")}
      <span class="mx-1.5 opacity-30">•</span>
      {track.album || tr("unknownAlbum")}
    </div>
  </div>
  <div
    class="text-[11px] font-mono font-bold text-text-muted opacity-50 bg-white/5 px-2 py-1 rounded group-hover:opacity-100 transition-all shrink-0"
  >
    {formatFileSize(track.size)}
  </div>
</button>
