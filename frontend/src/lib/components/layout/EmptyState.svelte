<script lang="ts">
  import { t } from "../../stores/i18n";
  import logoRaw from '../../../assets/images/hamonizer_logo.svg?raw'

  interface Props {
    onBrowse: () => void;
  }

  let { onBrowse }: Props = $props();
  let isDragOver = $state(false);
</script>

<div class="h-full flex flex-col items-center justify-center p-12 text-center">
  <div
    class="w-full max-w-lg border-2 border-dashed rounded-3xl flex flex-col items-center justify-center gap-6 py-14 px-10 transition-all duration-500
           {isDragOver
      ? 'border-accent bg-accent/5 scale-105 shadow-2xl shadow-accent/10'
      : 'border-border bg-surface/30 hover:border-text-muted hover:bg-surface/50'}"
    role="button"
    tabindex="0"
    data-file-drop-target
    style="--wails-drop-target: drop"
    ondragover={(e) => { e.preventDefault(); isDragOver = true; }}
    ondragleave={() => { isDragOver = false; }}
    ondrop={(e) => { e.preventDefault(); isDragOver = false; }}
  >
    <!-- Logo -->
    <div
      class="w-20 h-20 text-accent transition-all duration-500
             {isDragOver ? 'scale-110 drop-shadow-[0_0_20px_var(--color-accent)]' : 'opacity-80'}"
    >
      {@html logoRaw}
    </div>

    <div>
      <h3 class="text-xl font-bold font-display mb-2">{$t('emptyStateTitle')}</h3>
      <p class="text-text-secondary text-sm max-w-xs mx-auto leading-relaxed">
        {$t('emptyStateSubtitle')}
      </p>
    </div>

    <div class="flex items-center gap-3">
      <button
        onclick={onBrowse}
        class="px-8 py-2.5 bg-accent text-background font-bold rounded-xl hover:scale-105 transition-all shadow-lg shadow-accent/20 active:scale-95"
      >
        {$t('openFolder')}
      </button>
    </div>

    <span class="text-[10px] uppercase tracking-widest font-bold text-text-muted opacity-50">
      {$t('dropZone')}
    </span>
  </div>
</div>
