<script lang="ts">
  import {
    GlobeIcon,
    MoonIcon,
    SunIcon,
    MonitorIcon,
    TranslateIcon,
  } from 'phosphor-svelte'
  import { t } from '../../stores/i18n'
  import { locale, setLocale, type Locale } from '../../stores/i18n'
  import { theme } from '../../stores/theme'
  import { updateStore } from '../../stores/update'
  import CustomSelect from '../ui/CustomSelect.svelte'
  import ThemePreviewButton from '../ui/ThemePreviewButton.svelte'
  import logoRaw from '../../../assets/images/hamonizer_logo.svg?raw'

  const languages: { id: Locale; label: string }[] = [
    { id: 'en', label: 'English' },
    { id: 'fr', label: 'Français' },
  ]

  async function handleCheckUpdates() {
    await updateStore.check()
  }

  function formatLastChecked(iso: string | null): string {
    if (!iso) return ''
    try {
      return new Date(iso).toLocaleString()
    } catch {
      return ''
    }
  }

  const themeDescription = $derived(() => {
    if ($theme === 'system') {
      return $t('matchingSystemPreference')
    }
    return $theme === 'dark' ? $t('currentlyUsingDark') : $t('currentlyUsingLight')
  })

  const versionLabel = $derived(
    $updateStore.currentVersion ? `v${$updateStore.currentVersion}` : 'v—',
  )
</script>

<div class="flex-1 p-8 max-w-4xl mx-auto w-full">
  <div class="mb-8">
    <h1 class="text-2xl font-bold mb-1 font-display">{$t('settings')}</h1>
    <p class="text-text-secondary text-sm">{$t('settingsDescription')}</p>
  </div>

  <div class="space-y-8 pb-12">
    <!-- Appearance -->
    <section>
      <div class="flex items-center gap-3 mb-4">
        <SunIcon size={20} weight="bold" class="text-accent" />
        <h2 class="text-md font-bold font-display">{$t('appearance')}</h2>
      </div>

      <div class="bg-surface rounded-2xl border border-border p-5 shadow-sm space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-start justify-items-center max-w-[600px] mx-auto">
          <ThemePreviewButton
            value="dark"
            selectedValue={$theme}
            label={$t('dark')}
            icon={MoonIcon}
            previewBg="#0f1115"
            previewBorder="#2c2e36"
            previewIconClass="text-[#eaeaeb]"
            onClick={() => $theme = 'dark'}
          />

          <ThemePreviewButton
            value="light"
            selectedValue={$theme}
            label={$t('light')}
            icon={SunIcon}
            previewBg="#f8fafc"
            previewBorder="#e2e8f0"
            previewIconClass="text-[#0f172a]"
            onClick={() => $theme = 'light'}
          />

          <ThemePreviewButton
            value="system"
            selectedValue={$theme}
            label={$t('system')}
            icon={MonitorIcon}
            previewBg="linear-gradient(to bottom right, #0f1115, #f8fafc)"
            previewBorder="var(--border)"
            previewIconClass="text-accent mix-blend-difference relative z-10"
            isSystem={true}
            onClick={() => $theme = 'system'}
          />
        </div>

        <p class="text-[10px] text-text-muted px-1 leading-relaxed">
          {themeDescription()}
        </p>
      </div>
    </section>

    <!-- Localization -->
    <section>
      <div class="flex items-center gap-3 mb-4">
        <TranslateIcon size={20} weight="bold" class="text-accent" />
        <h2 class="text-md font-bold font-display">{$t('language')}</h2>
      </div>

      <div class="bg-surface rounded-2xl border border-border p-5 flex items-center justify-between shadow-sm">
        <div class="flex items-center gap-4">
          <div class="w-10 h-10 bg-accent/5 rounded-xl flex items-center justify-center text-accent">
            <GlobeIcon size={20} weight="duotone" />
          </div>
          <div>
            <div class="font-bold text-sm">{$t('appLanguage')}</div>
            <div class="text-[11px] text-text-secondary">{$t('changeLanguage')}</div>
          </div>
        </div>

        <div class="w-40">
          <CustomSelect
            options={languages}
            value={$locale}
            onSelect={(id) => setLocale(id as Locale)}
          />
        </div>
      </div>
    </section>

    <!-- About -->
    <section class="pt-8 border-t border-border/50">
      <div
        class="rounded-2xl p-5 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border shadow-sm
          {$updateStore.updateAvailable
            ? 'bg-amber-500/5 border-amber-500/25'
            : 'bg-accent/5 border-accent/20'}"
      >
        <div class="flex items-center gap-4 min-w-0">
          <div class="w-10 h-10 text-accent shrink-0 flex items-center justify-center">
            {@html logoRaw}
          </div>
          <div class="min-w-0">
            <div class="font-bold text-accent tracking-tight font-display text-md truncate">
              Harmoniz {versionLabel}
            </div>
            <div class="text-[11px] text-text-secondary">{$t('aboutDescription')}</div>
            {#if $updateStore.lastCheckedAt}
              <div class="text-[10px] text-text-muted mt-1">
                {$t('lastChecked')}: {formatLastChecked($updateStore.lastCheckedAt)}
              </div>
            {/if}
          </div>
        </div>

        <div class="flex flex-col items-stretch sm:items-end gap-2 shrink-0">
          {#if $updateStore.updateAvailable}
            <p class="text-[11px] font-semibold text-amber-500/90 text-center sm:text-right">
              {$t('updateAvailable').replace('{version}', $updateStore.latestVersion)}
            </p>
            <div class="flex flex-wrap gap-2 justify-end">
              <button
                onclick={() => updateStore.openDownload()}
                class="px-5 py-2 bg-accent text-background font-bold text-xs rounded-lg hover:bg-accent-hover transition-all active:scale-[0.98] shadow-lg shadow-accent/20"
              >
                {$t('downloadUpdate')}
              </button>
              <button
                onclick={() => updateStore.dismissUpdate()}
                class="px-4 py-2 text-xs font-semibold text-text-secondary hover:text-text-primary rounded-lg hover:bg-white/5 transition-colors"
              >
                {$t('dismissUpdate')}
              </button>
            </div>
          {:else if $updateStore.status === 'error'}
            <span class="text-[10px] font-semibold text-red-400/90 uppercase tracking-wide text-center sm:text-right">
              {$t('updateCheckFailed')}
            </span>
          {:else if $updateStore.status === 'ready'}
            <span class="text-[10px] font-bold text-emerald-500 uppercase tracking-widest text-center sm:text-right">
              {$t('latestVersion')}
            </span>
          {/if}

          <button
            onclick={handleCheckUpdates}
            disabled={$updateStore.status === 'checking'}
            class="px-5 py-2 bg-surface border border-border text-text-primary font-bold text-xs rounded-lg hover:bg-white/5 transition-all active:scale-[0.98] disabled:opacity-50 disabled:cursor-wait"
          >
            {$updateStore.status === 'checking' ? $t('checking') : $t('checkForUpdates')}
          </button>
        </div>
      </div>
    </section>
  </div>
</div>
