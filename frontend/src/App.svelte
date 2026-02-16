<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Sidebar from "./lib/components/layout/Sidebar.svelte";
  import ContextPanel from "./lib/components/layout/ContextPanel.svelte";
  import Settings from "./lib/components/layout/Settings.svelte";
  import StatusBar from "./lib/components/layout/StatusBar.svelte";
  import TopBar from "./lib/components/layout/TopBar.svelte";
  import EmptyState from "./lib/components/layout/EmptyState.svelte";
  import LibraryContent from "./lib/components/layout/LibraryContent.svelte";
  import { theme } from "./lib/stores/theme";
  import type { Track, TabId } from "./lib/types";

  let activeTab = $state<TabId>("library");
  let selectedTrack = $state<Track | null>(null);
  let currentLibraryPath = $state("");
  let tracks = $state<Track[]>([]);
  let totalTrackCount = $state(0);
  let isLoadingTracks = $state(false);

  const appTitle = "Harmoniz";

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
  } from "../wailsjs/go/main/App.js";

  function parentDir(p: string): string {
    const i = Math.max(p.lastIndexOf("/"), p.lastIndexOf("\\"));
    return i <= 0 ? p : p.slice(0, i);
  }

  function applyDroppedPaths(paths: string[]) {
    console.log("applyDroppedPaths called with:", paths);
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
      WindowSetTitle(path ? `${appTitle} — ${path}` : appTitle);
    } catch (_) {}
  }

  $effect(() => {
    if (!currentLibraryPath) {
      try {
        WindowSetTitle(appTitle);
      } catch (_) {}
    }
  });

  onMount(() => {
    theme.init();
    OnFileDrop((_x: number, _y: number, paths: string[]) => {
      console.log("OnFileDrop callback triggered with paths:", paths);
      applyDroppedPaths(paths);
    }, true);

    if (currentLibraryPath) {
      loadTracks();
    }
  });

  onDestroy(() => {
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
      console.log("Starting scan for:", rootPath);
      isLoadingTracks = true;
      await ScanLibrary(rootPath);
      console.log("Scan completed");
      await loadTracks();
    } catch (error) {
      console.error("Scan failed:", error);
      isLoadingTracks = false;
    }
  }

  async function loadTracks() {
    if (!currentLibraryPath) {
      tracks = [];
      totalTrackCount = 0;
      updateWindowTitle("");
      return;
    }
    isLoadingTracks = true;
    try {
      const result = await ListTracks(currentLibraryPath, 100, 0);
      if (result && result.Tracks) {
        tracks = result.Tracks.map((t) => ({
          ...t,
          artist: t.artist_raw || "",
          album: t.album_raw || "",
        }));
        totalTrackCount = result.Total ?? 0;
        console.log("Loaded tracks:", tracks.length, "total:", result.Total);
      }
      updateWindowTitle(currentLibraryPath);
    } catch (error) {
      console.error("Failed to load tracks:", error);
      tracks = [];
      totalTrackCount = 0;
    } finally {
      isLoadingTracks = false;
    }
  }
</script>

<div
  class="flex h-screen w-screen overflow-hidden text-sm selection:bg-accent/30 font-sans"
  oncontextmenu={(e) => e.preventDefault()}
  role="presentation"
>
  <Sidebar bind:activeTab />

  <div class="flex-1 flex flex-col min-w-0 min-h-0 bg-background relative overflow-hidden">
    <main class="flex-1 flex flex-col min-w-0 min-h-0 overflow-hidden relative">
      <TopBar {activeTab} onBrowse={handleBrowse} />

      <div
        class="flex-1 relative overflow-hidden"
        oncontextmenu={(e) => e.preventDefault()}
        role="presentation"
      >
        {#if activeTab === "settings"}
          <Settings />
        {:else if !currentLibraryPath}
          <EmptyState onBrowse={handleBrowse} />
        {:else}
          <LibraryContent
            {tracks}
            {selectedTrack}
            {isLoadingTracks}
            {currentLibraryPath}
            onSelectTrack={(track) => (selectedTrack = track)}
          />
        {/if}
      </div>
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
    <ContextPanel bind:selectedTrack />
  </aside>
</div>
