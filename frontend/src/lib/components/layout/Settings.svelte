<script lang="ts">
  import { 
    GlobeIcon, 
    MoonIcon, 
    SunIcon, 
    MonitorIcon, 
    TranslateIcon, 
    MusicNotesIcon 
  } from 'phosphor-svelte'
  import { t, locale, type Locale } from '../../stores/i18n'
  import { theme } from '../../stores/theme'
  import CustomSelect from '../ui/CustomSelect.svelte'
  import ThemePreviewButton from '../ui/ThemePreviewButton.svelte'

  let isCheckingUpdates = $state(false)
  let updateStatusKey = $state<'latestVersion' | null>(null)
  
  const languages: { id: Locale; label: string }[] = [
    { id: 'en', label: 'English' },
    { id: 'fr', label: 'Français' },
  ]

  async function handleCheckUpdates() {
    isCheckingUpdates = true
    updateStatusKey = null
    
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    isCheckingUpdates = false
    updateStatusKey = 'latestVersion'
  }

  const themeDescription = $derived(() => {
    if ($theme === 'system') {
      return $t('matchingSystemPreference')
    }
    return $theme === 'dark' ? $t('currentlyUsingDark') : $t('currentlyUsingLight')
  })
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
          <!-- Dark Preview -->
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

          <!-- Light Preview -->
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

          <!-- System Preview -->
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
            bind:value={$locale}
          />
        </div>
      </div>
    </section>

    <!-- About -->
    <section class="pt-8 border-t border-border/50">
      <div class="bg-accent/5 border border-accent/20 rounded-2xl p-5 flex items-center justify-between group">
        <div class="flex items-center gap-4">
          <div class="w-10 h-10 bg-accent rounded-xl flex items-center justify-center shadow-lg shadow-accent/20 group-hover:scale-110 transition-transform">
            <MusicNotesIcon size={20} weight="bold" class="text-background" />
          </div>
          <div>
            <div class="font-bold text-accent italic tracking-tight font-display text-md">Harmonizr v0.1.0-alpha</div>
            <div class="text-[11px] text-text-secondary">{$t('aboutDescription')}</div>
          </div>
        </div>
        
        <div class="flex flex-col items-end gap-1.5">
          <button 
            onclick={handleCheckUpdates}
            disabled={isCheckingUpdates}
            class="px-5 py-2 bg-accent text-background font-bold text-xs rounded-lg hover:bg-accent-hover transition-all active:scale-[0.98] shadow-lg shadow-accent/20 disabled:opacity-50 disabled:cursor-wait"
          >
            {isCheckingUpdates ? $t('checking') : $t('checkForUpdates')}
          </button>
          {#if updateStatusKey}
            <span class="text-[9px] font-bold text-emerald-500 uppercase tracking-widest animate-pulse">
              {$t(updateStatusKey)}
            </span>
          {/if}
        </div>
      </div>
    </section>
  </div>
</div>
