<script lang="ts">
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import {
    MagicWandIcon,
    ArrowClockwiseIcon,
    FolderIcon,
    SpinnerIcon,
    CheckCircleIcon,
    EyeIcon,
    XIcon,
  } from "phosphor-svelte";
  // @ts-ignore
  import {
    AnalyzeForOrganize,
    PreviewFilenameTemplate,
    ApplyAllHighConfidence,
  } from "../../../../wailsjs/go/main/App.js";
  // @ts-ignore
  import { EventsOn, EventsOff } from "../../../../wailsjs/runtime/runtime.js";
  import { organizer } from "../../stores/organizer";
  import { toast } from "../../stores/toast";
  import { t } from "../../stores/i18n";
  import type { FilenamePreview } from "../../types";

  // Svelte 4 legacy mode — no runes
  export let libraryPath = "";
  export let onBrowse: () => void;

  // Store refs
  const status = organizer.status;
  const suggestions = organizer.suggestions;
  const withSuggestions = organizer.withSuggestions;
  const filenamePreviews = organizer.filenamePreviews;
  const filenameTemplate = organizer.filenameTemplate;
  const selectedTrackId = organizer.selectedTrackId;
  const highConfidenceCount = organizer.highConfidenceCount;

  // Local state
  let showAll = false;
  let showPreviewPanel = false;
  let mounted = false;
  let lastAnalyzedPath = "";
  let groupByArtist = false;
  let expandedArtists = new Set<string>();

  type SortBy = "score" | "filename";
  type SortDir = "asc" | "desc";
  let sortBy: SortBy = "score";
  let sortDir: SortDir = "desc";

  const SORT_OPTIONS: { value: SortBy; label: string }[] = [
    { value: "score", label: "Score" },
    { value: "filename", label: "File" },
  ];

  import type { OrganizerSuggestion } from "../../types";

  function sortSuggestions(
    list: OrganizerSuggestion[],
    by: SortBy,
    dir: SortDir,
  ): OrganizerSuggestion[] {
    return [...list].sort((a, b) => {
      let cmp = 0;
      switch (by) {
        case "score":
          cmp = a.score - b.score;
          break;
        case "filename":
          cmp = (a.track.filename ?? "").localeCompare(b.track.filename ?? "");
          break;
      }
      return dir === "desc" ? -cmp : cmp;
    });
  }

  // i18n
  let tr: (key: any) => string = (key: any) => String(key);
  onMount(() => {
    const unsub = t.subscribe((fn) => {
      tr = fn as any;
    });
    mounted = true;

    EventsOn("organize:analyze_done", (summary: any) => {
      if (summary?.with_suggestions !== undefined) {
        toast.success(
          `${summary.with_suggestions} suggestions · ${summary.high_confidence} high confidence`,
        );
      }
    });

    if (
      libraryPath &&
      get(status) !== "analyzing" &&
      get(status) !== "applying"
    ) {
      lastAnalyzedPath = libraryPath;
      handleAnalyze();
    }
    return () => {
      unsub();
      EventsOff("organize:analyze_done");
    };
  });

  // Re-analyze when path changes (after mount)
  $: if (mounted && libraryPath && libraryPath !== lastAnalyzedPath) {
    if (get(status) !== "analyzing" && get(status) !== "applying") {
      lastAnalyzedPath = libraryPath;
      handleAnalyze();
    }
  }

  $: displayList = sortSuggestions(
    showAll ? $suggestions : $withSuggestions,
    sortBy,
    sortDir,
  );

  type ArtistGroup = { artist: string; items: typeof displayList };
  let artistGroups: ArtistGroup[] | null = null;
  $: artistGroups = groupByArtist
    ? displayList.reduce((acc: ArtistGroup[], s) => {
        const name = (s.fields["artist"]?.value ||
          s.track.artist_raw ||
          "—") as string;
        let g = acc.find((g) => g.artist === name);
        if (!g) {
          g = { artist: name, items: [] };
          acc.push(g);
        }
        g.items.push(s);
        return acc;
      }, [])
    : null;

  // Ensure newly seen artists start expanded
  $: if (artistGroups) {
    artistGroups.forEach((g) => {
      if (!expandedArtists.has(g.artist)) expandedArtists.add(g.artist);
    });
  }

  function toggleArtistGroup(artist: string) {
    expandedArtists = new Set(expandedArtists);
    if (expandedArtists.has(artist)) {
      expandedArtists.delete(artist);
    } else {
      expandedArtists.add(artist);
    }
  }

  async function handleAnalyze() {
    if (!libraryPath) return;
    if (get(status) === "analyzing" || get(status) === "applying") return;
    organizer.startAnalysis();
    organizer.setSelectedTrack(null);
    try {
      const raw = await AnalyzeForOrganize(libraryPath);
      organizer.setResults(raw ?? []);
    } catch (e) {
      organizer.setError(e instanceof Error ? e.message : String(e));
      toast.error(`Analysis failed: ${e}`);
    }
  }

  async function handlePreviewTemplate() {
    if (!libraryPath || !get(filenameTemplate).trim()) return;
    try {
      const previews: FilenamePreview[] = await PreviewFilenameTemplate(
        libraryPath,
        get(filenameTemplate),
      );
      organizer.setPreviews(previews ?? []);
      showPreviewPanel = true;
    } catch (e) {
      toast.error(`Preview failed: ${e}`);
    }
  }

  async function handleApplyAll() {
    if (!libraryPath || get(status) === "applying") return;
    organizer.startApplying();
    try {
      const count: number = await ApplyAllHighConfidence(
        libraryPath,
        0.75,
        get(filenameTemplate),
      );
      organizer.setApplied(count);
      toast.success(`Applied ${count} suggestion${count !== 1 ? "s" : ""}`);
      await handleAnalyze();
    } catch (e) {
      toast.error(`Batch apply failed: ${e}`);
      organizer.setError(e instanceof Error ? e.message : String(e));
    }
  }

  function scoreColor(score: number): string {
    if (score >= 75) return "bg-green-500";
    if (score >= 50) return "bg-yellow-500";
    if (score > 0) return "bg-yellow-500/50";
    return "bg-white/10";
  }
</script>

<div class="relative h-full flex flex-col min-h-0 overflow-hidden">
  <!-- ───── Header ───── -->
  <div class="shrink-0 px-5 pt-4 pb-3 border-b border-border space-y-2.5">
    <!-- Top row -->
    <div class="flex items-center gap-3">
      <p class="text-xs text-text-muted flex-1">{tr("organizerDescription")}</p>
      <div class="flex items-center gap-2 shrink-0">
        {#if $status === "ready" || $status === "applying"}
          <!-- Apply All (shown when high-confidence suggestions exist) -->
          {#if $highConfidenceCount > 0}
            <button
              on:click={handleApplyAll}
              disabled={$status === "applying" || !libraryPath}
              class="px-3 py-1.5 text-xs font-bold rounded-lg bg-accent/15 text-accent border border-accent/25
                     flex items-center gap-1.5 hover:bg-accent/25 transition-colors disabled:opacity-40"
            >
              {#if $status === "applying"}
                <SpinnerIcon size={12} class="animate-spin" />
                Applying…
              {:else}
                <CheckCircleIcon size={12} />
                Apply all ({$highConfidenceCount})
              {/if}
            </button>
          {/if}
          <!-- Re-analyze -->
          <button
            on:click={handleAnalyze}
            disabled={!libraryPath || $status === "applying"}
            class="px-3 py-1.5 text-xs font-semibold rounded-lg border border-border hover:bg-white/5
                   text-text-secondary flex items-center gap-1.5 transition-colors disabled:opacity-40"
          >
            <ArrowClockwiseIcon size={12} />
            {tr("reanalyze")}
          </button>
        {:else}
          <button
            on:click={handleAnalyze}
            disabled={$status === "analyzing" || !libraryPath}
            class="px-4 py-1.5 text-xs font-bold rounded-lg bg-accent text-background
                   flex items-center gap-1.5 transition-all active:scale-95 disabled:opacity-40
                   hover:bg-accent/90 shadow-md shadow-accent/20"
          >
            {#if $status === "analyzing"}
              <SpinnerIcon size={12} class="animate-spin" />
              {tr("analyzing")}
            {:else}
              <MagicWandIcon size={12} weight="bold" />
              {tr("analyzeLibrary")}
            {/if}
          </button>
        {/if}
      </div>
    </div>

    <!-- Filename template -->
    <div class="flex items-center gap-2">
      <label
        for="org-tmpl"
        class="text-[10px] font-bold uppercase tracking-wider text-text-muted/70 shrink-0"
      >
        {tr("filenameTemplate")}
      </label>
      <input
        id="org-tmpl"
        type="text"
        bind:value={$filenameTemplate}
        class="flex-1 bg-background border border-border rounded-md px-2.5 py-1 text-xs font-mono
               text-text-primary outline-none focus:border-accent/60 transition-all"
        placeholder={"{artist} - {title}.{ext}"}
      />
      <button
        on:click={handlePreviewTemplate}
        disabled={!libraryPath || !$filenameTemplate.trim()}
        class="px-2.5 py-1 text-xs font-semibold rounded-md border border-border hover:bg-white/5
               text-text-secondary flex items-center gap-1 transition-colors disabled:opacity-40 shrink-0"
      >
        <EyeIcon size={12} />
        {tr("previewTemplate")}
      </button>
    </div>
    <p class="text-[10px] text-text-muted/40 font-mono">
      {tr("templatePlaceholders")}
    </p>
  </div>

  <!-- ───── Progress bar ───── -->
  {#if $status === "analyzing"}
    <div class="shrink-0 h-0.5 bg-background overflow-hidden">
      <div
        class="h-full bg-accent"
        style="animation: indeterminate 1.5s ease-in-out infinite alternate"
      ></div>
    </div>
  {/if}

  <!-- ───── Track list ───── -->
  <div class="flex-1 flex flex-col min-h-0 overflow-hidden">
    <!-- List header / filter toggle -->
    {#if $status === "ready" && $suggestions.length > 0}
      <div
        class="shrink-0 px-4 py-2 flex items-center gap-2 border-b border-border/50"
      >
        <!-- Count -->
        <span class="text-[10px] text-text-muted shrink-0">
          <span class="font-bold text-text-primary"
            >{$withSuggestions.length}</span
          >
          with suggestions
          {#if $suggestions.length !== $withSuggestions.length}
            · {$suggestions.length} total
          {/if}
        </span>

        <!-- Sort controls -->
        <div class="flex items-center gap-1 ml-3">
          <span
            class="text-[9px] font-bold uppercase tracking-wider text-text-muted/40 shrink-0"
            >Sort</span
          >
          <div
            class="flex rounded-md overflow-hidden border border-border/60 text-[10px] font-semibold"
          >
            {#each SORT_OPTIONS as opt}
              <button
                on:click={() => {
                  if (sortBy === opt.value) {
                    sortDir = sortDir === "desc" ? "asc" : "desc";
                  } else {
                    sortBy = opt.value;
                    sortDir = opt.value === "filename" ? "asc" : "desc";
                  }
                }}
                class="px-2 py-1 border-l first:border-l-0 border-border/60 transition-colors
                       {sortBy === opt.value
                  ? 'bg-accent/15 text-accent'
                  : 'text-text-muted/50 hover:bg-white/5 hover:text-text-muted'}"
              >
                {opt.label}{#if sortBy === opt.value}<span
                    class="ml-0.5 opacity-70"
                    >{sortDir === "desc" ? "↓" : "↑"}</span
                  >{/if}
              </button>
            {/each}
          </div>
        </div>

        <!-- View-mode toggles -->
        <div
          class="ml-auto flex rounded-md overflow-hidden border border-border text-[10px] font-bold shrink-0"
        >
          <button
            on:click={() => (showAll = false)}
            class="px-2 py-1 transition-colors {!showAll
              ? 'bg-accent/20 text-accent'
              : 'text-text-muted hover:bg-white/5'}"
          >
            Suggestions
          </button>
          <button
            on:click={() => (showAll = true)}
            class="px-2 py-1 border-l border-border transition-colors {showAll
              ? 'bg-accent/20 text-accent'
              : 'text-text-muted hover:bg-white/5'}"
          >
            All
          </button>
          <button
            on:click={() => {
              groupByArtist = !groupByArtist;
            }}
            class="px-2 py-1 border-l border-border transition-colors {groupByArtist
              ? 'bg-accent/20 text-accent'
              : 'text-text-muted hover:bg-white/5'}"
          >
            By artist
          </button>
        </div>
      </div>
    {/if}

    <div class="flex-1 overflow-y-auto min-h-0">
      {#if !libraryPath}
        <div
          class="flex flex-col items-center justify-center h-full gap-4 text-center px-8"
        >
          <div
            class="w-14 h-14 rounded-2xl bg-accent/10 flex items-center justify-center"
          >
            <MagicWandIcon size={28} class="text-accent/60" weight="thin" />
          </div>
          <div>
            <p class="text-sm text-text-secondary font-medium mb-1">
              {tr("noLibraryForOrganizer")}
            </p>
            <p class="text-xs text-text-muted">{tr("organizerEmptyHint")}</p>
          </div>
          <button
            on:click={onBrowse}
            class="px-5 py-2 bg-accent/15 text-accent rounded-xl font-bold text-sm
                   hover:bg-accent/25 transition-colors flex items-center gap-2"
          >
            <FolderIcon size={15} />
            {tr("browse")}
          </button>
        </div>
      {:else if $status === "idle"}
        <div
          class="flex flex-col items-center justify-center h-full gap-4 text-center px-8"
        >
          <div
            class="w-14 h-14 rounded-2xl bg-accent/10 flex items-center justify-center"
          >
            <MagicWandIcon size={28} class="text-accent/60" weight="thin" />
          </div>
          <div>
            <p class="text-sm text-text-secondary font-medium">
              {tr("organizerEmptyHint")}
            </p>
            <p
              class="text-[10px] text-text-muted/50 font-mono mt-1 truncate max-w-xs"
            >
              {libraryPath}
            </p>
          </div>
          <button
            on:click={handleAnalyze}
            class="px-5 py-2.5 bg-accent text-background rounded-xl font-bold text-sm
                   hover:bg-accent/90 active:scale-95 transition-all shadow-lg shadow-accent/20
                   flex items-center gap-2"
          >
            <MagicWandIcon size={15} weight="bold" />
            {tr("analyzeLibrary")}
          </button>
        </div>
      {:else if $status === "analyzing"}
        <div
          class="flex flex-col items-center justify-center h-full gap-3 text-text-muted"
        >
          <SpinnerIcon size={28} class="animate-spin text-accent" />
          <p class="text-sm">{tr("analyzing")}</p>
          <p class="text-[10px] font-mono text-text-muted/50 max-w-xs truncate">
            {libraryPath}
          </p>
        </div>
      {:else if ($status === "ready" || $status === "applying") && $withSuggestions.length === 0 && !showAll}
        <div
          class="flex flex-col items-center justify-center h-full gap-4 text-center px-8"
        >
          <CheckCircleIcon size={36} class="text-green-400/70" weight="thin" />
          <p class="text-sm text-text-secondary font-medium">
            {tr("noSuggestionsFound")}
          </p>
          <p class="text-xs text-text-muted">
            {$suggestions.length}
            {tr("filesAnalyzed")}
          </p>
          <button
            on:click={() => (showAll = true)}
            class="text-xs text-accent hover:underline"
          >
            {tr("showAllFiles")}
          </button>
        </div>
      {:else if $status === "ready" || $status === "applying"}
        <!-- Track list — full width, click sets selected in store -->
        <div class="py-2">
          {#if artistGroups}
            <!-- Grouped by artist view -->
            {#each artistGroups as group}
              <button
                type="button"
                on:click={() => toggleArtistGroup(group.artist)}
                class="w-full flex items-center gap-2 px-4 py-1.5 text-left bg-surface/30
                       border-b border-t border-border/30 hover:bg-white/5 transition-colors sticky top-0 z-10"
              >
                <span
                  class="text-[10px] font-bold uppercase tracking-wider text-text-muted"
                  >{group.artist}</span
                >
                <span class="text-[9px] text-text-muted/40"
                  >({group.items.length})</span
                >
                <span class="ml-auto text-[9px] text-text-muted/30"
                  >{expandedArtists.has(group.artist) ? "▾" : "▸"}</span
                >
              </button>
              {#if expandedArtists.has(group.artist)}
                {#each group.items as sugg (sugg.track_id)}
                  {@const hasFields = Object.keys(sugg.fields).length > 0}
                  {@const isSelected = $selectedTrackId === sugg.track_id}
                  <button
                    type="button"
                    on:click={() =>
                      organizer.setSelectedTrack(
                        isSelected ? null : sugg.track_id,
                      )}
                    class="w-full flex items-center gap-3 px-4 py-2.5 text-left transition-all
                           hover:bg-surface/50
                           {isSelected
                      ? 'bg-surface border-l-2 border-l-accent'
                      : 'border-l-2 border-l-transparent'}"
                  >
                    <div
                      class="w-2 h-2 rounded-full shrink-0 {scoreColor(
                        sugg.score,
                      )}"
                    ></div>
                    <div class="flex-1 min-w-0">
                      {#if sugg.fields["title"]}
                        <div
                          class="text-xs font-semibold text-accent truncate leading-tight"
                        >
                          {sugg.fields["title"].value}
                        </div>
                        <div
                          class="text-[9px] text-text-muted/40 line-through truncate leading-tight"
                        >
                          {sugg.track.title || sugg.track.filename}
                        </div>
                      {:else}
                        <div
                          class="text-xs font-semibold text-text-primary truncate leading-tight"
                        >
                          {sugg.track.title || sugg.track.filename}
                        </div>
                      {/if}
                      {#if sugg.fields["artist"]}
                        <div class="flex items-center gap-1 mt-0.5 min-w-0">
                          <span
                            class="text-[9px] text-text-muted/40 line-through truncate shrink min-w-0 max-w-[40%]"
                            >{sugg.track.artist_raw || "—"}</span
                          >
                          <span class="text-[9px] text-text-muted/30 shrink-0"
                            >→</span
                          >
                          <span
                            class="text-[9px] text-accent font-medium truncate shrink min-w-0"
                            >{sugg.fields["artist"].value}</span
                          >
                        </div>
                      {:else}
                        <div
                          class="text-[10px] text-text-muted truncate mt-0.5"
                        >
                          {sugg.track.artist_raw || "—"}
                        </div>
                      {/if}
                    </div>
                    {#if hasFields}
                      <span
                        class="shrink-0 text-[9px] font-bold tabular-nums {sugg.score >=
                        75
                          ? 'text-green-400'
                          : 'text-yellow-400'}">{Math.round(sugg.score)}%</span
                      >
                    {:else}
                      <CheckCircleIcon
                        size={12}
                        class="text-green-400/60 shrink-0"
                      />
                    {/if}
                  </button>
                {/each}
              {/if}
            {/each}
          {:else}
            <!-- Flat list (default) -->
            {#each displayList as sugg (sugg.track_id)}
              {@const hasFields = Object.keys(sugg.fields).length > 0}
              {@const isSelected = $selectedTrackId === sugg.track_id}

              <button
                type="button"
                on:click={() =>
                  organizer.setSelectedTrack(isSelected ? null : sugg.track_id)}
                class="w-full flex items-center gap-3 px-4 py-2.5 text-left transition-all
                       hover:bg-surface/50
                       {isSelected
                  ? 'bg-surface border-l-2 border-l-accent'
                  : 'border-l-2 border-l-transparent'}"
              >
                <!-- Score dot -->
                <div
                  class="w-2 h-2 rounded-full shrink-0 {scoreColor(sugg.score)}"
                ></div>

                <!-- Track info: current → suggestion -->
                <div class="flex-1 min-w-0">
                  {#if sugg.fields["title"]}
                    <div
                      class="text-xs font-semibold text-accent truncate leading-tight"
                    >
                      {sugg.fields["title"].value}
                    </div>
                    <div
                      class="text-[9px] text-text-muted/40 line-through truncate leading-tight"
                    >
                      {sugg.track.title || sugg.track.filename}
                    </div>
                  {:else}
                    <div
                      class="text-xs font-semibold text-text-primary truncate leading-tight"
                    >
                      {sugg.track.title || sugg.track.filename}
                    </div>
                  {/if}

                  {#if sugg.fields["artist"]}
                    <div class="flex items-center gap-1 mt-0.5 min-w-0">
                      <span
                        class="text-[9px] text-text-muted/40 line-through truncate shrink min-w-0 max-w-[40%]"
                      >
                        {sugg.track.artist_raw || "—"}
                      </span>
                      <span class="text-[9px] text-text-muted/30 shrink-0"
                        >→</span
                      >
                      <span
                        class="text-[9px] text-accent font-medium truncate shrink min-w-0"
                      >
                        {sugg.fields["artist"].value}
                      </span>
                    </div>
                  {:else}
                    <div class="text-[10px] text-text-muted truncate mt-0.5">
                      {sugg.track.artist_raw || "—"}
                    </div>
                  {/if}
                </div>

                <!-- Score / complete -->
                {#if hasFields}
                  <span
                    class="shrink-0 text-[9px] font-bold tabular-nums
                               {sugg.score >= 75
                      ? 'text-green-400'
                      : 'text-yellow-400'}"
                  >
                    {Math.round(sugg.score)}%
                  </span>
                {:else}
                  <CheckCircleIcon
                    size={12}
                    class="text-green-400/60 shrink-0"
                  />
                {/if}
              </button>
            {/each}
          {/if}
        </div>
      {/if}
    </div>
  </div>

  <!-- ───── Filename Preview Panel (modal overlay) ───── -->
  {#if showPreviewPanel}
    <div
      class="absolute inset-0 z-40 bg-background/80 backdrop-blur-sm flex items-center justify-center p-6"
    >
      <div
        class="bg-surface border border-border rounded-2xl shadow-2xl w-full max-w-2xl max-h-[80vh] flex flex-col overflow-hidden"
      >
        <div
          class="px-5 py-4 border-b border-border flex items-center justify-between"
        >
          <div>
            <h3 class="text-sm font-bold">{tr("templatePreviewTitle")}</h3>
            <p class="text-xs text-text-muted mt-0.5 font-mono">
              {$filenameTemplate}
            </p>
          </div>
          <button
            on:click={() => (showPreviewPanel = false)}
            class="p-1.5 rounded-lg hover:bg-white/5 text-text-muted transition-colors"
          >
            <XIcon size={16} />
          </button>
        </div>
        <div class="flex-1 overflow-y-auto px-5 py-3 space-y-1.5">
          {#if $filenamePreviews.length === 0}
            <p class="text-xs text-text-muted text-center py-8">
              No filenames would change.
            </p>
          {:else}
            {#each $filenamePreviews.slice(0, 200) as preview}
              <div class="text-[11px] rounded-lg bg-background p-2.5">
                <div class="font-mono text-text-muted/70 line-through truncate">
                  {preview.before}
                </div>
                <div
                  class="font-mono text-text-primary font-semibold truncate mt-0.5"
                >
                  {preview.after}
                </div>
              </div>
            {/each}
            {#if $filenamePreviews.length > 200}
              <p class="text-xs text-text-muted text-center pt-2">
                …and {$filenamePreviews.length - 200} more
              </p>
            {/if}
          {/if}
        </div>
        <div class="px-5 py-3 border-t border-border">
          <button
            on:click={() => (showPreviewPanel = false)}
            class="w-full py-2 text-xs font-bold rounded-lg border border-border hover:bg-white/5 transition-colors"
          >
            {tr("closePreview")}
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  @keyframes indeterminate {
    0% {
      transform: scaleX(0.1) translateX(-100%);
    }
    100% {
      transform: scaleX(0.8) translateX(200%);
    }
  }
</style>
