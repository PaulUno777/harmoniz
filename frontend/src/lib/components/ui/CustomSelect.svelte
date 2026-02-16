<script lang="ts">
  import { CaretDownIcon, CheckIcon } from 'phosphor-svelte'
  import { fade, scale } from 'svelte/transition'

  interface Option {
    id: string;
    label: string;
  }

  interface Props {
    options: Option[];
    value: string;
    label?: string;
    placeholder?: string;
    onSelect?: (value: string) => void;
  }

  let { 
    options, 
    value = $bindable(), 
    label, 
    placeholder = 'Select...', 
    onSelect 
  }: Props = $props()

  let isOpen = $state(false)
  let containerRef = $state<HTMLDivElement>()

  const selectedOption = $derived(options.find(o => o.id === value))

  function handleSelect(id: string) {
    value = id
    isOpen = false
    onSelect?.(id)
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') isOpen = false
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      isOpen = !isOpen
    }
  }

  // Close when clicking outside
  function handleClickOutside(e: MouseEvent) {
    if (containerRef && !containerRef.contains(e.target as Node)) {
      isOpen = false
    }
  }
</script>

<svelte:window onclick={handleClickOutside} />

<div class="space-y-1.5" bind:this={containerRef}>
  {#if label}
    <span class="text-[10px] font-bold uppercase text-text-secondary tracking-widest block px-1">
      {label}
    </span>
  {/if}

  <div class="relative">
    <button
      type="button"
      onclick={() => isOpen = !isOpen}
      onkeydown={handleKeydown}
      class="w-full bg-surface border border-border rounded-xl px-4 py-2.5 text-sm transition-all text-left flex items-center justify-between group
             hover:border-text-muted active:scale-[0.99]
             {isOpen ? 'ring-2 ring-accent/20 border-accent' : ''}"
      aria-haspopup="listbox"
      aria-expanded={isOpen}
    >
      <span class={selectedOption ? 'text-text-primary font-medium' : 'text-text-muted'}>
        {selectedOption ? selectedOption.label : placeholder}
      </span>
      <CaretDownIcon 
        size={16} 
        class="text-text-secondary transition-transform duration-300 {isOpen ? 'rotate-180 text-accent' : ''}" 
      />
    </button>

    {#if isOpen}
      <div
        in:scale={{ duration: 150, start: 0.95, opacity: 0 }}
        out:fade={{ duration: 100 }}
        class="absolute z-50 bottom-full mb-2 w-full bg-surface border border-border rounded-xl shadow-2xl shadow-black/50 overflow-hidden backdrop-blur-xl"
        role="listbox"
      >
        <div class="p-1.5 max-h-60 overflow-y-auto custom-scrollbar">
          {#each options as option}
            <button
              type="button"
              onclick={() => handleSelect(option.id)}
              class="w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm transition-all text-left group
                     {value === option.id ? 'bg-accent/10 text-accent font-bold' : 'hover:bg-white/5 text-text-secondary hover:text-text-primary'}"
              role="option"
              aria-selected={value === option.id}
            >
              <span>{option.label}</span>
              {#if value === option.id}
                <CheckIcon size={14} weight="bold" />
              {/if}
            </button>
          {/each}
        </div>
      </div>
    {/if}
  </div>
</div>
