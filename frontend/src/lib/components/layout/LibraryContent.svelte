<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { t } from "../../stores/i18n";
  import type { Track } from "../../types";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { LogDebug } from "../../../../wailsjs/runtime/runtime";
  import TrackItem from "./TrackItem.svelte";

  interface Props {
    tracks: Track[];
    selectedTrack: Track | null;
    isLoadingTracks: boolean;
    currentLibraryPath: string;
    isLoadingMore: boolean; // Whether more tracks are loading
    hasMoreTracks: boolean; // Whether more tracks are available
    onSelectTrack: (track: Track) => void;
    onLoadMore: () => void; // Callback to load more tracks
  }

  let {
    tracks,
    selectedTrack,
    isLoadingTracks,
    currentLibraryPath,
    hasMoreTracks,
    isLoadingMore,
    onSelectTrack,
    onLoadMore,
  }: Props = $props();

  let scrollContainer = $state<HTMLDivElement | null>(null);
  const ROW_HEIGHT = 85; // Height of each row including content
  const ROW_GAP = 8; // Gap between rows (matches gap-2 = 0.5rem = 8px)
  const ROW_TOTAL_HEIGHT = ROW_HEIGHT + ROW_GAP; // Total height per row including gap
  const PREFETCH_ROWS = 12;

  // Avoid `$t(...)` auto-subscription edge cases with runes + tooling:
  // keep a reactive translation function via explicit subscription.
  let tr = $state<(key: any) => string>((key) => String(key));
  onMount(() => {
    const unsub = t.subscribe((fn) => {
      tr = fn as any;
    });
    onDestroy(unsub);
  });

  // Create virtualizer once when scrollContainer is available
  let rowVirtualizer = $state<ReturnType<typeof createVirtualizer<HTMLDivElement, Element>> | null>(null);

  /**
   * Create virtualizer when scrollContainer becomes available.
   * Only create once to avoid losing scroll position.
   */
  $effect(() => {
    if (scrollContainer && !rowVirtualizer) {
      rowVirtualizer = createVirtualizer<HTMLDivElement, Element>({
        count: tracks.length,
        getScrollElement: () => scrollContainer,
        estimateSize: () => ROW_TOTAL_HEIGHT,
        overscan: 10,
      });
    }
  });

  /**
   * Update virtualizer count when tracks change.
   * Use setOptions to update without recreating the virtualizer.
   */
  $effect(() => {
    const count = tracks.length;
    if (rowVirtualizer && scrollContainer) {
      // Update count using setOptions
      $rowVirtualizer?.setOptions({ count });
    }
  });

  // Create derived state for virtual items to ensure reactivity
  // Access the store value with $ to get the virtualizer instance
  let virtualItems = $derived(rowVirtualizer ? ($rowVirtualizer?.getVirtualItems() ?? []) : []);

  /**
   * Trigger backend pagination when
   * user scrolls near the end of loaded tracks.
   */
  $effect(() => {
    if (!rowVirtualizer || virtualItems.length === 0) return;
    
    const lastItem = virtualItems[virtualItems.length - 1];

    if (!lastItem) return;

    if (
      hasMoreTracks &&
      !isLoadingMore &&
      lastItem.index >= Math.max(0, tracks.length - 1 - PREFETCH_ROWS)
    ) {
      onLoadMore();
    }
  });
</script>

<!-- Scroll Container -->
<!-- Added onscroll handler and bind:this -->
<div
  bind:this={scrollContainer}
  class="h-full min-w-0 overflow-y-auto overflow-x-hidden p-8 custom-scrollbar relative"
>
  {#if isLoadingTracks}
    <!-- Initial Loading State -->
    <div class="flex items-center justify-center h-full min-w-0">
      <div class="text-center min-w-0 px-4 max-w-full">
        <div class="text-text-secondary mb-2">
          {tr("scanningLibrary")}
        </div>
        <div
          class="text-[10px] text-text-muted truncate"
          title={currentLibraryPath}
        >
          {currentLibraryPath}
        </div>
      </div>
    </div>
  {:else if tracks.length === 0}
    <!-- Empty State -->
    <div class="flex items-center justify-center h-full min-w-0">
      <div class="text-center min-w-0 px-4">
        <div class="text-text-secondary mb-2 truncate">
          {tr("noTracksFound")}
        </div>
        <div class="text-[10px] text-text-muted truncate">
          {tr("tryScanningFolder")}
        </div>
      </div>
    </div>
  {:else if tracks.length > 0}
    {LogDebug(`tracks.length: ${tracks.length}`)}
    <!-- Virtual List Container (use virtualizer if scrollContainer is ready, else fallback) -->
    {#if scrollContainer && rowVirtualizer}
      {LogDebug(`virtualItems.length: ${virtualItems.length}`)}
      <div
        class="w-full max-w-5xl mx-auto relative"
        style="height: {$rowVirtualizer?.getTotalSize() ?? 0}px"
      >
        {#each virtualItems as virtualRow}
          {#if tracks[virtualRow.index] !== undefined}
            <div
              style={`position:absolute;top:0;left:0;width:100%;height:${ROW_HEIGHT}px;transform:translateY(${virtualRow.start}px);`}
            >
              <TrackItem
                track={tracks[virtualRow.index]}
                isSelected={selectedTrack?.path === tracks[virtualRow.index].path}
                onSelect={onSelectTrack}
              />
            </div>
          {/if}
        {/each}
      </div>

      {#if isLoadingMore}
        <div class="max-w-5xl mx-auto mt-6 text-xs text-text-muted opacity-80">
          Loading more…
        </div>
      {/if}
    {:else}
      {LogDebug(`fallback: ${tracks.length}`)}
      <!-- Fallback: render normally until scrollContainer is ready -->
      <div class="grid gap-2 max-w-5xl mx-auto min-w-0">
        {#each tracks as track}
          <TrackItem
            {track}
            isSelected={selectedTrack?.path === track.path}
            onSelect={onSelectTrack}
          />
        {/each}
      </div>

      {#if isLoadingMore}
        <div class="max-w-5xl mx-auto mt-6 text-xs text-text-muted opacity-80">
          Loading more…
        </div>
      {/if}
    {/if}
  {/if}
</div>
