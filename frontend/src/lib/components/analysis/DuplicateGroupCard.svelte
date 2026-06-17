<script lang="ts">
  import { formatFileSize } from "../../utils/format";
  import { t } from "../../stores/i18n";
  import type { Track } from "../../types";
  import { StarIcon, TrashIcon, CheckIcon, SpinnerIcon } from "phosphor-svelte";

  interface Props {
    tracks: Track[];
    recommendedKeepId: number;
    onDelete: (id: number) => Promise<void>;
    onKeepThis: (keepId: number) => Promise<void>;
  }

  let { tracks, recommendedKeepId, onDelete, onKeepThis }: Props = $props();

  let expanded  = $state(true);
  let resolving = $state(false);

  // Sort so recommended (backend-scored highest quality) is always first
  const ranked = $derived.by(() =>
    [...tracks].sort((a, b) => {
      if (a.id === recommendedKeepId) return -1;
      if (b.id === recommendedKeepId) return 1;
      return 0;
    })
  );

  const recommended = $derived(ranked[0]);

  const saveBytes = $derived(
    tracks.length > 1 ? (tracks.length - 1) * (tracks[0]?.size ?? 0) : 0
  );

  async function handleKeepThis(keepId: number) {
    resolving = true;
    try {
      await onKeepThis(keepId);
    } finally {
      resolving = false;
    }
  }

  async function handleDelete(id: number) {
    resolving = true;
    try {
      await onDelete(id);
    } finally {
      resolving = false;
    }
  }

  async function handleAutoResolve() {
    if (!recommended?.id) return;
    resolving = true;
    try {
      await onKeepThis(recommended.id);
    } finally {
      resolving = false;
    }
  }

  function shortPath(path: string): string {
    const parts = path.split("/");
    return parts.length > 3 ? "…/" + parts.slice(-2).join("/") : path;
  }
</script>

<div class="rounded-xl border border-border bg-surface/50 overflow-hidden">

  <!-- Group header — two separate click zones to avoid button-in-button -->
  <div class="flex items-center justify-between px-4 py-3 hover:bg-white/5 transition-colors">
    <!-- Left: click to expand/collapse -->
    <button
      type="button"
      class="flex items-center gap-3 min-w-0 text-left flex-1"
      onclick={() => (expanded = !expanded)}
    >
      <span class="text-text-primary font-medium text-sm truncate" title={tracks[0]?.filename ?? tracks[0]?.path}>
        {tracks[0]?.filename ?? tracks[0]?.title ?? tracks[0]?.path ?? "—"}
      </span>
      <span class="text-xs text-text-muted shrink-0 bg-white/5 px-2 py-0.5 rounded-full">
        {$t("andNDuplicates").replace("{n}", String(tracks.length - 1))}
      </span>
      {#if saveBytes > 0}
        <span class="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-accent/15 text-accent shrink-0">
          {$t("saveMB").replace("{size}", formatFileSize(saveBytes))}
        </span>
      {/if}
    </button>
    <!-- Right: actions (separate from the expand toggle) -->
    <div class="flex items-center gap-2 shrink-0 ml-2">
      {#if resolving}
        <SpinnerIcon size={14} class="animate-spin text-accent" />
      {:else if expanded}
        <button
          type="button"
          onclick={handleAutoResolve}
          class="text-[10px] font-bold px-2.5 py-1 rounded-md bg-green-500/15 text-green-400
                 border border-green-500/20 hover:bg-green-500/25 transition-colors"
        >
          Auto-resolve
        </button>
      {/if}
      <span class="text-text-muted text-xs">
        {#if tracks[0]?.hash_partial}
          {tracks[0].hash_partial.slice(0, 8)}…
        {/if}
      </span>
    </div>
  </div>

  <!-- Expanded track list -->
  {#if expanded}
    <div class="border-t border-border/50">
      {#each ranked as track, i (track.path)}
        {@const isRecommended = i === 0}
        <div
          class="flex items-start gap-3 px-4 py-3 border-b border-border/10 last:border-0 transition-colors
                 {isRecommended ? 'bg-green-500/5' : 'hover:bg-white/5'}"
        >
          <!-- Recommended star -->
          <div class="shrink-0 mt-0.5 w-4">
            {#if isRecommended}
              <StarIcon size={14} weight="fill" class="text-green-400/70" />
            {/if}
          </div>

          <!-- Track info -->
          <div class="flex-1 min-w-0 space-y-0.5">
            <div class="text-xs font-mono text-text-primary truncate" title={track.path}>
              {shortPath(track.path)}
            </div>
            <div class="flex items-center gap-3 text-[10px] text-text-muted">
              <span>{formatFileSize(track.size ?? 0)}</span>
              {#if track.bitrate}
                <span class="opacity-50">·</span>
                <span>{track.bitrate} kbps</span>
              {/if}
              {#if track.title && track.artist_raw}
                <span class="opacity-50">·</span>
                <span class="truncate max-w-[120px]">{track.artist_raw} — {track.title}</span>
              {/if}
              {#if isRecommended}
                <span class="text-green-400/70 font-bold uppercase text-[9px]">recommended</span>
              {/if}
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center gap-1.5 shrink-0">
            <button
              type="button"
              disabled={resolving}
              onclick={(e) => { e.stopPropagation(); if (track.id !== undefined) handleKeepThis(track.id); }}
              class="px-2.5 py-1 text-[10px] font-bold text-green-400 bg-green-400/10
                     border border-green-400/20 rounded hover:bg-green-400/20 transition-colors
                     flex items-center gap-1 disabled:opacity-40"
              title="Keep this copy, delete all others"
            >
              <CheckIcon size={10} weight="bold" />
              Keep
            </button>
            <button
              type="button"
              disabled={resolving}
              onclick={(e) => { e.stopPropagation(); if (track.id !== undefined) handleDelete(track.id); }}
              class="px-2.5 py-1 text-[10px] font-bold text-red-400 bg-red-400/10
                     border border-red-400/20 rounded hover:bg-red-400/20 transition-colors
                     flex items-center gap-1 disabled:opacity-40"
              title="Delete this copy"
            >
              <TrashIcon size={10} weight="bold" />
              Delete
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

</div>
