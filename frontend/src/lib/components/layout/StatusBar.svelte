<script lang="ts">
  import { MusicNotesIcon, FileIcon } from "phosphor-svelte";
  import { t } from "../../stores/i18n";
  import { updateStore, showUpdatePrompt } from "../../stores/update";

  interface Props {
    tracked: number;
    hasLibrary: boolean;
    selectedFileName?: string | null;
    onOpenSettings: () => void;
  }

  let {
    tracked = 0,
    hasLibrary = false,
    selectedFileName = null,
    onOpenSettings,
  }: Props = $props();

  function handleUpdateClick() {
    updateStore.openDownload();
  }
</script>

<footer
  class="h-8 border-t border-border bg-surface px-4 flex items-center justify-between text-[10px] font-bold uppercase tracking-widest text-text-secondary select-none"
>
  <div class="flex items-center gap-4 min-w-0">
    {#if hasLibrary}
      <div class="flex items-center gap-2 shrink-0">
        <MusicNotesIcon size={12} class="text-accent shrink-0" />
        <span class="text-[9px] normal-case opacity-90">
          <span class="text-accent font-mono">{tracked}</span>
          <span class="opacity-70 ml-1">{$t("statusTracked")}</span>
        </span>
      </div>
      {#if selectedFileName}
        <div class="w-px h-3 bg-border shrink-0"></div>
        <div class="flex items-center gap-2 min-w-0">
          <FileIcon size={12} class="text-text-muted shrink-0" />
          <span
            class="text-[9px] normal-case opacity-90 truncate max-w-64 font-mono"
            title={selectedFileName}
          >
            {selectedFileName}
          </span>
        </div>
      {/if}
    {:else}
      <span class="opacity-50 text-[9px] normal-case"
        >{$t("noLibrarySelected")}</span
      >
    {/if}
  </div>

  <div
    class="flex items-center gap-2 shrink-0 text-[9px] uppercase tracking-widest font-bold opacity-60"
  >
    <button
      type="button"
      onclick={onOpenSettings}
      class="flex items-center gap-1.5 min-w-0 hover:opacity-90 transition-opacity"
    >
      {#if $showUpdatePrompt}
        <span
          class="w-1.5 h-1.5 rounded-full bg-accent shrink-0"
          aria-hidden="true"
        ></span>
      {/if}
      <span class="truncate"
        >{$t("appName")}
        {#if $updateStore.currentVersion}
          v{$updateStore.currentVersion}
        {:else}
          v—
        {/if}
      </span>
    </button>
    {#if $showUpdatePrompt}
      <button
        type="button"
        onclick={handleUpdateClick}
        class="shrink-0 px-2 py-0.5 rounded-full bg-accent text-background text-[8px] font-bold tracking-wide normal-case hover:bg-accent-hover transition-colors"
      >
        {$t("update")}
      </button>
    {/if}
  </div>
</footer>
