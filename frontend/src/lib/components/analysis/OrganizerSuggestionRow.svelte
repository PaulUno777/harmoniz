<script lang="ts">
  import { CheckCircleIcon } from "phosphor-svelte";
  import type { OrganizerSuggestion } from "../../types";

  export let suggestion: OrganizerSuggestion;
  export let isSelected = false;
  export let onSelect: () => void;

  function scoreColor(score: number): string {
    if (score >= 75) return "bg-green-500";
    if (score >= 50) return "bg-yellow-500";
    if (score > 0) return "bg-yellow-500/50";
    return "bg-white/10";
  }

  $: hasFields = Object.keys(suggestion.fields).length > 0;
</script>

<button
  type="button"
  on:click={onSelect}
  class="w-full flex items-center gap-3 px-4 py-2.5 text-left transition-all
         hover:bg-surface/50
         {isSelected
    ? 'bg-surface border-l-2 border-l-accent'
    : 'border-l-2 border-l-transparent'}"
>
  <div
    class="w-2 h-2 rounded-full shrink-0 {scoreColor(suggestion.score)}"
  ></div>

  <div class="flex-1 min-w-0">
    {#if suggestion.fields["title"]}
      <div class="text-xs font-semibold text-accent truncate leading-tight">
        {suggestion.fields["title"].value}
      </div>
      <div
        class="text-[9px] text-text-muted/40 line-through truncate leading-tight"
      >
        {suggestion.track.title || suggestion.track.filename}
      </div>
    {:else}
      <div class="text-xs font-semibold text-text-primary truncate leading-tight">
        {suggestion.track.title || suggestion.track.filename}
      </div>
    {/if}

    {#if suggestion.fields["artist"]}
      <div class="flex items-center gap-1 mt-0.5 min-w-0">
        <span
          class="text-[9px] text-text-muted/40 line-through truncate shrink min-w-0 max-w-[40%]"
        >
          {suggestion.track.artist_raw || "—"}
        </span>
        <span class="text-[9px] text-text-muted/30 shrink-0">→</span>
        <span
          class="text-[9px] text-accent font-medium truncate shrink min-w-0"
        >
          {suggestion.fields["artist"].value}
        </span>
      </div>
    {:else}
      <div class="text-[10px] text-text-muted truncate mt-0.5">
        {suggestion.track.artist_raw || "—"}
      </div>
    {/if}
  </div>

  {#if hasFields}
    <span
      class="shrink-0 text-[9px] font-bold tabular-nums {suggestion.score >= 75
        ? 'text-green-400'
        : 'text-yellow-400'}"
    >
      {Math.round(suggestion.score)}%
    </span>
  {:else}
    <CheckCircleIcon size={12} class="text-green-400/60 shrink-0" />
  {/if}
</button>
