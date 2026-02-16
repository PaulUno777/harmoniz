<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Sidebar from "./lib/components/layout/Sidebar.svelte";
  import ContextPanel from "./lib/components/layout/ContextPanel.svelte";
  import Settings from "./lib/components/layout/Settings.svelte";
  import StatusBar from "./lib/components/layout/StatusBar.svelte";
  import {
    MagnifyingGlassIcon,
    FunnelIcon,
    ListIcon,
    FileAudioIcon,
    MusicNotesIcon,
    FolderIcon,
  } from "phosphor-svelte";
  import { t } from "./lib/stores/i18n";
  import { theme } from "./lib/stores/theme";
  import type { Track, TabId } from "./lib/types";

  let activeTab = $state<TabId>("library");
  let selectedTrack = $state<Track | null>(null);
  let currentLibraryPath = $state("");
  let isDragOver = $state(false);
  let backendConnected = $state<boolean | null>(null);
  let tracks = $state<Track[]>([]);
  let totalTrackCount = $state(0);
  let isLoadingTracks = $state(false);

  const appTitle = "Harmoniz";

  // @ts-ignore
  import { OnFileDrop, OnFileDropOff, WindowSetTitle } from "../wailsjs/runtime/runtime.js";
  // @ts-ignore
  import { OpenFolderDialog, ScanLibrary, ListTracks } from "../wailsjs/go/main/App.js";

  function parentDir(p: string): string {
    const i = Math.max(p.lastIndexOf('/'), p.lastIndexOf('\\'));
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
    // Check if it's a folder (ends with / or \ or doesn't have a file extension)
    const isFolder = p.endsWith('/') || p.endsWith('\\') || !/\.([^/\\]+)$/.test(p);
    const folder = isFolder
      ? p.replace(/[/\\]+$/, '') // Remove trailing slashes
      : parentDir(p); // If it's a file, get parent directory
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
    backendConnected = true;
    OnFileDrop((_x: number, _y: number, paths: string[]) => {
      console.log("OnFileDrop callback triggered with paths:", paths);
      applyDroppedPaths(paths);
    }, true);
    
    // Load tracks if library path is already set
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
      // Refresh track list after scan
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

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  }

</script>

<div
  class="flex h-screen w-screen overflow-hidden text-sm selection:bg-accent/30 font-sans"
  oncontextmenu={(e) => e.preventDefault()}
  role="presentation"
>
  <Sidebar bind:activeTab />

  <div class="flex-1 flex flex-col min-w-0 bg-background relative">
    <main class="flex-1 flex flex-col min-w-0 relative">
      <!-- Top Bar -->
      <header
        class="h-16 border-b border-border flex items-center px-8 justify-between gap-8 bg-background/80 backdrop-blur-md z-10 sticky top-0"
      >
        <div class="flex items-center gap-4 flex-1">
          <h2 class="text-lg font-bold capitalize font-display">
            {$t(activeTab as any)}
          </h2>
          <div class="h-4 w-px bg-border"></div>

          <div class="flex items-center gap-3 flex-1">
            <button
              onclick={handleBrowse}
              class="px-4 py-1.5 bg-accent/10 hover:bg-accent/20 text-accent rounded-lg font-bold text-xs transition-all flex items-center gap-2 group whitespace-nowrap"
            >
              <FolderIcon size={16} />
              {$t("browse")}
            </button>

            <div class="relative flex-1 max-w-md group">
              <MagnifyingGlassIcon
                size={18}
                class="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary group-focus-within:text-accent transition-colors"
              />
              <input
                type="text"
                placeholder={$t("search")}
                class="w-full bg-surface border border-border rounded-full py-2 pl-10 pr-4 outline-none focus:ring-1 focus:ring-accent transition-all"
              />
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button
            class="p-2 hover:bg-white/5 rounded-lg text-text-secondary transition-colors"
            title="Filter"
          >
            <FunnelIcon size={20} />
          </button>
          <button
            class="p-2 hover:bg-white/5 rounded-lg text-text-secondary transition-colors"
            title="View Options"
          >
            <ListIcon size={20} />
          </button>
        </div>
      </header>

      <!-- Content Area -->
      <div
        class="flex-1 relative overflow-hidden"
        oncontextmenu={(e) => e.preventDefault()}
        role="presentation"
      >
        {#if activeTab === "settings"}
          <Settings />
        {:else if !currentLibraryPath}
          <!-- Empty state / Drop Zone -->
          <div
            class="h-full flex flex-col items-center justify-center p-12 text-center"
          >
            <div
              class="w-full max-w-lg aspect-video border-2 border-dashed rounded-3xl flex flex-col items-center justify-center gap-6 transition-all duration-500 pt-4
                     {isDragOver
                ? 'border-accent bg-accent/5 scale-105 shadow-2xl shadow-accent/10'
                : 'border-border bg-surface/30 hover:border-text-muted hover:bg-surface/50'}"
              role="button"
              tabindex="0"
              data-file-drop-target
              style="--wails-drop-target: drop"
              ondragover={(e) => { e.preventDefault(); isDragOver = true; }}
              ondragleave={() => { isDragOver = false; }}
              ondrop={(e) => { e.preventDefault(); isDragOver = false; }}
            >
              <div
                class="w-16 h-16 bg-accent/10 rounded-2xl flex items-center justify-center text-accent"
              >
                <MusicNotesIcon size={32} weight="duotone" />
              </div>
              <div>
                <h3 class="text-xl font-bold font-display mb-2">
                  {$t("emptyStateTitle")}
                </h3>
                <p
                  class="text-text-secondary text-sm max-w-xs mx-auto leading-relaxed"
                >
                  {$t("emptyStateSubtitle")}
                </p>
              </div>

              <div class="flex items-center gap-3">
                <button
                  onclick={handleBrowse}
                  class="px-8 py-2.5 bg-accent text-background font-bold rounded-xl hover:scale-105 transition-all shadow-lg shadow-accent/20 active:scale-95"
                >
                  {$t("openFolder")}
                </button>
              </div>

              <span
                class="text-[10px] uppercase tracking-widest font-bold text-text-muted opacity-60"
              >
                {$t("dropZone")}
              </span>
            </div>
          </div>
        {:else}
          <div class="h-full overflow-y-auto p-8 custom-scrollbar">
            {#if isLoadingTracks}
              <div class="flex items-center justify-center h-full">
                <div class="text-center">
                  <div class="text-text-secondary mb-2">Scanning library...</div>
                  <div class="text-[10px] text-text-muted">{currentLibraryPath}</div>
                </div>
              </div>
            {:else if tracks.length === 0}
              <div class="flex items-center justify-center h-full">
                <div class="text-center">
                  <div class="text-text-secondary mb-2">No tracks found</div>
                  <div class="text-[10px] text-text-muted">Try scanning a folder with audio files</div>
                </div>
              </div>
            {:else}
              <div class="grid gap-2 max-w-5xl mx-auto">
                {#each tracks as track}
                  <button
                    onclick={() => {
                      selectedTrack = track;
                    }}
                    class="flex items-center gap-4 p-4 rounded-xl border border-transparent hover:border-border hover:bg-surface/30 transition-all text-left group
                           {selectedTrack?.path === track.path
                      ? 'bg-surface border-border ring-1 ring-accent/20'
                      : ''}"
                  >
                    <div
                      class="w-11 h-11 bg-surface rounded-lg flex items-center justify-center border border-border group-hover:border-accent/30 transition-all group-hover:shadow-lg group-hover:shadow-accent/5"
                    >
                      <FileAudioIcon
                        size={20}
                        weight="duotone"
                        class="text-text-muted group-hover:text-accent transition-colors"
                      />
                    </div>
                    <div class="flex-1 min-w-0">
                      <div
                        class="font-bold text-text-primary truncate group-hover:text-accent transition-colors"
                      >
                        {track.title}
                      </div>
                      <div class="text-xs text-text-secondary truncate mt-0.5">
                        {track.artist || "Unknown Artist"} <span class="mx-1.5 opacity-30">•</span>
                        {track.album || "Unknown Album"}
                      </div>
                    </div>
                    <div
                      class="text-[11px] font-mono font-bold text-text-muted opacity-50 bg-white/5 px-2 py-1 rounded group-hover:opacity-100 transition-all"
                    >
                      {formatFileSize(track.size)}
                    </div>
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </main>

    <StatusBar
      tracked={totalTrackCount}
      cleaned={0}
      hasLibrary={!!currentLibraryPath}
    />
  </div>

  <ContextPanel bind:selectedTrack />
</div>
