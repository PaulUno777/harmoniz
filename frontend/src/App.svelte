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

  // @ts-ignore
  import { OnFileDrop, OnFileDropOff } from "../wailsjs/runtime/runtime.js";
  // @ts-ignore
  import { Greet } from "../wailsjs/go/main/App.js";

  onMount(() => {
    theme.init();
    backendConnected = true; // Mocked for now
    OnFileDrop((x: number, y: number, paths: string[]) => {
      if (paths.length > 0) {
        currentLibraryPath = paths[0];
        console.log("Dropped path:", currentLibraryPath);
      }
    }, true);
  });

  onDestroy(() => {
    OnFileDropOff();
  });

  async function handleBrowse() {
    // @ts-ignore - window.runtime is provided by Wails
    const path = await window.runtime.OpenDirectoryDialog({
      Title: $t("selectLibrary"),
    });
    if (path) {
      currentLibraryPath = path;
      console.log("Selected library:", path);
    }
  }

  // Mock data for initial layout testing
  const mockTracks: Track[] = [
    {
      title: "Blinding Lights",
      artist: "The Weeknd",
      album: "After Hours",
      year: 2020,
      size: "8.4 MB",
      path: "/Music/The Weeknd/After Hours/01 Blinding Lights.mp3",
    },
    {
      title: "Levitating",
      artist: "Dua Lipa",
      album: "Future Nostalgia",
      year: 2020,
      size: "7.2 MB",
      path: "/Music/Dua Lipa/Future Nostalgia/05 Levitating.mp3",
    },
    {
      title: "Stay",
      artist: "The Kid LAROI",
      album: "F*CK LOVE",
      year: 2021,
      size: "5.1 MB",
      path: "/Music/The Kid LAROI/Stay.mp3",
    },
  ];
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

            {#if currentLibraryPath}
              <div
                class="text-[11px] font-mono text-text-secondary bg-white/5 px-3 py-1.5 rounded-full truncate max-w-sm"
              >
                {currentLibraryPath}
              </div>
            {/if}

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
              style="--wails-drop-target: drop"
              onmouseenter={() => (isDragOver = true)}
              onmouseleave={() => (isDragOver = false)}
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
            <div class="grid gap-2 max-w-5xl mx-auto">
              {#each mockTracks as track}
                <button
                  onclick={() => {
                    selectedTrack = track;
                    // Just to satisfy lint warning for Greet
                    Greet("User").then(console.log);
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
                      {track.artist} <span class="mx-1.5 opacity-30">•</span>
                      {track.album}
                    </div>
                  </div>
                  <div
                    class="text-[11px] font-mono font-bold text-text-muted opacity-50 bg-white/5 px-2 py-1 rounded group-hover:opacity-100 transition-all"
                  >
                    {track.size}
                  </div>
                </button>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    </main>

    <StatusBar connected={backendConnected} />
  </div>

  <ContextPanel bind:selectedTrack />
</div>
