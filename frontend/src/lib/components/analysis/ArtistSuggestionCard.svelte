<script lang="ts">
  import { t } from "../../stores/i18n";
  import type { ArtistSuggestion } from "../../types";

  interface Props {
    suggestion: ArtistSuggestion;
  }

  let { suggestion }: Props = $props();

  function onApply() {
    // Phase 5 (Organize) will implement apply
    console.log("Apply suggestion (Phase 5):", suggestion.original, "->", suggestion.suggested);
  }

  function onDismiss() {
    console.log("Dismiss suggestion:", suggestion.original);
  }
</script>

<div
  class="rounded-xl border border-border bg-surface/50 p-4 hover:border-border hover:bg-surface/70 transition-all"
>
  <div class="flex items-start justify-between gap-4">
    <div class="min-w-0 flex-1">
      <div class="flex items-center gap-2 flex-wrap">
        <span class="text-text-secondary font-medium truncate" title={suggestion.original}>
          {suggestion.original}
        </span>
        <span class="text-text-muted">→</span>
        <span class="text-text-primary font-semibold truncate" title={suggestion.suggested}>
          {suggestion.suggested}
        </span>
      </div>
      <div class="flex items-center gap-2 mt-1.5">
        <span class="text-xs text-text-muted">
          {Math.round(suggestion.score * 100)}% match
        </span>
        <span
          class="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded bg-white/10 text-text-muted"
        >
          {suggestion.confidence_level}
        </span>
      </div>
    </div>
    <div class="flex items-center gap-2 shrink-0">
      <button
        type="button"
        onclick={onApply}
        class="px-2.5 py-1 text-xs font-medium bg-accent/20 text-accent rounded-lg hover:bg-accent/30 transition-colors"
      >
        {$t("useSuggested")}
      </button>
      <button
        type="button"
        onclick={onDismiss}
        class="px-2.5 py-1 text-xs font-medium text-text-muted hover:text-text-primary hover:bg-white/5 rounded-lg transition-colors"
      >
        {$t("dismiss")}
      </button>
    </div>
  </div>
</div>
