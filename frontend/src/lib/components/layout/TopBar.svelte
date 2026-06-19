<script lang="ts">
  import {
    MagnifyingGlassIcon,
    FunnelIcon,
    ListIcon,
    FolderIcon,
  } from "phosphor-svelte";
  import { t } from "../../stores/i18n";
  import type { TabId } from "../../types";

  interface Props {
    activeTab: TabId;
    onBrowse: () => void;
    searchQuery: string;
    onSearchChange: (query: string) => void;
    onFilterToggle: () => void;
    activeFilterCount: number;
    onConfigToggle: () => void;
    onRunAnalysis?: () => void;
    analysisLoading?: boolean;
    hasLibrary?: boolean;
  }

  let {
    activeTab,
    onBrowse,
    searchQuery,
    onSearchChange,
    onFilterToggle,
    activeFilterCount,
    onConfigToggle,
  }: Props = $props();
</script>

<header
  class="h-16 border-b border-border flex items-center px-8 justify-between gap-8 bg-background/80 backdrop-blur-md z-10 sticky top-0"
>
  <div class="flex items-center gap-4 flex-1">
    <h2 class="text-lg font-bold capitalize font-display">
      {$t(activeTab)}
    </h2>
    <div class="h-4 w-px bg-border"></div>

    <div class="flex items-center gap-3 flex-1">
      <button
        onclick={onBrowse}
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
          value={searchQuery}
          oninput={(e) => onSearchChange((e.target as HTMLInputElement).value)}
          placeholder={$t("search")}
          class="w-full bg-surface border border-border rounded-full py-2 pl-10 pr-4 outline-none focus:ring-1 focus:ring-accent transition-all"
        />
      </div>
    </div>
  </div>

  <div class="flex items-center gap-2">
    <button
      onclick={onFilterToggle}
      class="p-2 hover:bg-white/5 rounded-lg text-text-secondary transition-colors relative"
      title={$t("filter")}
    >
      <FunnelIcon size={20} />
      {#if activeFilterCount > 0}
        <span
          class="absolute -top-1 -right-1 bg-accent text-white text-[10px] font-bold rounded-full w-4 h-4 flex items-center justify-center"
        >
          {activeFilterCount}
        </span>
      {/if}
    </button>
    <button
      onclick={onConfigToggle}
      class="p-2 hover:bg-white/5 rounded-lg text-text-secondary transition-colors"
      title={$t("algorithmPrecision")}
    >
      <ListIcon size={20} />
    </button>
  </div>
</header>
