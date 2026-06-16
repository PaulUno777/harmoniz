<script lang="ts">
  import { onMount } from 'svelte'
  import { get } from 'svelte/store'
  import {
    FolderIcon,
    SpinnerIcon,
    CheckCircleIcon,
    ArrowClockwiseIcon,
    FloppyDiskIcon,
    EyeIcon,
    FileAudioIcon,
    MagicWandIcon,
    XIcon,
  } from 'phosphor-svelte'
  // @ts-ignore
  import { AnalyzeForOrganize, ApplyOrganizerSuggestion, ApplyAllHighConfidence, PreviewFilenameTemplate } from '../../../../wailsjs/go/main/App.js'
  import { organizer } from '../../stores/organizer'
  import { toast } from '../../stores/toast'
  import { t } from '../../stores/i18n'
  import type { OrganizerSuggestion, FilenamePreview } from '../../types'
  import ContextMenu from '../organizer/ContextMenu.svelte'

  // Svelte 4 style props (no runes — matches CleanerView pattern)
  export let libraryPath = ''
  export let onBrowse: () => void

  // Store refs — $store syntax works correctly in legacy (non-runes) mode
  const status          = organizer.status
  const suggestions     = organizer.suggestions
  const withSuggestions = organizer.withSuggestions
  const highConfCount   = organizer.highConfidenceCount
  const completeCount   = organizer.completeCount
  const filenamePreviews = organizer.filenamePreviews

  // Local state (plain Svelte 4 reactive variables)
  let template         = '{artist} - {title}.{ext}'
  let showAll          = false
  let showPreviewPanel = false
  let contextMenu: { suggestion: OrganizerSuggestion; x: number; y: number } | null = null
  let lastAnalyzedPath = ''

  // i18n
  let tr: (key: any) => string = (key: any) => String(key)
  let mounted = false

  onMount(() => {
    const unsub = t.subscribe(fn => { tr = fn as any })
    // Defer initial analysis to after mount — avoids reactive cascade during init
    mounted = true
    if (libraryPath && get(status) !== 'analyzing' && get(status) !== 'applying') {
      lastAnalyzedPath = libraryPath
      handleAnalyze()
    }
    return unsub
  })

  // Watch for path changes after mount — use get() so status is NOT a reactive dep here
  $: if (mounted && libraryPath && libraryPath !== lastAnalyzedPath) {
    if (get(status) !== 'analyzing' && get(status) !== 'applying') {
      lastAnalyzedPath = libraryPath
      handleAnalyze()
    }
  }

  async function handleAnalyze() {
    if (!libraryPath) return
    if (get(status) === 'analyzing' || get(status) === 'applying') return
    organizer.startAnalysis()
    try {
      const raw = await AnalyzeForOrganize(libraryPath)
      organizer.setResults(raw ?? [])
    } catch (e) {
      organizer.setError(e instanceof Error ? e.message : String(e))
      toast.error(`Analysis failed: ${e}`)
    }
  }

  async function handleApplyAll() {
    if (!libraryPath) return
    if (get(status) === 'analyzing' || get(status) === 'applying') return
    organizer.startApplying()
    try {
      const count = await ApplyAllHighConfidence(libraryPath, 75, template)
      organizer.setApplied(count)
      toast.success(`${count} ${tr('appliedSuccess')}`)
      lastAnalyzedPath = ''  // force re-analyze after apply
      await handleAnalyze()
    } catch (e) {
      organizer.setApplied(0)
      toast.error(`Apply failed: ${e}`)
    }
  }

  async function handleApplySuggestion(trackID: number, fields: Record<string, string>) {
    try {
      const artist   = fields['artist']    ?? ''
      const title    = fields['title']     ?? ''
      const album    = fields['album']     ?? ''
      const year     = fields['year']      ? parseInt(fields['year']) : 0
      const trackNum = fields['track_num'] ? parseInt(fields['track_num']) : 0
      const tmpl     = Object.keys(fields).length > 0 ? template : ''
      await ApplyOrganizerSuggestion(trackID, artist, title, album, year, trackNum, tmpl)
      organizer.markApplied(trackID)
      toast.success(tr('saved'))
    } catch (e) {
      toast.error(`${tr('saveError')}: ${e}`)
    }
  }

  async function handlePreviewTemplate() {
    if (!libraryPath || !template.trim()) return
    try {
      const previews: FilenamePreview[] = await PreviewFilenameTemplate(libraryPath, template)
      organizer.setPreviews(previews ?? [])
      showPreviewPanel = true
    } catch (e) {
      toast.error(`Preview failed: ${e}`)
    }
  }

  function handleContextMenu(e: MouseEvent, sugg: OrganizerSuggestion) {
    e.preventDefault()
    contextMenu = {
      suggestion: sugg,
      x: Math.min(e.clientX, window.innerWidth  - 328),
      y: Math.min(e.clientY, window.innerHeight - 448),
    }
  }

  function scoreClass(score: number): string {
    if (score >= 75) return 'bg-green-500/20 text-green-400 border-green-500/30'
    if (score >= 50) return 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30'
    return 'bg-white/5 text-text-muted border-border'
  }

  function scoreDot(score: number, issues: string[], fields: Record<string, unknown>): string {
    if (issues.length === 0 && Object.keys(fields).length === 0) return 'bg-green-500'
    if (score >= 75) return 'bg-green-500'
    if (score >= 50) return 'bg-yellow-500'
    if (score > 0)   return 'bg-yellow-500/60'
    return 'bg-red-500/50'
  }
</script>

<div class="relative h-full flex flex-col min-h-0 overflow-hidden">

  <!-- ───── Header ───── -->
  <div class="shrink-0 px-6 pt-5 pb-3 border-b border-border space-y-3">
    <div class="flex items-start justify-between gap-4">
      <div>
        <p class="text-xs text-text-muted mt-0.5">{tr('organizerDescription')}</p>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        {#if $status === 'ready'}
          <button
            on:click={handleAnalyze}
            disabled={$status === 'analyzing' || $status === 'applying' || !libraryPath}
            class="px-3 py-1.5 text-xs font-semibold rounded-lg border border-border hover:bg-white/5
                   text-text-secondary flex items-center gap-1.5 transition-colors disabled:opacity-40"
          >
            <ArrowClockwiseIcon size={13} />
            {tr('reanalyze')}
          </button>
        {:else}
          <button
            on:click={handleAnalyze}
            disabled={$status === 'analyzing' || $status === 'applying' || !libraryPath}
            class="px-4 py-1.5 text-xs font-bold rounded-lg bg-accent text-background
                   flex items-center gap-1.5 transition-all active:scale-95 disabled:opacity-40
                   hover:bg-accent/90 shadow-lg shadow-accent/20"
          >
            {#if $status === 'analyzing'}
              <SpinnerIcon size={13} class="animate-spin" />
              {tr('analyzing')}
            {:else}
              <MagicWandIcon size={13} weight="bold" />
              {tr('analyzeLibrary')}
            {/if}
          </button>
        {/if}
      </div>
    </div>

    <!-- Filename template row -->
    <div class="flex items-center gap-2">
      <label for="org-template" class="text-[10px] font-bold uppercase tracking-wider text-text-muted shrink-0">
        {tr('filenameTemplate')}
      </label>
      <input
        id="org-template"
        type="text"
        bind:value={template}
        class="flex-1 bg-background border border-border rounded-md px-2.5 py-1.5 text-xs font-mono
               text-text-primary outline-none focus:border-accent focus:ring-1 focus:ring-accent/20 transition-all"
        placeholder={'{artist} - {title}.{ext}'}
      />
      <button
        on:click={handlePreviewTemplate}
        disabled={$status === 'analyzing' || $status === 'applying' || !libraryPath || !template.trim()}
        class="px-2.5 py-1.5 text-xs font-semibold rounded-md border border-border hover:bg-white/5
               text-text-secondary flex items-center gap-1 transition-colors disabled:opacity-40 shrink-0"
      >
        <EyeIcon size={12} />
        {tr('previewTemplate')}
      </button>
    </div>

    <!-- Template hint -->
    <p class="text-[10px] text-text-muted/60 font-mono">
      {tr('templatePlaceholders')}
    </p>
  </div>

  <!-- ───── Progress bar (indeterminate) ───── -->
  {#if $status === 'analyzing' || $status === 'applying'}
    <div class="shrink-0 h-1 bg-background overflow-hidden">
      <div
        class="h-full bg-accent w-full"
        style="animation: indeterminate 1.5s ease-in-out infinite alternate"
      ></div>
    </div>
  {/if}

  <!-- ───── Stats bar ───── -->
  {#if ($status === 'ready' || $status === 'applying') && $suggestions.length > 0}
    <div class="shrink-0 px-6 py-2.5 bg-background/40 border-b border-border flex items-center gap-4 text-xs">
      <span class="text-text-muted">
        <span class="font-bold text-text-primary">{$suggestions.length}</span>
        &nbsp;{tr('filesAnalyzed')}
      </span>
      <span class="text-yellow-400/80">
        <span class="font-bold">{$withSuggestions.length}</span>
        &nbsp;{tr('withSuggestions')}
      </span>
      <span class="text-green-400/80">
        <span class="font-bold">{$highConfCount}</span>
        &nbsp;{tr('highConfidenceFiles')}
      </span>
      <span class="ml-auto text-text-muted">{$completeCount} complete</span>

      {#if $highConfCount > 0}
        <button
          on:click={handleApplyAll}
          disabled={$status === 'analyzing' || $status === 'applying'}
          class="ml-2 px-3 py-1 bg-accent text-background text-xs font-bold rounded-lg
                 flex items-center gap-1.5 hover:bg-accent/90 active:scale-95 transition-all
                 shadow-md shadow-accent/20 disabled:opacity-50"
        >
          {#if $status === 'applying'}
            <SpinnerIcon size={11} class="animate-spin" />
            {tr('applying')}
          {:else}
            <FloppyDiskIcon size={11} weight="bold" />
            {tr('applyAllHighConf')} ({$highConfCount})
          {/if}
        </button>
      {/if}

      <div class="flex rounded-md overflow-hidden border border-border text-[10px] font-bold shrink-0">
        <button
          on:click={() => showAll = false}
          class="px-2 py-1 transition-colors {!showAll ? 'bg-accent/20 text-accent' : 'text-text-muted hover:bg-white/5'}"
        >
          {tr('showSuggestionsOnly')}
        </button>
        <button
          on:click={() => showAll = true}
          class="px-2 py-1 transition-colors border-l border-border {showAll ? 'bg-accent/20 text-accent' : 'text-text-muted hover:bg-white/5'}"
        >
          {tr('showAllFiles')}
        </button>
      </div>
    </div>
  {/if}

  <!-- ───── Main content ───── -->
  <div class="flex-1 overflow-y-auto min-h-0">

    {#if !libraryPath}
      <!-- No library selected -->
      <div class="flex flex-col items-center justify-center h-full gap-5 text-center px-8">
        <div class="w-16 h-16 rounded-2xl bg-accent/10 flex items-center justify-center">
          <MagicWandIcon size={32} class="text-accent/60" weight="thin" />
        </div>
        <div>
          <p class="text-text-secondary font-medium mb-1">{tr('noLibraryForOrganizer')}</p>
          <p class="text-xs text-text-muted">{tr('organizerEmptyHint')}</p>
        </div>
        <button
          on:click={onBrowse}
          class="px-5 py-2 bg-accent/15 text-accent rounded-xl font-bold text-sm
                 hover:bg-accent/25 transition-colors flex items-center gap-2"
        >
          <FolderIcon size={16} />
          {tr('browse')}
        </button>
      </div>

    {:else if $status === 'idle'}
      <!-- Library set — analysis not yet started -->
      <div class="flex flex-col items-center justify-center h-full gap-5 text-center px-8">
        <div class="w-16 h-16 rounded-2xl bg-accent/10 flex items-center justify-center">
          <MagicWandIcon size={32} class="text-accent/60" weight="thin" />
        </div>
        <div>
          <p class="text-text-secondary font-medium mb-1">{tr('organizerEmptyHint')}</p>
          <p class="text-xs text-text-muted font-mono truncate max-w-xs">{libraryPath}</p>
        </div>
        <button
          on:click={handleAnalyze}
          class="px-5 py-2.5 bg-accent text-background rounded-xl font-bold text-sm
                 hover:bg-accent/90 active:scale-95 transition-all shadow-lg shadow-accent/20
                 flex items-center gap-2"
        >
          <MagicWandIcon size={16} weight="bold" />
          {tr('analyzeLibrary')}
        </button>
      </div>

    {:else if $status === 'analyzing'}
      <!-- Analysis in progress -->
      <div class="flex flex-col items-center justify-center h-full gap-4 text-text-muted">
        <SpinnerIcon size={32} class="animate-spin text-accent" />
        <p class="text-sm">{tr('analyzing')}</p>
        <p class="text-xs font-mono text-text-muted/60 max-w-xs truncate">{libraryPath}</p>
      </div>

    {:else if $status === 'applying'}
      <!-- Batch apply in progress — dimmed ghost list -->
      <div class="max-w-4xl mx-auto px-6 py-4 space-y-1 pointer-events-none opacity-40">
        {#each (showAll ? $suggestions : $withSuggestions).slice(0, 20) as sugg (sugg.track_id)}
          <div class="flex items-center gap-3 px-3 py-2.5 rounded-xl border border-border">
            <div class="w-2 h-2 rounded-full shrink-0 bg-text-muted/30"></div>
            <div class="w-8 h-8 shrink-0 rounded-lg bg-surface border border-border"></div>
            <div class="flex-1 min-w-0">
              <div class="text-xs font-semibold text-text-primary truncate">{sugg.track.title || sugg.track.filename}</div>
              <div class="text-[10px] text-text-muted truncate">{sugg.track.artist_raw || '—'}</div>
            </div>
          </div>
        {/each}
      </div>

    {:else if $status === 'ready' && $withSuggestions.length === 0 && !showAll}
      <!-- Ready — no suggestions, all metadata is complete -->
      <div class="flex flex-col items-center justify-center h-full gap-4 text-center px-8">
        <CheckCircleIcon size={40} class="text-green-400/70" weight="thin" />
        <p class="text-text-secondary font-medium">{tr('noSuggestionsFound')}</p>
        <p class="text-xs text-text-muted">{$suggestions.length} {tr('filesAnalyzed')}</p>
        <button on:click={() => showAll = true} class="text-xs text-accent hover:underline">
          {tr('showAllFiles')}
        </button>
      </div>

    {:else if $status === 'ready'}
      <!-- Ready — track list with suggestions -->
      <div class="max-w-4xl mx-auto px-6 py-4 space-y-1">
        {#each (showAll ? $suggestions : $withSuggestions) as sugg (sugg.track_id)}
          {@const hasFields = Object.keys(sugg.fields).length > 0}
          {@const isComplete = sugg.issues.length === 0 && !hasFields}

          <!-- svelte-ignore a11y-no-static-element-interactions -->
          <div
            class="flex items-center gap-3 px-3 py-2.5 rounded-xl border border-transparent
                   hover:bg-surface/60 hover:border-border transition-all group cursor-default"
            on:contextmenu={(e) => handleContextMenu(e, sugg)}
          >
            <!-- Score dot -->
            <div class="w-2 h-2 rounded-full shrink-0 {scoreDot(sugg.score, sugg.issues, sugg.fields)}"></div>

            <!-- Track icon -->
            <div class="w-8 h-8 shrink-0 rounded-lg bg-surface border border-border flex items-center justify-center">
              {#if isComplete}
                <CheckCircleIcon size={14} class="text-green-400/70" />
              {:else}
                <FileAudioIcon size={14} class="text-text-muted/60" />
              {/if}
            </div>

            <!-- Track info -->
            <div class="flex-1 min-w-0">
              <div class="text-xs font-semibold text-text-primary truncate">
                {sugg.track.title || sugg.track.filename}
              </div>
              <div class="text-[10px] text-text-muted truncate">
                {sugg.track.artist_raw || '—'}
                {#if sugg.track.album_raw}
                  <span class="opacity-50 mx-1">•</span>{sugg.track.album_raw}
                {/if}
              </div>
            </div>

            <!-- Field tags -->
            {#if hasFields}
              <div class="hidden sm:flex items-center gap-1 shrink-0">
                {#each Object.keys(sugg.fields).slice(0, 3) as key}
                  <span class="text-[9px] px-1.5 py-0.5 rounded-full bg-accent/10 text-accent font-bold uppercase tracking-wide">
                    {key === 'track_num' ? 'track#' : key}
                  </span>
                {/each}
                {#if Object.keys(sugg.fields).length > 3}
                  <span class="text-[9px] text-text-muted">+{Object.keys(sugg.fields).length - 3}</span>
                {/if}
              </div>
            {:else if isComplete}
              <span class="text-[10px] text-green-400/70 shrink-0 hidden sm:block">{tr('trackComplete')}</span>
            {/if}

            <!-- Score badge -->
            {#if hasFields}
              <span class="shrink-0 text-[10px] font-bold tabular-nums px-2 py-0.5 rounded-full border {scoreClass(sugg.score)}">
                {Math.round(sugg.score)}%
              </span>
            {/if}

            <!-- Quick apply (high confidence only) -->
            {#if hasFields && sugg.score >= 75}
              <button
                on:click={async (e) => {
                  e.stopPropagation()
                  const allFields: Record<string, string> = {}
                  for (const [k, v] of Object.entries(sugg.fields)) allFields[k] = v.value
                  await handleApplySuggestion(sugg.track_id, allFields)
                }}
                class="shrink-0 p-1.5 rounded-md bg-accent/10 text-accent hover:bg-accent/25
                       opacity-0 group-hover:opacity-100 transition-all active:scale-95"
                title={tr('applyAllForFile')}
              >
                <FloppyDiskIcon size={13} weight="bold" />
              </button>
            {:else}
              <div class="w-7 shrink-0"></div>
            {/if}

            <!-- Right-click hint -->
            <div class="text-[9px] text-text-muted/40 shrink-0 hidden sm:block opacity-0 group-hover:opacity-100 transition-opacity">
              right-click
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- ───── Filename Preview Panel ───── -->
  {#if showPreviewPanel}
    <div class="absolute inset-0 z-40 bg-background/80 backdrop-blur-sm flex items-center justify-center p-6">
      <div class="bg-surface border border-border rounded-2xl shadow-2xl w-full max-w-2xl max-h-[80vh] flex flex-col overflow-hidden">
        <div class="px-5 py-4 border-b border-border flex items-center justify-between">
          <div>
            <h3 class="text-sm font-bold">{tr('templatePreviewTitle')}</h3>
            <p class="text-xs text-text-muted mt-0.5 font-mono">{template}</p>
          </div>
          <button
            on:click={() => showPreviewPanel = false}
            class="p-1.5 rounded-lg hover:bg-white/5 text-text-muted transition-colors"
          >
            <XIcon size={16} />
          </button>
        </div>
        <div class="flex-1 overflow-y-auto px-5 py-3 space-y-1.5">
          {#if $filenamePreviews.length === 0}
            <p class="text-xs text-text-muted text-center py-8">No filenames would change.</p>
          {:else}
            {#each $filenamePreviews.slice(0, 200) as preview}
              <div class="text-[11px] rounded-lg bg-background p-2.5">
                <div class="font-mono text-text-muted/70 line-through truncate">{preview.before}</div>
                <div class="font-mono text-text-primary font-semibold truncate mt-0.5">{preview.after}</div>
              </div>
            {/each}
            {#if $filenamePreviews.length > 200}
              <p class="text-xs text-text-muted text-center pt-2">
                …and {$filenamePreviews.length - 200} more
              </p>
            {/if}
          {/if}
        </div>
        <div class="px-5 py-3 border-t border-border">
          <button
            on:click={() => showPreviewPanel = false}
            class="w-full py-2 text-xs font-bold rounded-lg border border-border hover:bg-white/5 transition-colors"
          >
            {tr('closePreview')}
          </button>
        </div>
      </div>
    </div>
  {/if}

  <!-- ───── Context Menu ───── -->
  {#if contextMenu}
    <ContextMenu
      suggestion={contextMenu.suggestion}
      x={contextMenu.x}
      y={contextMenu.y}
      filenameTemplate={template}
      onClose={() => { contextMenu = null }}
      onApply={async (trackID, fields) => {
        await handleApplySuggestion(trackID, fields)
        contextMenu = null
      }}
    />
  {/if}
</div>

<style>
  @keyframes indeterminate {
    0%   { transform: scaleX(0.1) translateX(-100%); }
    100% { transform: scaleX(0.8) translateX(200%); }
  }
</style>
