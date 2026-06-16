<script lang="ts">
  import {
    CheckIcon,
    XIcon,
    SpinnerIcon,
    FileAudioIcon,
  } from "phosphor-svelte";
  import type { OrganizerSuggestion } from "../../types";
  import { t } from "../../stores/i18n";

  // Svelte 4 style props (no runes)
  export let suggestion: OrganizerSuggestion;
  export let x: number;
  export let y: number;
  export let filenameTemplate: string;
  export let onClose: () => void;
  export let onApply: (
    trackID: number,
    fields: Record<string, string>,
  ) => Promise<void>;

  let applying = false;

  const fieldLabels: Record<string, string> = {
    artist: "Artist",
    album: "Album",
    title: "Title",
    year: "Year",
    track_num: "Track #",
  };

  const fieldOrder = ["title", "artist", "album", "year", "track_num"];

  $: sortedFields = fieldOrder
    .filter((k) => k in suggestion.fields)
    .map((k) => [k, suggestion.fields[k]] as const);

  function confidenceClass(conf: number) {
    if (conf >= 0.75) return "text-green-400";
    if (conf >= 0.5) return "text-yellow-400";
    return "text-text-muted";
  }

  function sourceLabel(source: string) {
    if (source === "path") return $t("suggestionSource_path" as any);
    if (source === "filename") return $t("suggestionSource_filename" as any);
    return $t("suggestionSource_neighbor" as any);
  }

  function currentValue(key: string): string {
    const tr = suggestion.track;
    if (key === "artist") return tr.artist_raw ?? "";
    if (key === "album") return tr.album_raw ?? "";
    if (key === "title") return tr.title ?? "";
    if (key === "year") return tr.year > 0 ? String(tr.year) : "";
    if (key === "track_num") return tr.track_num ? String(tr.track_num) : "";
    return "";
  }

  async function handleApplyAll() {
    if (applying) return;
    applying = true;
    try {
      const toApply: Record<string, string> = {};
      for (const [key, field] of Object.entries(suggestion.fields)) {
        toApply[key] = field.value;
      }
      await onApply(suggestion.track_id, toApply);
      onClose();
    } finally {
      applying = false;
    }
  }

  async function handleApplySingle(key: string, value: string) {
    if (applying) return;
    applying = true;
    try {
      await onApply(suggestion.track_id, { [key]: value });
    } finally {
      applying = false;
    }
  }

  // Close on click outside or Escape
  let menuEl: HTMLElement;

  function handleWindowClick(e: MouseEvent) {
    if (menuEl && !menuEl.contains(e.target as Node)) onClose();
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") onClose();
  }
</script>

<svelte:window on:mousedown={handleWindowClick} on:keydown={handleKeyDown} />

<div
  bind:this={menuEl}
  class="fixed z-50 w-80 bg-surface border border-border rounded-xl shadow-2xl overflow-hidden"
  style="left: {Math.min(x, window.innerWidth - 340)}px; top: {Math.min(
    y,
    window.innerHeight - 440,
  )}px"
>
  <!-- Header -->
  <div
    class="px-4 py-3 bg-background/60 border-b border-border flex items-center gap-2"
  >
    <FileAudioIcon size={14} class="text-accent shrink-0" />
    <span
      class="text-xs font-mono text-text-secondary truncate flex-1"
      title={suggestion.track.filename}
    >
      {suggestion.track.filename}
    </span>
    <button
      on:click={onClose}
      class="text-text-muted hover:text-text-primary transition-colors shrink-0"
    >
      <XIcon size={14} />
    </button>
  </div>

  <!-- Score badge -->
  <div class="px-4 py-2 border-b border-border/50 flex items-center gap-2">
    <span
      class="text-[10px] font-bold uppercase tracking-wider text-text-muted"
    >
      {$t("confidence" as any)}
    </span>
    <span
      class="ml-auto text-xs font-bold tabular-nums {suggestion.score >= 75
        ? 'text-green-400'
        : suggestion.score >= 50
          ? 'text-yellow-400'
          : 'text-text-muted'}"
    >
      {Math.round(suggestion.score)}%
    </span>
  </div>

  <!-- Per-field suggestions -->
  {#if sortedFields.length === 0}
    <div class="px-4 py-6 text-center text-xs text-text-muted">
      {$t("trackComplete" as any)}
    </div>
  {:else}
    <div class="divide-y divide-border/40 max-h-64 overflow-y-auto">
      {#each sortedFields as [key, field]}
        <div class="px-4 py-2.5">
          <div class="flex items-center justify-between mb-1">
            <span
              class="text-[10px] font-bold uppercase tracking-wider text-text-secondary/70"
            >
              {fieldLabels[key] ?? key}
            </span>
            <div class="flex items-center gap-1.5">
              <span
                class="text-[9px] px-1.5 py-0.5 rounded-full bg-white/5 text-text-muted uppercase"
              >
                {sourceLabel(field.source)}
              </span>
              <span
                class="text-[10px] font-bold tabular-nums {confidenceClass(
                  field.confidence,
                )}"
              >
                {Math.round(field.confidence * 100)}%
              </span>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <div class="flex-1 min-w-0">
              {#if currentValue(key)}
                <div
                  class="text-[10px] text-text-muted truncate line-through opacity-50"
                >
                  {currentValue(key)}
                </div>
              {/if}
              <div class="text-xs font-semibold text-text-primary truncate">
                {field.value}
              </div>
            </div>
            <button
              on:click={() => handleApplySingle(key, field.value)}
              disabled={applying}
              class="shrink-0 p-1.5 rounded-md bg-accent/10 text-accent hover:bg-accent/20 transition-colors disabled:opacity-40"
              title="Apply this field"
            >
              <CheckIcon size={12} weight="bold" />
            </button>
          </div>
        </div>
      {/each}
    </div>

    <!-- Template preview -->
    {#if filenameTemplate && filenameTemplate.trim()}
      <div class="px-4 py-2 border-t border-border/50 bg-background/30">
        <div
          class="text-[10px] font-bold uppercase tracking-wider text-text-muted mb-1"
        >
          {$t("filenameTemplate" as any)}
        </div>
        <div class="text-[11px] font-mono text-text-secondary/70 truncate">
          {filenameTemplate}
        </div>
      </div>
    {/if}

    <!-- Footer actions -->
    <div class="px-3 py-2.5 border-t border-border flex gap-2">
      <button
        on:click={handleApplyAll}
        disabled={applying}
        class="flex-1 py-2 px-3 bg-accent text-background text-xs font-bold rounded-lg flex items-center justify-center gap-1.5
               hover:bg-accent/90 active:scale-[0.98] transition-all disabled:opacity-50"
      >
        {#if applying}
          <SpinnerIcon size={12} class="animate-spin" />
          {$t("applying" as any)}
        {:else}
          <CheckIcon size={12} weight="bold" />
          {$t("applyAllForFile" as any)}
        {/if}
      </button>
      <button
        on:click={onClose}
        class="px-3 py-2 text-xs text-text-secondary hover:text-text-primary rounded-lg hover:bg-white/5 transition-colors"
      >
        {$t("dismiss")}
      </button>
    </div>
  {/if}
</div>
