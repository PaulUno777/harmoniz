<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Sidebar from "./lib/components/layout/Sidebar.svelte";
  import ContextPanel from "./lib/components/layout/ContextPanel.svelte";
  import Settings from "./lib/components/layout/Settings.svelte";
  import StatusBar from "./lib/components/layout/StatusBar.svelte";
  import TopBar from "./lib/components/layout/TopBar.svelte";
  import EmptyState from "./lib/components/layout/EmptyState.svelte";
  import LibraryContent from "./lib/components/layout/LibraryContent.svelte";
  import FilterPanel, { type FilterState } from "./lib/components/layout/FilterPanel.svelte";
  import FloatingPlayer from "./lib/components/layout/FloatingPlayer.svelte";
  import Toast from "./lib/components/ui/Toast.svelte";
  import { theme } from "./lib/stores/theme";
  import { t } from "./lib/stores/i18n";
  import { get } from "svelte/store";
  import type { Track, TabId } from "./lib/types";

  // @ts-ignore
  import {
    OnFileDrop,
    OnFileDropOff,
    WindowSetTitle,
  } from "../wailsjs/runtime/runtime.js";
  // @ts-ignore
  import {
    OpenFolderDialog,
    ScanLibrary,
    ListTracks,
    AnalyzeArtists,
    DetectDuplicates,
    DeleteTrack,
  } from "../wailsjs/go/main/App.js";
  import { analysis } from "./lib/stores/analysis";
  import { organizer } from "./lib/stores/organizer";
  import { playbackStore } from "./lib/stores/playback";
  import { toast } from "./lib/stores/toast";
  import { updateStore } from "./lib/stores/update";
  import {
    loadSession,
    debouncedSaveSession,
    flushSessionSave,
    clearLibrarySession,
    type PersistedSession,
  } from "./lib/stores/session";

  import { devMode } from "./lib/stores/devMode";
  // Extract derived store so $orgSelectedSuggestion works (organizer is a plain object, not a store)
  const orgSelectedSuggestion = organizer.selectedSuggestion
  import OrganizerView from "./lib/components/analysis/OrganizerView.svelte";
  import OrganizerDetailPanel from "./lib/components/organizer/OrganizerDetailPanel.svelte";
  import CleanerView from "./lib/components/analysis/CleanerView.svelte";
  // @ts-ignore
  import { ApplyOrganizerSuggestion } from "../wailsjs/go/main/App.js";

  let activeTab = $state<TabId>("library");
  let selectedTrack = $state<Track | null>(null);
  let currentLibraryPath = $state("");
  let tracks = $state<Track[]>([]);
  let totalTrackCount = $state(0);

  let isLoadingTracks = $state(false);
  let isLoadingMore = $state(false);

  // Search and Filter State
  let searchQuery = $state("");
  let filters = $state<FilterState>({
    yearMin: 0,
    yearMax: 0,
    sizeMin: 0,
    sizeMax: 0,
  });
  let isFilterPanelOpen = $state(false);

  // Pagination
  const PAGE_SIZE = 200;
  let currentOffset = $state(0);

  let isRestoringSession = $state(false);
  let sessionRestored = $state(false);
  let previousTab = $state<TabId>("library");
  let initialScrollIndex = $state<number | null>(null);

  function collectSessionSnapshot(): Partial<PersistedSession> {
    return {
      libraryPath: currentLibraryPath,
      activeTab,
      searchQuery,
      filters: { ...filters },
      listOffset: currentOffset,
      selectedTrackPath: selectedTrack?.path ?? null,
      playback: playbackStore.getPersistedPlayback(),
    };
  }

  function persistAppSession() {
    debouncedSaveSession(collectSessionSnapshot());
  }

  function handleBeforeUnload() {
    playbackStore.persistPlaybackNow();
    flushSessionSave(collectSessionSnapshot());
  }

  async function restoreSession() {
    const saved = loadSession();
    activeTab = saved.activeTab;
    previousTab = saved.activeTab;
    searchQuery = saved.searchQuery;
    filters = { ...saved.filters };

    if (!saved.libraryPath) {
      sessionRestored = true;
      return;
    }

    isRestoringSession = true;
    currentLibraryPath = saved.libraryPath;
    updateWindowTitle(saved.libraryPath);

    try {
      isLoadingTracks = true;
      await ScanLibrary(saved.libraryPath);
      await loadInitialTracks();

      while (currentOffset < saved.listOffset && tracks.length < totalTrackCount) {
        await loadMoreTracks();
      }

      if (saved.selectedTrackPath) {
        selectedTrack =
          tracks.find((t) => t.path === saved.selectedTrackPath) ?? null;
      } else if (saved.listOffset > PAGE_SIZE && tracks.length > 0) {
        initialScrollIndex = tracks.length - 1;
      }

      if (saved.playback?.trackPath) {
        let track = tracks.find((t) => t.path === saved.playback!.trackPath);
        if (!track && tracks.length < totalTrackCount) {
          await loadMoreTracks();
          track = tracks.find((t) => t.path === saved.playback!.trackPath);
        }
        if (track) {
          const { trackPath: _path, ...playbackOpts } = saved.playback;
          playbackStore.restore(track, tracks, playbackOpts);
        }
      }
    } catch (error) {
      console.error("Session restore failed:", error);
      clearLibrarySession();
      currentLibraryPath = "";
      selectedTrack = null;
      tracks = [];
      totalTrackCount = 0;
      currentOffset = 0;
    } finally {
      isLoadingTracks = false;
      isRestoringSession = false;
      sessionRestored = true;
    }
  }

  function parentDir(p: string): string {
    const i = Math.max(p.lastIndexOf("/"), p.lastIndexOf("\\"));
    return i <= 0 ? p : p.slice(0, i);
  }

  function applyDroppedPaths(paths: string[]) {
    if (!paths?.length) {
      console.log("No paths provided");
      return;
    }
    const p = paths[0];
    console.log("Processing path:", p);
    const isFolder =
      p.endsWith("/") || p.endsWith("\\") || !/\.([^/\\]+)$/.test(p);
    const folder = isFolder ? p.replace(/[/\\]+$/, "") : parentDir(p);
    console.log("Extracted folder:", folder, "isFolder:", isFolder);
    if (folder) {
      currentLibraryPath = folder;
      updateWindowTitle(folder);
      console.log("Dropped folder:", folder);
      handleScan(folder);
    }
  }

  function updateWindowTitle(path: string) {
    try {
      const title = get(t)("appTitle");
      WindowSetTitle(path ? `${title}  —  ${path}` : title);
    } catch (_) {}
  }

  $effect(() => {
    $t("appTitle");
    updateWindowTitle(currentLibraryPath);
  });

  async function runStartupUpdateCheck() {
    await updateStore.init();
    await updateStore.check(true);
    if (updateStore.isDismissed()) return;
    const u = get(updateStore);
    if (u.updateAvailable && u.latestVersion) {
      toast.warning(
        get(t)("updateAvailable").replace("{version}", u.latestVersion),
      );
    }
  }

  onMount(() => {
    theme.init();
    void restoreSession().then(() => runStartupUpdateCheck());

    window.addEventListener("beforeunload", handleBeforeUnload);

    OnFileDrop((_x: number, _y: number, paths: string[]) => {
      applyDroppedPaths(paths);
    }, true);
  });

  onDestroy(() => {
    window.removeEventListener("beforeunload", handleBeforeUnload);
    playbackStore.persistPlaybackNow();
    flushSessionSave(collectSessionSnapshot());
    OnFileDropOff();
  });

  async function handleBrowse() {
    try {
      const path = await OpenFolderDialog();
      if (path) {
        currentLibraryPath = path;
        updateWindowTitle(path);
        console.log("Selected library:", path);
        await handleScan(path);
      }
    } catch (error) {
      console.error("Failed to open folder dialog:", error);
    }
  }

  async function handleScan(rootPath: string) {
    try {
      isLoadingTracks = true;
      await ScanLibrary(rootPath);

      // Reset pagination state
      currentOffset = 0;
      tracks = [];
      selectedTrack = null;

      // Load initial tracks
      await loadInitialTracks();
    } catch (error) {
      console.error("Scan failed:", error);
      isLoadingTracks = false;
    }
  }
  /**
   * Loads the first page of tracks.
   * Resets pagination state.
   */
  async function loadInitialTracks() {
    if (!currentLibraryPath) {
      return;
    }

    try {
      // Convert MB to bytes for size filters
      const sizeMinBytes = filters.sizeMin > 0 ? filters.sizeMin * 1024 * 1024 : 0;
      const sizeMaxBytes = filters.sizeMax > 0 ? filters.sizeMax * 1024 * 1024 : 0;

      const result = await ListTracks(
        currentLibraryPath,
        searchQuery,
        filters.yearMin,
        filters.yearMax,
        sizeMinBytes,
        sizeMaxBytes,
        PAGE_SIZE,
        0
      );

      tracks =
        result?.Tracks?.map((t: any) => ({
          ...t,
          artist: t.artist_raw || "",
          album: t.album_raw || "",
        })) ?? [];

      totalTrackCount = result?.Total ?? 0;
      currentOffset = tracks.length;

      console.log("Initial load:", tracks.length, "/", totalTrackCount);
    } catch (error) {
      console.error("Initial load failed:", error);
      tracks = [];
      totalTrackCount = 0;
    } finally {
      isLoadingTracks = false;
    }
  }
  /**
   * Loads additional tracks when virtualization
   * detects that user scrolls near the end.
   */
  async function loadMoreTracks() {
    if (isLoadingMore) return;
    if (tracks.length >= totalTrackCount) return;

    isLoadingMore = true;

    try {
      // Convert MB to bytes for size filters
      const sizeMinBytes = filters.sizeMin > 0 ? filters.sizeMin * 1024 * 1024 : 0;
      const sizeMaxBytes = filters.sizeMax > 0 ? filters.sizeMax * 1024 * 1024 : 0;

      const result = await ListTracks(
        currentLibraryPath,
        searchQuery,
        filters.yearMin,
        filters.yearMax,
        sizeMinBytes,
        sizeMaxBytes,
        PAGE_SIZE,
        currentOffset
      );

      const newTracks =
        result?.Tracks?.map((t: any) => ({
          ...t,
          artist: t.artist_raw || "",
          album: t.album_raw || "",
        })) ?? [];

      tracks = [...tracks, ...newTracks];
      currentOffset += newTracks.length;

      console.log("Loaded more:", tracks.length, "/", totalTrackCount);
    } catch (error) {
      console.error("Load more failed:", error);
    } finally {
      isLoadingMore = false;
    }
  }

  // Debounced search
  let searchDebounceTimer: number;
  $effect(() => {
    // Watch searchQuery changes
    searchQuery;

    if (!currentLibraryPath || isRestoringSession || !sessionRestored) return;

    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = setTimeout(() => {
      handleFilterChange();
    }, 300);
  });

  $effect(() => {
    if (!sessionRestored || isRestoringSession) return;
    currentLibraryPath;
    activeTab;
    searchQuery;
    filters.yearMin;
    filters.yearMax;
    filters.sizeMin;
    filters.sizeMax;
    currentOffset;
    selectedTrack?.path;
    persistAppSession();
  });

  async function handleFilterChange() {
    currentOffset = 0;
    tracks = [];
    isLoadingTracks = true;
    await loadInitialTracks();
  }

  function handleApplyFilters(newFilters: FilterState) {
    filters = newFilters;
    handleFilterChange();
  }

  // Count active filters
  const activeFilterCount = $derived(
    (filters.yearMin > 0 ? 1 : 0) +
    (filters.yearMax > 0 ? 1 : 0) +
    (filters.sizeMin > 0 ? 1 : 0) +
    (filters.sizeMax > 0 ? 1 : 0)
  );

  const hasMoreTracks = $derived(tracks.length < totalTrackCount);

  const analysisLoadingStore = analysis.loading;

  function handleTrackSaved(updated: Track) {
    selectedTrack = updated
    tracks = tracks.map(t => t.id === updated.id ? updated : t)
  }

  async function handleOrganizerApply(trackID: number, fields: Record<string, string>) {
    try {
      const template = get(organizer.filenameTemplate)
      const artist   = fields['artist']    ?? ''
      const title    = fields['title']     ?? ''
      const album    = fields['album']     ?? ''
      const year     = fields['year']      ? parseInt(fields['year'])      : 0
      const trackNum = fields['track_num'] ? parseInt(fields['track_num']) : 0
      await ApplyOrganizerSuggestion(trackID, artist, title, album, year, trackNum, template, currentLibraryPath)
      organizer.updateTrackSuggestion(trackID, fields)
      toast.success(get(t)('saved'))
    } catch (e) {
      toast.error(`${get(t)('saveError')}: ${e}`)
      throw e
    }
  }

  // Reload tracks whenever we switch back to the library tab so organizer/cleaner
  // edits (renames, deletes) are reflected immediately.
  $effect(() => {
    const tab = activeTab;
    if (
      tab === "library" &&
      previousTab !== "library" &&
      currentLibraryPath &&
      !isRestoringSession &&
      sessionRestored
    ) {
      currentOffset = 0;
      tracks = [];
      loadInitialTracks();
    }
    previousTab = tab;
  });

  // Sync selectedTrack with playback when prev/next is used while on the library tab.
  // The VirtualList scroll-to-selected effect in LibraryContent fires automatically.
  $effect(() => {
    const unsub = playbackStore.subscribe(state => {
      if (activeTab !== 'library' || !state.currentTrack) return
      if (state.currentTrack.path !== selectedTrack?.path) {
        const match = tracks.find(t => t.path === state.currentTrack!.path)
        if (match) selectedTrack = match
      }
    })
    return () => unsub()
  })

  // Ctrl+Shift+D toggles dev mode (shows suggestion trace in OrganizerDetailPanel).
  $effect(() => {
    function onKeydown(e: KeyboardEvent) {
      if (e.ctrlKey && e.shiftKey && e.key === 'D') {
        devMode.update(v => !v)
      }
    }
    window.addEventListener('keydown', onKeydown)
    return () => window.removeEventListener('keydown', onKeydown)
  })

  async function handleDeleteTrack(id: number) {
    await DeleteTrack(id);
    analysis.removeFromGroup(id);
  }

  async function handleRunAnalysis() {
    if (!currentLibraryPath) return;
    analysis.setError(null);
    analysis.setLoading(true);
    try {
      const [suggestions, duplicates] = await Promise.all([
        AnalyzeArtists(currentLibraryPath),
        DetectDuplicates(currentLibraryPath),
      ]);
      analysis.setArtistSuggestions((suggestions ?? []) as import("./lib/types").ArtistSuggestion[]);
      analysis.setDuplicates((duplicates ?? []) as Track[][]);
      analysis.setHasRunOnce(true);
    } catch (e) {
      analysis.setError(e instanceof Error ? e.message : String(e));
    } finally {
      analysis.setLoading(false);
    }
  }
</script>

<div
  class="flex h-screen w-screen overflow-hidden text-sm selection:bg-accent/30 font-sans"
  oncontextmenu={(e) => e.preventDefault()}
  role="presentation"
>
  <Sidebar bind:activeTab />

  <div
    class="flex-1 flex flex-col min-w-0 min-h-0 bg-background relative overflow-hidden"
  >
    <main class="flex-1 flex flex-col min-w-0 min-h-0 overflow-hidden relative">
      <TopBar 
        {activeTab} 
        onBrowse={handleBrowse} 
        {searchQuery}
        onSearchChange={(query) => searchQuery = query}
        onFilterToggle={() => isFilterPanelOpen = !isFilterPanelOpen}
        {activeFilterCount}
        onRunAnalysis={handleRunAnalysis}
        analysisLoading={$analysisLoadingStore}
        hasLibrary={!!currentLibraryPath}
      />

      <div
        class="flex-1 relative overflow-hidden"
        oncontextmenu={(e) => e.preventDefault()}
        role="presentation"
      >
        {#if activeTab === "settings"}
          <Settings />
        {:else if activeTab === "organizer"}
          <OrganizerView libraryPath={currentLibraryPath} onBrowse={handleBrowse} />
        {:else if activeTab === "cleaner"}
          <CleanerView
            onBrowse={handleBrowse}
            libraryPath={currentLibraryPath}
            onDeleteTrack={handleDeleteTrack}
          />
        {:else if !currentLibraryPath}
          <EmptyState onBrowse={handleBrowse} />
        {:else}
          <LibraryContent
            {tracks}
            {selectedTrack}
            {isLoadingTracks}
            {currentLibraryPath}
            {hasMoreTracks}
            {isLoadingMore}
            {initialScrollIndex}
            onInitialScrollDone={() => (initialScrollIndex = null)}
            onSelectTrack={(track) => (selectedTrack = track)}
            onLoadMore={loadMoreTracks}
          />
        {/if}
      </div>

      <FloatingPlayer {activeTab} />
    </main>

    <footer class="shrink-0">
      <StatusBar
        tracked={totalTrackCount}
        cleaned={0}
        hasLibrary={!!currentLibraryPath}
      />
    </footer>
  </div>

  <aside class="w-80 shrink-0 h-full overflow-hidden flex flex-col">
    {#if activeTab === 'organizer'}
      <OrganizerDetailPanel
        suggestion={$orgSelectedSuggestion}
        onApply={handleOrganizerApply}
        onClose={() => organizer.setSelectedTrack(null)}
      />
    {:else}
      <ContextPanel bind:selectedTrack onTrackSaved={handleTrackSaved} />
    {/if}
  </aside>

  <!-- Filter Panel -->
  <FilterPanel
    isOpen={isFilterPanelOpen}
    onClose={() => isFilterPanelOpen = false}
    onApply={handleApplyFilters}
    initialFilters={filters}
  />
</div>

<!-- Global toast notifications -->
<Toast />
