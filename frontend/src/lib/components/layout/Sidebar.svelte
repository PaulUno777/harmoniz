<script lang="ts">
  import { t } from '../../stores/i18n'
  import { 
    MusicNotesIcon, 
    BrowserIcon, 
    TrashIcon, 
    GearIcon, 
    InfoIcon 
  } from 'phosphor-svelte'

  let { activeTab = $bindable('library') } = $props()

  const navItems = $derived([
    { id: 'library', label: $t('library'), icon: MusicNotesIcon },
    { id: 'organizer', label: $t('organizer'), icon: BrowserIcon },
    { id: 'cleaner', label: $t('cleaner'), icon: TrashIcon },
  ])
</script>

<aside class="w-52 h-full bg-surface border-r border-border flex flex-col pt-8">
  <div class="px-6 mb-8 flex items-center gap-3">
    <div class="w-8 h-8 bg-accent rounded-lg flex items-center justify-center shadow-lg shadow-accent/20">
      <MusicNotesIcon size={20} weight="bold" class="text-background" />
    </div>
    <h1 class="text-xl font-bold tracking-tight">{$t('appName')}</h1>
  </div>

  <nav class="flex-1 px-4 space-y-2">
    {#each navItems as item}
      <button 
        onclick={() => activeTab = item.id}
        class="w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 group
               {activeTab === item.id 
                 ? 'bg-accent/10 text-accent' 
                 : 'hover:bg-white/5 text-text-secondary hover:text-text-primary'}"
      >
        <item.icon 
          size={20} 
          weight={activeTab === item.id ? 'fill' : 'regular'} 
          class="transition-transform group-hover:scale-110"
        />
        <span class="font-medium">{item.label}</span>
      </button>
    {/each}
  </nav>

  <div class="p-4 space-y-2">
    <button 
      onclick={() => activeTab = 'settings'}
      class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-text-secondary hover:bg-white/5 hover:text-text-primary transition-all group"
    >
      <GearIcon size={20} class="group-hover:rotate-45 transition-transform duration-300" />
      <span class="font-medium">{$t('settings')}</span>
    </button>
    <div class="pt-4 px-3 flex items-center justify-between text-[10px] text-text-secondary uppercase tracking-widest font-bold opacity-50">
      <span>v0.1.0</span>
      <InfoIcon size={14} />
    </div>
  </div>
</aside>
