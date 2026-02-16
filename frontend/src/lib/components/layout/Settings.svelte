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

  let isCheckingUpdates = $state(false)
  let updateStatus = $state<string | null>(null)
  
  const languages: { id: Locale; label: string }[] = [
    { id: 'en', label: 'English' },
    { id: 'fr', label: 'Français' },
  ]

  async function handleCheckUpdates() {
    isCheckingUpdates = true
    updateStatus = null
    
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    isCheckingUpdates = false
    updateStatus = 'You are on the latest version!'
  }
</script>

<div class="flex-1 overflow-y-auto p-8 max-w-4xl mx-auto w-full custom-scrollbar">
  <div class="mb-8">
    <h1 class="text-2xl font-bold mb-1 font-display">{$t('settings')}</h1>
    <p class="text-text-secondary text-sm">Configure Harmonizr to your preferences.</p>
  </div>

  <div class="space-y-8 pb-12">
    <!-- Appearance -->
    <section>
      <div class="flex items-center gap-3 mb-4">
        <SunIcon size={20} weight="bold" class="text-accent" />
        <h2 class="text-md font-bold font-display">{$t('appearance')}</h2>
      </div>

      <div class="bg-surface rounded-2xl border border-border p-5 shadow-sm space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 items-start">
          <!-- Dark Preview -->
          <button 
            onclick={() => $theme = 'dark'}
            class="flex flex-col items-center gap-2 p-3 rounded-xl border-2 transition-all 
                      {$theme === 'dark' ? 'border-accent bg-accent/5' : 'border-border opacity-50 hover:opacity-80 hover:border-text-muted'}">
            <div class="w-full aspect-[4/3] bg-[#0f1115] rounded-lg border border-[#2c2e36] flex items-center justify-center shadow-inner">
              <MoonIcon size={20} class="text-[#eaeaeb]" />
            </div>
            <span class="text-[9px] font-bold uppercase tracking-widest opacity-70">{$t('dark')}</span>
          </button>

          <!-- Light Preview -->
          <button 
            onclick={() => $theme = 'light'}
            class="flex flex-col items-center gap-2 p-3 rounded-xl border-2 transition-all
                      {$theme === 'light' ? 'border-accent bg-accent/5' : 'border-border opacity-50 hover:opacity-80 hover:border-text-muted'}">
            <div class="w-full aspect-[4/3] bg-[#f8fafc] rounded-lg border border-[#e2e8f0] flex items-center justify-center shadow-inner">
              <SunIcon size={20} class="text-[#0f172a]" />
            </div>
            <span class="text-[9px] font-bold uppercase tracking-widest opacity-70">{$t('light')}</span>
          </button>

          <!-- System Preview -->
          <button 
            onclick={() => $theme = 'system'}
            class="flex flex-col items-center gap-2 p-3 rounded-xl border-2 transition-all
                      {$theme === 'system' ? 'border-accent bg-accent/5' : 'border-border opacity-50 hover:opacity-80 hover:border-text-muted'}">
            <div class="w-full aspect-[4/3] bg-gradient-to-br from-[#0f1115] to-[#f8fafc] rounded-lg border border-border flex items-center justify-center shadow-inner overflow-hidden relative">
              <div class="absolute inset-0 flex">
                <div class="flex-1 bg-[#0f1115]"></div>
                <div class="flex-1 bg-[#f8fafc]"></div>
              </div>
              <MonitorIcon size={20} class="relative z-10 text-accent mix-blend-difference" />
            </div>
            <span class="text-[9px] font-bold uppercase tracking-widest opacity-70">{$t('system')}</span>
          </button>
        </div>
        
        <p class="text-[10px] text-text-muted px-1 leading-relaxed">
          {$theme === 'system' ? 'Matching your operating system preference.' : `Currently using ${$theme === 'dark' ? 'Dark' : 'Light'} mode.`}
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
            <div class="text-[11px] text-text-secondary">A premium tool for music library organization.</div>
          </div>
        </div>
        
        <div class="flex flex-col items-end gap-1.5">
          <button 
            onclick={handleCheckUpdates}
            disabled={isCheckingUpdates}
            class="px-5 py-2 bg-accent text-background font-bold text-xs rounded-lg hover:bg-accent-hover transition-all active:scale-[0.98] shadow-lg shadow-accent/20 disabled:opacity-50 disabled:cursor-wait"
          >
            {isCheckingUpdates ? 'Checking...' : $t('checkForUpdates')}
          </button>
          {#if updateStatus}
            <span class="text-[9px] font-bold text-emerald-500 uppercase tracking-widest animate-pulse">
              {updateStatus}
            </span>
          {/if}
        </div>
      </div>
    </section>
  </div>
</div>
