<script lang="ts">
  import type { OrganizerSuggestion } from "../../types";
  import OrganizerSuggestionRow from "./OrganizerSuggestionRow.svelte";

  export let items: OrganizerSuggestion[] = [];
  export let selectedTrackId: number | null = null;
  export let onToggleSelect: (trackId: number) => void;
</script>

{#each items as sugg (sugg.track_id)}
  <OrganizerSuggestionRow
    suggestion={sugg}
    isSelected={selectedTrackId === sugg.track_id}
    onSelect={() => onToggleSelect(sugg.track_id)}
  />
{/each}
