<script lang="ts">
  import { t } from "../../stores/i18n";
  import { analysis } from "../../stores/analysis";
  import ArtistSuggestionCard from "./ArtistSuggestionCard.svelte";

  interface Props {
    onBrowse: () => void;
  }

  let { onBrowse }: Props = $props();

  const loading = analysis.loading;
  const error = analysis.error;
  const artistSuggestions = analysis.artistSuggestions;
  const hasRunOnce = analysis.hasRunOnce;
</script>

<div class="h-full flex flex-col min-h-0 overflow-hidden">
  <div class="shrink-0 px-8 py-4 border-b border-border">
    <h3 class="text-base font-bold text-text-primary">{$t("artistSuggestions")}</h3>
    <p class="text-xs text-text-muted mt-0.5">{$t("artistSuggestionsDescription")}</p>
  </div>

  <div class="flex-1 overflow-y-auto p-8">
    {#if $loading}
      <div class="flex flex-col items-center justify-center py-16 text-text-muted">
        <div class="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin mb-4"></div>
        <p class="text-sm">{$t("analyzing")}</p>
      </div>
    {:else if $error}
      <div class="rounded-xl border border-red-500/30 bg-red-500/10 p-4 text-red-200 text-sm">
        <p>{$error}</p>
      </div>
    {:else if !$artistSuggestions.length && !$loading}
      <div class="flex flex-col items-center justify-center py-16 text-center">
        {#if !$hasRunOnce && !$error}
          <p class="text-text-secondary mb-2">{$t("runAnalysisToSeeSuggestions")}</p>
          <p class="text-xs text-text-muted mb-4">{$t("selectLibraryToAnalyze")}</p>
          <button
            type="button"
            onclick={onBrowse}
            class="px-4 py-2 bg-accent/20 text-accent rounded-lg font-medium text-sm hover:bg-accent/30 transition-colors"
          >
            {$t("browse")}
          </button>
        {:else}
          <p class="text-text-secondary">{$t("noSimilarArtistNamesFound")}</p>
        {/if}
      </div>
    {:else}
      <div class="space-y-3 max-w-3xl">
        {#each $artistSuggestions as suggestion (suggestion.original + suggestion.suggested)}
          <ArtistSuggestionCard {suggestion} />
        {/each}
      </div>
    {/if}
  </div>
</div>
