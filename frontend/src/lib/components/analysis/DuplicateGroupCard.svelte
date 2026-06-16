<script lang="ts">
  import { formatFileSize } from "../../utils/format";
  import { t } from "../../stores/i18n";
  import type { Track } from "../../types";

  interface Props {
    group: Track[];
    onResolve?: (trackId: number | undefined, action: string) => void;
  }

  let { group, onResolve }: Props = $props();
  let expanded = $state(true);

  const first = $derived.by(() => group[0]);
  const saveBytes = $derived.by(() => {
    const f = group[0];
    return f && group.length > 1 ? (group.length - 1) * (f.size ?? 0) : 0;
  });
</script>

<div class="rounded-xl border border-border bg-surface/50 overflow-hidden mb-4">
  <button
    type="button"
    class="w-full flex items-center justify-between p-4 text-left hover:bg-white/5 transition-colors"
    onclick={() => (expanded = !expanded)}
  >
    <div class="flex items-center gap-3 min-w-0">
      <span class="text-text-primary font-medium truncate" title={first?.filename ?? first?.path}>
        {first?.filename ?? first?.title ?? first?.path ?? "—"}
      </span>
      <span class="text-xs text-text-muted shrink-0">
        {$t("andNDuplicates").replace("{n}", String(group.length - 1))}
      </span>
      {#if saveBytes > 0}
        <span class="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-accent/20 text-accent">
          {$t("saveMB").replace("{size}", formatFileSize(saveBytes))}
        </span>
      {/if}
    </div>
    <span class="text-text-muted text-xs shrink-0 ml-2">
      {#if first?.hash_partial}
        {first.hash_partial.slice(0, 8)}…
      {/if}
    </span>
  </button>

  {#if expanded}
    <div class="border-t border-border/50">
      {#each group as track (track.path)}
        <div
          class="flex items-center gap-4 p-3 border-b border-border/10 last:border-0 hover:bg-white/5 transition-colors"
        >
          <div class="flex-1 min-w-0 grid grid-cols-12 gap-4 items-center">
            <div class="col-span-6 truncate text-text-primary text-xs font-mono" title={track.path}>
              {track.path}
            </div>
            <div class="col-span-2 text-text-secondary text-xs">
              {formatFileSize(track.size ?? 0)}
            </div>
          </div>
          <button
            type="button"
            class="px-3 py-1.5 text-xs font-medium text-red-400 bg-red-400/10 border border-red-400/20 rounded hover:bg-red-400/20 transition-colors shrink-0"
            onclick={(e) => {
              e.stopPropagation();
              onResolve?.(track.id as number, "delete");
            }}
          >
            {$t("delete")}
          </button>
        </div>
      {/each}
    </div>
  {/if}
</div>
