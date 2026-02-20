<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { t } from "../../stores/i18n";
  import type { Track } from "../../types";
  import VirtualList from "@humanspeak/svelte-virtual-list";
  import TrackItem from "./TrackItem.svelte";
  import { playbackStore } from "../../stores/playback";

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

  // Avoid `$t(...)` auto-subscription edge cases with runes + tooling:
  // keep a reactive translation function via explicit subscription.
  let tr = $state<(key: any) => string>((key) => String(key));
  onMount(() => {
    const unsub = t.subscribe((fn) => {
      tr = fn as any;
    });
    onDestroy(unsub);

    // Enable debug mode with Ctrl+Shift+D (or Cmd+Shift+D on Mac)
    const handleKeyPress = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === "D") {
        e.preventDefault();
        debugMode = !debugMode;
        console.log(`Debug mode ${debugMode ? "enabled" : "disabled"}`);
      }
    };
    window.addEventListener("keydown", handleKeyPress);
    onDestroy(() => window.removeEventListener("keydown", handleKeyPress));
  });

  // Estimated item height (matches ROW_HEIGHT from previous implementation)
  const ESTIMATED_ITEM_HEIGHT = 85;

  // Debug mode - set to true to enable virtualization debugging
  // In production, you can toggle this via environment variable or settings
  let debugMode = $state(false);

  // Debug function to log virtualization info
  function handleDebugInfo(info: any) {
    console.log("🔍 Virtual List Debug Info:", {
      totalItems: tracks.length,
      visibleStart: info.visibleStart,
      visibleEnd: info.visibleEnd,
      renderedItems: info.visibleEnd - info.visibleStart + 1,
      bufferSize: 10,
      viewportHeight: info.viewportHeight,
      totalHeight: info.totalHeight,
    });
  }

  function handleTrackSelect(track: Track) {
    onSelectTrack(track);
    playbackStore.play(track, tracks);
  }
</script>

<div class="h-full min-w-0 overflow-hidden flex flex-col">
  {#if isLoadingTracks}
    <!-- Initial Loading State -->
    <div class="flex items-center justify-center h-full min-w-0 p-8">
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
    <div class="flex items-center justify-center h-full min-w-0 p-8">
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
    <!-- Virtual List Container -->
    <VirtualList
      items={tracks}
      defaultEstimatedItemHeight={ESTIMATED_ITEM_HEIGHT}
      bufferSize={10}
      onLoadMore={hasMoreTracks && !isLoadingMore ? onLoadMore : undefined}
      loadMoreThreshold={12}
      hasMore={hasMoreTracks}
      debug={debugMode}
      debugFunction={debugMode ? handleDebugInfo : undefined}
      containerClass="h-full flex-1 min-h-0 overflow-x-hidden"
      viewportClass="custom-scrollbar overflow-y-auto overflow-x-hidden h-full w-full"
      contentClass="max-w-5xl mx-auto p-8 overflow-x-hidden"
      itemsClass="grid grid-cols-[minmax(0,1fr)] gap-2"
    >
      {#snippet renderItem(track)}
        <div class="min-w-0 overflow-hidden">
          <TrackItem
            {track}
            isSelected={selectedTrack?.path === track.path}
            onSelect={handleTrackSelect}
          />
        </div>
      {/snippet}
    </VirtualList>

    {#if isLoadingMore}
      <div
        class="max-w-5xl mx-auto px-8 py-4 text-xs text-text-muted opacity-80 shrink-0"
      >
        Loading more…
      </div>
    {/if}

    <!-- Debug Info Panel (only visible when debugMode is true) -->
    {#if debugMode}
      <div
        class="fixed bottom-4 right-4 bg-black/80 text-white text-xs p-3 rounded-lg font-mono z-50 max-w-xs"
      >
        <div class="font-bold mb-2">🔍 Virtual List Debug</div>
        <div>Total Items: {tracks.length}</div>
        <div>Check console for details</div>
        <button
          onclick={() => (debugMode = false)}
          class="mt-2 text-xs underline text-blue-300"
        >
          Hide Debug
        </button>
      </div>
    {/if}
  {/if}
</div>
