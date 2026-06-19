<script lang="ts">
  import { SlidersIcon, XIcon, SparkleIcon, BroomIcon, ArrowCounterClockwiseIcon } from 'phosphor-svelte'
  import { t } from '../../stores/i18n'
  import { appConfig } from '../../stores/settings'
  import type { domain } from '../../../../wailsjs/go/models'

  interface Props {
    isOpen: boolean
    onClose: () => void
  }

  let { isOpen, onClose }: Props = $props()

  type AppConfig = domain.AppConfig

  // Local display state — updated instantly on oninput for smooth slider feedback.
  // Backend is only called on onchange (user releases the slider).
  let localSimilarity = $state(0.88)
  let localArtist = $state(0.90)
  let localField = $state(0.75)

  // Sync local state whenever the store loads or external reset happens.
  $effect(() => {
    if ($appConfig) {
      localSimilarity = $appConfig.cleaner.similarity_threshold
      localArtist = $appConfig.cleaner.artist_group_threshold
      localField = $appConfig.organizer.field_confidence_threshold
    }
  })

  function pct(v: number): string {
    return Math.round(v * 100) + '%'
  }

  // oninput handlers — update only local display state while dragging
  function inputSimilarity(e: Event) {
    localSimilarity = parseFloat((e.target as HTMLInputElement).value)
  }
  function inputArtist(e: Event) {
    localArtist = parseFloat((e.target as HTMLInputElement).value)
  }
  function inputField(e: Event) {
    localField = parseFloat((e.target as HTMLInputElement).value)
  }

  // onchange handlers — save to backend when user releases the slider
  function saveSimilarity() {
    if (!$appConfig) return
    void appConfig.save({
      ...$appConfig,
      cleaner: { ...$appConfig.cleaner, similarity_threshold: localSimilarity },
    } as AppConfig)
  }
  function saveArtist() {
    if (!$appConfig) return
    void appConfig.save({
      ...$appConfig,
      cleaner: { ...$appConfig.cleaner, artist_group_threshold: localArtist },
    } as AppConfig)
  }
  function saveField() {
    if (!$appConfig) return
    void appConfig.save({
      ...$appConfig,
      organizer: { field_confidence_threshold: localField },
    } as AppConfig)
  }
</script>

{#if isOpen}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 bg-black/50 z-40 backdrop-blur-sm"
    onclick={onClose}
    role="presentation"
  ></div>

  <!-- Panel -->
  <div
    class="fixed right-0 top-0 h-full w-96 bg-surface border-l border-border z-50 flex flex-col shadow-2xl"
  >
    <!-- Header -->
    <div class="p-6 border-b border-border flex items-center justify-between">
      <div class="flex items-center gap-2">
        <SlidersIcon size={18} weight="bold" class="text-accent" />
        <h3 class="text-base font-bold font-display">{$t('algorithmPrecision')}</h3>
      </div>
      <button
        onclick={onClose}
        class="p-1.5 hover:bg-white/5 rounded-lg text-text-secondary transition-colors"
      >
        <XIcon size={18} />
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-6 space-y-8">
      {#if $appConfig}

        <!-- Cleaner section -->
        <section>
          <div class="flex items-center gap-2 mb-4">
            <BroomIcon size={16} weight="bold" class="text-accent" />
            <h4 class="text-sm font-bold font-display">{$t('cleanerPrecision')}</h4>
          </div>

          <div class="space-y-6">
            <!-- Duplicate similarity -->
            <div>
              <div class="flex justify-between items-baseline mb-1">
                <span class="text-xs font-semibold text-text-primary">{$t('duplicateSimilarity')}</span>
                <span class="text-xs font-bold text-accent tabular-nums">{pct(localSimilarity)}</span>
              </div>
              <p class="text-[10px] text-text-muted mb-2">{$t('duplicateSimilarityDesc')}</p>
              <input
                type="range"
                min="0.70"
                max="1.00"
                step="0.01"
                value={localSimilarity}
                oninput={inputSimilarity}
                onchange={saveSimilarity}
                class="w-full accent-accent"
              />
              <div class="flex justify-between text-[9px] text-text-muted mt-0.5">
                <span>{$t('loose')}</span>
                <span>{$t('strict')}</span>
              </div>
            </div>

            <!-- Artist grouping -->
            <div>
              <div class="flex justify-between items-baseline mb-1">
                <span class="text-xs font-semibold text-text-primary">{$t('artistGrouping')}</span>
                <span class="text-xs font-bold text-accent tabular-nums">{pct(localArtist)}</span>
              </div>
              <p class="text-[10px] text-text-muted mb-2">{$t('artistGroupingDesc')}</p>
              <input
                type="range"
                min="0.70"
                max="1.00"
                step="0.01"
                value={localArtist}
                oninput={inputArtist}
                onchange={saveArtist}
                class="w-full accent-accent"
              />
              <div class="flex justify-between text-[9px] text-text-muted mt-0.5">
                <span>{$t('loose')}</span>
                <span>{$t('strict')}</span>
              </div>
            </div>
          </div>
        </section>

        <!-- Organizer section -->
        <section>
          <div class="flex items-center gap-2 mb-4">
            <SparkleIcon size={16} weight="bold" class="text-accent" />
            <h4 class="text-sm font-bold font-display">{$t('organizerPrecision')}</h4>
          </div>

          <div>
            <div class="flex justify-between items-baseline mb-1">
              <span class="text-xs font-semibold text-text-primary">{$t('fieldConfidence')}</span>
              <span class="text-xs font-bold text-accent tabular-nums">{pct(localField)}</span>
            </div>
            <p class="text-[10px] text-text-muted mb-2">{$t('fieldConfidenceDesc')}</p>
            <input
              type="range"
              min="0.50"
              max="1.00"
              step="0.01"
              value={localField}
              oninput={inputField}
              onchange={saveField}
              class="w-full accent-accent"
            />
            <div class="flex justify-between text-[9px] text-text-muted mt-0.5">
              <span>{$t('loose')}</span>
              <span>{$t('strict')}</span>
            </div>
          </div>
        </section>

      {:else}
        <div class="flex items-center justify-center h-32 text-text-muted text-xs">
          Loading…
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-border">
      <button
        onclick={() => void appConfig.resetDefaults()}
        class="w-full flex items-center justify-center gap-2 px-4 py-2 text-xs font-semibold text-text-secondary hover:text-text-primary hover:bg-white/5 rounded-lg transition-colors"
      >
        <ArrowCounterClockwiseIcon size={14} />
        {$t('resetDefaults')}
      </button>
    </div>
  </div>
{/if}
