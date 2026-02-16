<script lang="ts">
  import type { Component } from 'svelte'

  interface Props {
    value: string
    selectedValue: string
    label: string
    icon: Component
    previewBg: string
    previewBorder?: string
    previewIconClass?: string
    isSystem?: boolean
    onClick: () => void
  }

  let {
    value,
    selectedValue,
    label,
    icon: Icon,
    previewBg,
    previewBorder = 'border-border',
    previewIconClass = 'text-text-primary',
    isSystem = false,
    onClick
  }: Props = $props()

  const isSelected = $derived(value === selectedValue)
  
  const bgStyle = $derived(isSystem 
    ? `background: ${previewBg}; border-color: ${previewBorder};`
    : `background-color: ${previewBg}; border-color: ${previewBorder};`
  )
</script>

<button 
  onclick={onClick}
  class="flex flex-col items-center gap-2 p-3 rounded-xl border-2 transition-all max-w-[120px] w-full
        {isSelected ? 'border-accent bg-accent/5' : 'border-border opacity-50 hover:opacity-80 hover:border-text-muted'}"
>
  <div 
    class="w-full aspect-4/3 rounded-lg border flex items-center justify-center shadow-inner max-h-[120px] overflow-hidden relative"
    style={bgStyle}
  >
    {#if isSystem}
      <div class="absolute inset-0 flex">
        <div class="flex-1 bg-background"></div>
        <div class="flex-1 bg-[#f8fafc]"></div>
      </div>
    {/if}
    <Icon size={20} class={previewIconClass} />
  </div>
  <span class="text-[9px] font-bold uppercase tracking-widest opacity-70 text-center">{label}</span>
</button>
