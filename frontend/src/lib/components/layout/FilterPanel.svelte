<script lang="ts">
  interface Props {
    isOpen: boolean;
    onClose: () => void;
    onApply: (filters: FilterState) => void;
    initialFilters: FilterState;
  }

  export interface FilterState {
    yearMin: number;
    yearMax: number;
    sizeMin: number;
    sizeMax: number;
  }

  let { isOpen, onClose, onApply, initialFilters }: Props = $props();

  let filters = $state<FilterState>({
    yearMin: 0,
    yearMax: 0,
    sizeMin: 0,
    sizeMax: 0,
  });

  $effect(() => {
    if (isOpen) {
      filters = { ...initialFilters };
    }
  });

  function handleApply() {
    onApply(filters);
    onClose();
  }

  function handleClear() {
    filters = {
      yearMin: 0,
      yearMax: 0,
      sizeMin: 0,
      sizeMax: 0,
    };
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return "0";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i)) + " " + sizes[i];
  }
</script>

{#if isOpen}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 bg-black/50 z-40 backdrop-blur-sm"
    onclick={onClose}
    role="presentation"
  ></div>

  <!-- Filter Panel -->
  <div
    class="fixed right-0 top-0 h-full w-96 bg-surface border-l border-border z-50 flex flex-col shadow-2xl"
  >
    <!-- Header -->
    <div class="p-6 border-b border-border">
      <h3 class="text-lg font-bold font-display">Advanced Filters</h3>
      <p class="text-xs text-text-muted mt-1">
        Refine your search with specific criteria
      </p>
    </div>

    <!-- Filter Content -->
    <div class="flex-1 overflow-y-auto p-6 space-y-6 custom-scrollbar">
      <!-- Year Range -->
      <div class="space-y-4">
        <h4 class="text-sm font-bold text-text-secondary uppercase tracking-wide">
          Year Range
        </h4>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label
              for="filter-year-min"
              class="block text-xs font-medium text-text-secondary mb-2"
            >
              From
            </label>
            <input
              id="filter-year-min"
              type="number"
              bind:value={filters.yearMin}
              placeholder="1900"
              min="0"
              class="w-full bg-background border border-border rounded-lg px-3 py-2 text-sm outline-none focus:ring-1 focus:ring-accent transition-all"
            />
          </div>
          <div>
            <label
              for="filter-year-max"
              class="block text-xs font-medium text-text-secondary mb-2"
            >
              To
            </label>
            <input
              id="filter-year-max"
              type="number"
              bind:value={filters.yearMax}
              placeholder="2024"
              min="0"
              class="w-full bg-background border border-border rounded-lg px-3 py-2 text-sm outline-none focus:ring-1 focus:ring-accent transition-all"
            />
          </div>
        </div>
      </div>

      <!-- File Size Range -->
      <div class="space-y-4">
        <h4 class="text-sm font-bold text-text-secondary uppercase tracking-wide">
          File Size Range
        </h4>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label
              for="filter-size-min"
              class="block text-xs font-medium text-text-secondary mb-2"
            >
              Min (MB)
            </label>
            <input
              id="filter-size-min"
              type="number"
              bind:value={filters.sizeMin}
              placeholder="0"
              min="0"
              step="0.1"
              class="w-full bg-background border border-border rounded-lg px-3 py-2 text-sm outline-none focus:ring-1 focus:ring-accent transition-all"
            />
          </div>
          <div>
            <label
              for="filter-size-max"
              class="block text-xs font-medium text-text-secondary mb-2"
            >
              Max (MB)
            </label>
            <input
              id="filter-size-max"
              type="number"
              bind:value={filters.sizeMax}
              placeholder="100"
              min="0"
              step="0.1"
              class="w-full bg-background border border-border rounded-lg px-3 py-2 text-sm outline-none focus:ring-1 focus:ring-accent transition-all"
            />
          </div>
        </div>

        {#if filters.sizeMin > 0 || filters.sizeMax > 0}
          <div class="text-xs text-text-muted">
            {#if filters.sizeMin > 0}
              Min: {formatBytes(filters.sizeMin * 1024 * 1024)}
            {/if}
            {#if filters.sizeMin > 0 && filters.sizeMax > 0}•{/if}
            {#if filters.sizeMax > 0}
              Max: {formatBytes(filters.sizeMax * 1024 * 1024)}
            {/if}
          </div>
        {/if}
      </div>
    </div>

    <!-- Footer Actions -->
    <div class="p-6 border-t border-border flex gap-3">
      <button
        onclick={handleClear}
        class="flex-1 px-4 py-2 bg-surface border border-border rounded-lg font-medium text-sm hover:bg-white/5 transition-all"
      >
        Clear All
      </button>
      <button
        onclick={handleApply}
        class="flex-1 px-4 py-2 bg-accent text-white rounded-lg font-medium text-sm hover:bg-accent/90 transition-all"
      >
        Apply Filters
      </button>
    </div>
  </div>
{/if}
