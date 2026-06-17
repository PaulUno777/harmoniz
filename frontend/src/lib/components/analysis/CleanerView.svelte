<script lang="ts">
  import { t } from "../../stores/i18n";
  import { analysis } from "../../stores/analysis";
  import { toast } from "../../stores/toast";
  import DuplicateGroupCard from "./DuplicateGroupCard.svelte";
  import { ArrowClockwiseIcon, SpinnerIcon, FolderIcon, CheckCircleIcon } from "phosphor-svelte";
  // @ts-ignore
  import { DetectDuplicates } from "../../../../wailsjs/go/main/App.js";
  import type { Track } from "../../types";

  interface Props {
    onBrowse: () => void;
    libraryPath?: string;
    onDeleteTrack: (id: number) => Promise<void>;
  }

  let { onBrowse, libraryPath = "", onDeleteTrack }: Props = $props();

  const loading       = analysis.loading;
  const error         = analysis.error;
  const duplicateGroups = analysis.duplicateGroups;
  const hasRunOnce    = analysis.hasRunOnce;

  // Auto-analyze when a library path becomes available (first time only)
  $effect(() => {
    if (libraryPath && !$hasRunOnce && !$loading) {
      handleAnalyze();
    }
  });

  async function handleAnalyze() {
    if (!libraryPath || $loading) return;
    analysis.setLoading(true);
    analysis.setError(null);
    try {
      const groups = await DetectDuplicates(libraryPath);
      analysis.setDuplicates(groups ?? []);
      analysis.setHasRunOnce(true);
    } catch (e) {
      analysis.setError(String(e));
      toast.error(`Duplicate scan failed: ${e}`);
    } finally {
      analysis.setLoading(false);
    }
  }

  let resolvingAll = $state(false);

  async function handleResolveAll() {
    if (resolvingAll || $duplicateGroups.length === 0) return;
    resolvingAll = true;
    let resolved = 0;
    try {
      for (const group of [...$duplicateGroups]) {
        const sorted = [...group].sort((a, b) => qualityScore(b) - qualityScore(a));
        for (const track of sorted.slice(1)) {
          if (track.id !== undefined) {
            await onDeleteTrack(track.id);
          }
        }
        analysis.resolveGroup(group[0]?.path ?? "");
        resolved++;
      }
      toast.success(`Resolved ${resolved} duplicate group${resolved !== 1 ? "s" : ""}`);
    } catch (e) {
      toast.error(`Auto-resolve failed: ${e}`);
    } finally {
      resolvingAll = false;
    }
  }

  function qualityScore(t: Track): number {
    let s = 0;
    if (t.title)      s += 10;
    if (t.artist_raw) s += 10;
    if (t.album_raw)  s += 10;
    if (t.year)       s += 5;
    if (t.track_num)  s += 5;
    s += (t.bitrate ?? 0) / 100;
    s -= t.path.length * 0.01;
    return s;
  }

  async function handleKeepThis(keepId: number, group: Track[]) {
    for (const track of group) {
      if (track.id !== keepId && track.id !== undefined) {
        await onDeleteTrack(track.id);
      }
    }
    analysis.resolveGroup(group[0]?.path ?? "");
  }

  const totalSaveable = $derived(
    $duplicateGroups.reduce((sum, g) => {
      const f = g[0];
      return sum + (f && g.length > 1 ? (g.length - 1) * (f.size ?? 0) : 0);
    }, 0)
  );

  function formatBytes(b: number): string {
    if (b >= 1_073_741_824) return `${(b / 1_073_741_824).toFixed(1)} GB`;
    if (b >= 1_048_576)     return `${(b / 1_048_576).toFixed(1)} MB`;
    if (b >= 1_024)         return `${(b / 1_024).toFixed(0)} KB`;
    return `${b} B`;
  }
</script>

<div class="h-full flex flex-col min-h-0 overflow-hidden">

  <!-- Header -->
  <div class="shrink-0 px-8 py-4 border-b border-border">
    <div class="flex items-center justify-between gap-4">
      <div class="min-w-0">
        <h3 class="text-base font-bold text-text-primary">{$t("duplicateTracks")}</h3>
        <p class="text-xs text-text-muted mt-0.5">{$t("duplicateTracksDescription")}</p>
      </div>
      {#if libraryPath}
        <div class="flex items-center gap-2 shrink-0">
          {#if $duplicateGroups.length > 0}
            <button
              onclick={handleResolveAll}
              disabled={resolvingAll || $loading}
              class="px-3 py-1.5 text-xs font-bold rounded-lg bg-green-500/15 text-green-400
                     border border-green-500/25 flex items-center gap-1.5
                     hover:bg-green-500/25 transition-colors disabled:opacity-40"
            >
              {#if resolvingAll}
                <SpinnerIcon size={12} class="animate-spin" />
                Resolving…
              {:else}
                <CheckCircleIcon size={12} />
                Auto-resolve all
              {/if}
            </button>
          {/if}
          <button
            onclick={handleAnalyze}
            disabled={$loading}
            class="px-3 py-1.5 text-xs font-semibold rounded-lg border border-border hover:bg-white/5
                   text-text-secondary flex items-center gap-1.5 transition-colors disabled:opacity-40"
          >
            {#if $loading}
              <SpinnerIcon size={12} class="animate-spin" />
              Scanning…
            {:else}
              <ArrowClockwiseIcon size={12} />
              Re-scan
            {/if}
          </button>
        </div>
      {/if}
    </div>

    <!-- Stats strip when results exist -->
    {#if $duplicateGroups.length > 0 && !$loading}
      <div class="flex items-center gap-3 mt-3">
        <span class="text-[10px] font-bold text-text-muted uppercase tracking-wider">
          {$duplicateGroups.length} group{$duplicateGroups.length !== 1 ? "s" : ""}
        </span>
        <span class="text-[10px] text-text-muted/40">·</span>
        <span class="text-[10px] text-accent font-bold">
          {formatBytes(totalSaveable)} recoverable
        </span>
        <span class="text-[10px] text-text-muted/40">·</span>
        <span class="text-[10px] text-text-muted font-mono truncate max-w-xs" title={libraryPath}>
          {libraryPath}
        </span>
      </div>
    {/if}
  </div>

  <!-- Body -->
  <div class="flex-1 overflow-y-auto p-8">

    {#if $loading}
      <div class="flex flex-col items-center justify-center py-16 text-text-muted">
        <div class="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin mb-4"></div>
        <p class="text-sm">{$t("analyzing")}</p>
        {#if libraryPath}
          <p class="text-[10px] font-mono text-text-muted/40 mt-1 max-w-xs truncate">{libraryPath}</p>
        {/if}
      </div>

    {:else if $error}
      <div class="rounded-xl border border-red-500/30 bg-red-500/10 p-4 text-red-200 text-sm">
        <p>{$error}</p>
      </div>

    {:else if !libraryPath}
      <!-- No library loaded — show drop hint + browse -->
      <div class="flex flex-col items-center justify-center py-16 text-center gap-4">
        <div class="w-14 h-14 rounded-2xl bg-accent/10 flex items-center justify-center border-2 border-dashed border-accent/20">
          <FolderIcon size={28} class="text-accent/50" weight="thin" />
        </div>
        <div>
          <p class="text-text-secondary font-medium mb-1">No library loaded</p>
          <p class="text-xs text-text-muted">Browse a folder or drop it anywhere to start scanning</p>
        </div>
        <button
          type="button"
          onclick={onBrowse}
          class="px-4 py-2 bg-accent/20 text-accent rounded-lg font-medium text-sm hover:bg-accent/30 transition-colors"
        >
          {$t("browse")}
        </button>
      </div>

    {:else if $duplicateGroups.length === 0 && $hasRunOnce}
      <div class="flex flex-col items-center justify-center py-16 text-center gap-3">
        <CheckCircleIcon size={36} class="text-green-400/60" weight="thin" />
        <p class="text-text-secondary font-medium">{$t("noDuplicatesFound")}</p>
        <p class="text-xs text-text-muted font-mono truncate max-w-xs">{libraryPath}</p>
      </div>

    {:else if $duplicateGroups.length > 0}
      <div class="space-y-3">
        {#each $duplicateGroups as group (group[0]?.path ?? group.length.toString())}
          <DuplicateGroupCard
            {group}
            {qualityScore}
            onDelete={onDeleteTrack}
            onKeepThis={(id) => handleKeepThis(id, group)}
          />
        {/each}
      </div>
    {/if}

  </div>
</div>
