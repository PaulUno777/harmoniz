<script lang="ts">
  import { onMount } from "svelte";
  import {
    FloppyDiskIcon,
    XIcon,
    CheckCircleIcon,
    ArrowCounterClockwiseIcon,
    MusicNoteSimpleIcon,
    SpinnerIcon,
    CalendarIcon,
    HashIcon,
    FolderIcon,
    PlayIcon,
    PauseIcon,
  } from "phosphor-svelte";
  import type { OrganizerSuggestion } from "../../types";
  import { t } from "../../stores/i18n";
  import { organizer } from "../../stores/organizer";
  import { playbackStore } from "../../stores/playback";

  // Svelte 4 legacy mode — no runes
  export let suggestion: OrganizerSuggestion | null = null;
  export let onApply: (
    trackID: number,
    fields: Record<string, string>,
  ) => Promise<void>;
  export let onClose: () => void;

  let tr: (key: any) => string = (key: any) => String(key);
  onMount(() => {
    const unsub = t.subscribe((fn) => {
      tr = fn as any;
    });
    return unsub;
  });

  // Per-field edit state
  let editArtist = "";
  let editTitle = "";
  let editAlbum = "";
  let editYear = "";
  let editTrackNum = "";

  let suggestedFields: Set<string> = new Set();
  let pendingConfirm = false;
  let applying = false;
  let lastTrackId: number | null = null;

  // Reset when a different track is selected
  $: if (suggestion && suggestion.track_id !== lastTrackId) {
    lastTrackId = suggestion.track_id;
    pendingConfirm = false;
    applying = false;
    loadFromSuggestion(suggestion);
  }

  function loadFromSuggestion(s: OrganizerSuggestion) {
    const tk = s.track;
    const f = s.fields;
    suggestedFields.clear();
    for (const key in f) {
      suggestedFields.add(key);
    }
    editArtist = f["artist"] ? f["artist"].value : (tk.artist_raw ?? "");
    editTitle = f["title"] ? f["title"].value : (tk.title ?? "");
    editAlbum = f["album"] ? f["album"].value : (tk.album_raw ?? "");
    editYear = f["year"] ? f["year"].value : tk.year > 0 ? String(tk.year) : "";
    editTrackNum = f["track_num"]
      ? f["track_num"].value
      : tk.track_num
        ? String(tk.track_num)
        : "";
  }

  // Reactive diff — all local vars in args so Svelte 4 tracks them
  $: changedFields = computeChanges(
    suggestion,
    editArtist,
    editTitle,
    editAlbum,
    editYear,
    editTrackNum,
  );
  $: hasChanges = Object.keys(changedFields).length > 0;

  function computeChanges(
    s: OrganizerSuggestion | null,
    artist: string,
    title: string,
    album: string,
    year: string,
    trackNum: string,
  ): Record<string, { from: string; to: string }> {
    if (!s) return {};
    const tk = s.track;
    const result: Record<string, { from: string; to: string }> = {};
    const check = (key: string, from: string, to: string) => {
      if (to !== from) result[key] = { from, to };
    };
    check("artist", tk.artist_raw ?? "", artist);
    check("title", tk.title ?? "", title);
    check("album", tk.album_raw ?? "", album);
    check("year", tk.year > 0 ? String(tk.year) : "", year);
    check("track_num", tk.track_num ? String(tk.track_num) : "", trackNum);
    return result;
  }

  function revertField(key: string) {
    if (!suggestion) return;
    const tk = suggestion.track;
    if (key === "artist") editArtist = tk.artist_raw ?? "";
    if (key === "title") editTitle = tk.title ?? "";
    if (key === "album") editAlbum = tk.album_raw ?? "";
    if (key === "year") editYear = tk.year > 0 ? String(tk.year) : "";
    if (key === "track_num")
      editTrackNum = tk.track_num ? String(tk.track_num) : "";
  }

  function revertAll() {
    if (!suggestion) return;
    loadFromSuggestion(suggestion);
  }

  async function confirmApply() {
    if (!suggestion || applying || !hasChanges) return;
    applying = true;
    try {
      const fields: Record<string, string> = {};
      for (const [key, ch] of Object.entries(changedFields)) {
        fields[key] = ch.to;
      }
      await onApply(suggestion.track_id, fields);
      pendingConfirm = false;
    } finally {
      applying = false;
    }
  }

  function confClass(conf: number): string {
    if (conf >= 0.8) return "text-green-400";
    if (conf >= 0.6) return "text-yellow-400";
    return "text-text-muted";
  }

  let fieldLabels: Record<string, string> = {};
  $: fieldLabels = {
    artist: tr("artist"),
    title: tr("title"),
    album: tr("album"),
    year: tr("year"),
    track_num: tr("trackNum"),
  };

  $: saveNChangesText = (n: number) =>
    tr("saveNChangesQuestion").replace("{n}", String(n));

  const filenameTemplate = organizer.filenameTemplate;

  // Compute the suggested filename from the template + current edit values
  $: suggestedFilename = computeFilename(
    suggestion,
    $filenameTemplate,
    editArtist,
    editTitle,
    editAlbum,
    editYear,
    editTrackNum,
  );

  function computeFilename(
    s: OrganizerSuggestion | null,
    tmpl: string,
    artist: string,
    title: string,
    album: string,
    year: string,
    trackNum: string,
  ): string {
    if (!s || !tmpl) return "";
    const ext = s.track.filename!.split(".").pop() ?? "";
    const safeArtist = sanitizePart(artist || "Unknown Artist");
    const safeTitle = sanitizePart(title || "Untitled");
    const safeAlbum = sanitizePart(album || "Unknown Album");
    const tn = trackNum ? trackNum.padStart(2, "0") : "";
    let r = tmpl;
    r = r.replace(/{artist}/g, safeArtist);
    r = r.replace(/{album}/g, safeAlbum);
    r = r.replace(/{title}/g, safeTitle);
    r = r.replace(/{year}/g, year);
    r = r.replace(/{track:02d}/g, tn);
    r = r.replace(/{track}/g, trackNum);
    r = r.replace(/{ext}/g, ext);
    return r;
  }

  function sanitizePart(s: string): string {
    return s.replace(/[\\:*?"<>|]/g, "-").trim();
  }

  // Field suggestion objects (null when no suggestion for that field)
  $: artistField = suggestion?.fields["artist"] ?? null;
  $: titleField = suggestion?.fields["title"] ?? null;
  $: albumField = suggestion?.fields["album"] ?? null;
  $: yearField = suggestion?.fields["year"] ?? null;
  $: trackField = suggestion?.fields["track_num"] ?? null;

  // Which fields have pending edits vs. stored value
  $: artistChanged = "artist" in changedFields;
  $: titleChanged = "title" in changedFields;
  $: albumChanged = "album" in changedFields;
  $: yearChanged = "year" in changedFields;
  $: trackChanged = "track_num" in changedFields;

  // Playback state
  $: isThisPlaying =
    $playbackStore?.currentTrack?.path === suggestion?.track?.path &&
    $playbackStore?.isPlaying;

  function togglePlay() {
    if (!suggestion?.track) return;
    if (isThisPlaying) {
      playbackStore.pause();
    } else {
      playbackStore.play(suggestion.track as any, [suggestion.track as any]);
    }
  }
</script>

{#if !suggestion}
  <div
    class="h-full flex flex-col items-center justify-center text-center px-10 gap-4"
  >
    <div
      class="w-16 h-16 bg-white/5 rounded-full flex items-center justify-center"
    >
      <MusicNoteSimpleIcon size={32} weight="thin" class="text-text-muted/30" />
    </div>
    <p class="text-xs text-text-muted/60 leading-relaxed">
      {tr("selectTrackToViewSuggestions")}
    </p>
  </div>
{:else}
  <div class="h-full flex flex-col bg-surface border-l border-border">
    <!-- ── Header ── -->
    <div class="shrink-0 px-6 pt-5 pb-0 flex items-center justify-between">
      <h2
        class="text-[11px] font-bold uppercase tracking-wider text-text-secondary opacity-60"
      >
        {tr("suggestions")}
      </h2>
      <div class="flex items-center gap-1">
        <!-- Play/pause to audition before applying -->
        <button
          on:click={togglePlay}
          title={isThisPlaying ? tr("pause") : tr("playToPreview")}
          class="p-1.5 rounded-lg hover:bg-white/5 transition-colors
                 {isThisPlaying
            ? 'text-accent'
            : 'text-text-muted hover:text-text-primary'}"
        >
          {#if isThisPlaying}
            <PauseIcon size={14} weight="fill" />
          {:else}
            <PlayIcon size={14} weight="fill" />
          {/if}
        </button>
        <button
          on:click={onClose}
          class="p-1.5 rounded-lg hover:bg-white/5 text-text-muted hover:text-text-primary transition-colors"
        >
          <XIcon size={14} />
        </button>
      </div>
    </div>

    <!-- ── Scrollable body ── -->
    <div class="flex-1 overflow-y-auto px-6 py-4 space-y-4 min-h-0">
      <!-- Art placeholder with confidence bar -->
      <div
        class="aspect-square w-full bg-background rounded-xl flex flex-col items-center justify-center
                  border border-border relative overflow-hidden"
      >
        <div class="flex flex-col items-center gap-2 text-text-secondary/30">
          <MusicNoteSimpleIcon size={48} weight="thin" />
        </div>
        {#if Object.keys(suggestion.fields).length > 0}
          <div
            class="absolute bottom-0 left-0 right-0 px-3 pb-3 pt-6
                      bg-linear-to-t from-background/80 to-transparent"
          >
            <div class="flex items-center justify-between mb-1.5">
              <span
                class="text-[9px] font-bold uppercase tracking-wider text-text-muted/50"
                >{tr("confidence")}</span
              >
              <span
                class="text-[10px] font-bold tabular-nums
                           {suggestion.score >= 75
                  ? 'text-green-400'
                  : suggestion.score >= 50
                    ? 'text-yellow-400'
                    : 'text-orange-400'}"
              >
                {Math.round(suggestion.score)}%
              </span>
            </div>
            <div class="h-1 bg-white/5 rounded-full overflow-hidden">
              <div
                class="h-full rounded-full transition-all
                       {suggestion.score >= 75
                  ? 'bg-green-500'
                  : suggestion.score >= 50
                    ? 'bg-yellow-500'
                    : 'bg-orange-500'}"
                style="width: {suggestion.score}%"
              ></div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Quick info strip -->
      <div class="text-center -mt-1 pb-1">
        <div class="text-sm font-bold text-text-primary truncate">
          {suggestion.fields["title"]?.value ||
            suggestion.track.title ||
            suggestion.track.filename}
        </div>
        <div class="text-xs text-text-muted truncate mt-0.5">
          {suggestion.fields["artist"]?.value ||
            suggestion.track.artist_raw ||
            "—"}
          {#if suggestion.track.album_raw || suggestion.fields["album"]?.value}
            <span class="opacity-30 mx-1">·</span>{suggestion.fields["album"]
              ?.value || suggestion.track.album_raw}
          {/if}
        </div>
      </div>

      <!-- ── Fields ── -->
      <div class="space-y-3">
        <!-- Artist -->
        <div>
          <div class="flex items-center gap-2 mb-1">
            <label for="org-artist" class="field-label"
              >{fieldLabels["artist"]}</label
            >
            {#if artistField}
              <span
                class="text-[9px] px-1.5 py-0.5 rounded-full bg-white/5 text-text-muted uppercase leading-none"
              >
                {artistField.source}
              </span>
              <span
                class="text-[9px] font-bold {confClass(artistField.confidence)}"
              >
                {Math.round(artistField.confidence * 100)}%
              </span>
            {/if}
            {#if artistChanged}
              <button
                on:click={() => revertField("artist")}
                class="ml-auto p-1 rounded hover:bg-white/5 text-text-muted/40 hover:text-text-muted transition-colors"
                title={tr("revert")}
              >
                <ArrowCounterClockwiseIcon size={11} />
              </button>
            {/if}
          </div>
          <input
            id="org-artist"
            type="text"
            bind:value={editArtist}
            class="field-input {artistChanged
              ? 'border-accent/50 ring-1 ring-accent/20'
              : ''}"
          />
          {#if artistChanged && changedFields["artist"]?.from}
            <div class="text-[10px] text-text-muted/40 mt-1 font-mono truncate">
              <span class="line-through">{changedFields["artist"].from}</span>
            </div>
          {/if}
        </div>

        <!-- Title -->
        <div>
          <div class="flex items-center gap-2 mb-1">
            <label for="org-title" class="field-label"
              >{fieldLabels["title"]}</label
            >
            {#if titleField}
              <span
                class="text-[9px] px-1.5 py-0.5 rounded-full bg-white/5 text-text-muted uppercase leading-none"
              >
                {titleField.source}
              </span>
              <span
                class="text-[9px] font-bold {confClass(titleField.confidence)}"
              >
                {Math.round(titleField.confidence * 100)}%
              </span>
            {/if}
            {#if titleChanged}
              <button
                on:click={() => revertField("title")}
                class="ml-auto p-1 rounded hover:bg-white/5 text-text-muted/40 hover:text-text-muted transition-colors"
                title={tr("revert")}
              >
                <ArrowCounterClockwiseIcon size={11} />
              </button>
            {/if}
          </div>
          <input
            id="org-title"
            type="text"
            bind:value={editTitle}
            class="field-input {titleChanged
              ? 'border-accent/50 ring-1 ring-accent/20'
              : ''}"
          />
          {#if titleChanged && changedFields["title"]?.from}
            <div class="text-[10px] text-text-muted/40 mt-1 font-mono truncate">
              <span class="line-through">{changedFields["title"].from}</span>
            </div>
          {/if}
        </div>

        <!-- Album -->
        <div>
          <div class="flex items-center gap-2 mb-1">
            <label for="org-album" class="field-label"
              >{fieldLabels["album"]}</label
            >
            {#if albumField}
              <span
                class="text-[9px] px-1.5 py-0.5 rounded-full bg-white/5 text-text-muted uppercase leading-none"
              >
                {albumField.source}
              </span>
              <span
                class="text-[9px] font-bold {confClass(albumField.confidence)}"
              >
                {Math.round(albumField.confidence * 100)}%
              </span>
            {/if}
            {#if albumChanged}
              <button
                on:click={() => revertField("album")}
                class="ml-auto p-1 rounded hover:bg-white/5 text-text-muted/40 hover:text-text-muted transition-colors"
                title={tr("revert")}
              >
                <ArrowCounterClockwiseIcon size={11} />
              </button>
            {/if}
          </div>
          <input
            id="org-album"
            type="text"
            bind:value={editAlbum}
            class="field-input {albumChanged
              ? 'border-accent/50 ring-1 ring-accent/20'
              : ''}"
          />
          {#if albumChanged && changedFields["album"]?.from}
            <div class="text-[10px] text-text-muted/40 mt-1 font-mono truncate">
              <span class="line-through">{changedFields["album"].from}</span>
            </div>
          {/if}
        </div>

        <!-- Year + Track # -->
        <div class="grid grid-cols-2 gap-3">
          <!-- Year -->
          <div>
            <div class="flex items-center gap-1.5 mb-1">
              <CalendarIcon size={10} class="text-text-secondary/50" />
              <label for="org-year" class="field-label"
                >{fieldLabels["year"]}</label
              >
              {#if yearField}
                <span
                  class="text-[9px] font-bold {confClass(yearField.confidence)}"
                >
                  {Math.round(yearField.confidence * 100)}%
                </span>
              {/if}
            </div>
            <input
              id="org-year"
              type="number"
              bind:value={editYear}
              min="0"
              max="9999"
              placeholder="0"
              class="field-input {yearChanged
                ? 'border-accent/50 ring-1 ring-accent/20'
                : ''}"
            />
          </div>

          <!-- Track # -->
          <div>
            <div class="flex items-center gap-1.5 mb-1">
              <HashIcon size={10} class="text-text-secondary/50" />
              <label for="org-tracknum" class="field-label"
                >{fieldLabels["track_num"]}</label
              >
              {#if trackField}
                <span
                  class="text-[9px] font-bold {confClass(
                    trackField.confidence,
                  )}"
                >
                  {Math.round(trackField.confidence * 100)}%
                </span>
              {/if}
            </div>
            <input
              id="org-tracknum"
              type="number"
              bind:value={editTrackNum}
              min="0"
              placeholder="0"
              class="field-input {trackChanged
                ? 'border-accent/50 ring-1 ring-accent/20'
                : ''}"
            />
          </div>
        </div>

        <!-- Current filename (read-only) -->
        <div>
          <span class="field-label flex items-center gap-1 mb-1">
            <FolderIcon size={10} />{tr("currentFile")}
          </span>
          <div
            class="px-3 py-1.5 text-[11px] font-mono text-text-muted/50
                      bg-background border border-border rounded-md break-all select-all leading-relaxed
                      line-through"
          >
            {suggestion.track.filename}
          </div>
        </div>

        <!-- Suggested filename / path from template -->
        {#if suggestedFilename && suggestedFilename !== suggestion.track.filename}
          <div>
            <span class="field-label flex items-center gap-1 mb-1">
              <FolderIcon size={10} />{tr("suggestedName")}
            </span>
            <div
              class="px-3 py-1.5 text-[11px] font-mono text-accent/90
                        bg-accent/5 border border-accent/20 rounded-md break-all select-all leading-relaxed"
            >
              {suggestedFilename}
            </div>
            {#if $filenameTemplate.includes("/")}
              <p class="text-[9px] text-text-muted/40 mt-1">
                {tr("fileWillBeMovedHint")}
              </p>
            {/if}
          </div>
        {/if}

        <!-- Complete state -->
        {#if Object.keys(suggestion.fields).length === 0}
          <div class="flex flex-col items-center gap-2 py-4 text-center">
            <CheckCircleIcon
              size={28}
              weight="thin"
              class="text-green-400/60"
            />
            <p class="text-xs text-green-400/70 font-medium">
              {tr("metadataLooksGood")}
            </p>
            <p class="text-[10px] text-text-muted/40">
              {tr("canEditManually")}
            </p>
          </div>
        {/if}
      </div>
    </div>

    <!-- ── Footer ── -->
    <div class="shrink-0 p-5 border-t border-border">
      {#if pendingConfirm}
        <!-- Confirmation summary -->
        <div
          class="mb-3 rounded-xl bg-background border border-border/50 p-3 space-y-1.5"
        >
          <p class="text-xs font-bold text-text-primary">
            {saveNChangesText(Object.keys(changedFields).length)}
          </p>
          {#each Object.entries(changedFields) as [key, ch]}
            <div class="flex items-center gap-2 text-[10px] text-text-muted">
              <span
                class="font-bold text-text-secondary uppercase w-12 shrink-0"
                >{fieldLabels[key] ?? key}</span
              >
              <span class="line-through opacity-40 truncate max-w-[70px]"
                >{ch.from || tr("empty")}</span
              >
              <span class="opacity-30 shrink-0">→</span>
              <span class="text-text-primary font-semibold truncate"
                >{ch.to || tr("empty")}</span
              >
            </div>
          {/each}
        </div>
        <div class="flex gap-2">
          <button
            on:click={confirmApply}
            disabled={applying}
            class="flex-1 py-2.5 text-sm font-bold rounded-lg flex items-center justify-center gap-2
                   transition-all active:scale-[0.98] bg-accent hover:bg-accent/90
                   text-background shadow-lg shadow-accent/20 disabled:opacity-50"
          >
            {#if applying}
              <SpinnerIcon size={14} class="animate-spin" />
              {tr("saving")}
            {:else}
              <FloppyDiskIcon size={14} weight="bold" />
              {tr("confirmAndSave")}
            {/if}
          </button>
          <button
            on:click={() => (pendingConfirm = false)}
            disabled={applying}
            class="px-4 py-2 text-sm text-text-secondary hover:text-text-primary rounded-lg
                   hover:bg-white/5 transition-colors disabled:opacity-50"
          >
            {tr("cancel")}
          </button>
        </div>
      {:else}
        <!-- Normal footer -->
        <div class="flex gap-2">
          <button
            on:click={() => (pendingConfirm = true)}
            disabled={!hasChanges}
            class="w-full font-bold py-2.5 rounded-lg flex items-center justify-center gap-2
                   transition-all active:scale-[0.98] shadow-lg
                   {hasChanges
              ? 'bg-accent hover:bg-accent/90 text-background shadow-accent/20 cursor-pointer'
              : 'bg-accent/20 text-accent/50 shadow-none cursor-not-allowed'}"
          >
            {#if !hasChanges}
              <CheckCircleIcon size={16} weight="bold" />
              {tr("noChanges")}
            {:else}
              <FloppyDiskIcon size={16} weight="bold" />
              {tr("saveChanges")} ({Object.keys(changedFields).length})
            {/if}
          </button>
          {#if hasChanges}
            <button
              on:click={revertAll}
              class="shrink-0 px-3 py-2 text-text-muted hover:text-text-primary rounded-lg
                     hover:bg-white/5 transition-colors"
              title={tr("revertAll")}
            >
              <ArrowCounterClockwiseIcon size={16} />
            </button>
          {/if}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .field-label {
    display: block;
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--color-text-secondary);
    opacity: 0.7;
    margin-bottom: 0;
  }

  .field-input {
    width: 100%;
    background: var(--color-background);
    border: 1px solid var(--color-border);
    border-radius: 6px;
    padding: 6px 10px;
    font-size: 13px;
    outline: none;
    transition:
      border-color 0.15s,
      box-shadow 0.15s;
    color: var(--color-text-primary);
  }

  .field-input:focus {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px
      color-mix(in srgb, var(--color-accent) 15%, transparent);
  }
</style>
