<script lang="ts">
  import {
    FileAudioIcon,
    CalendarIcon,
    HashIcon,
    FolderIcon,
    TagIcon,
    FloppyDiskIcon,
    PencilSimpleIcon,
    CheckCircleIcon,
    SpinnerIcon,
    MusicNoteSimpleIcon,
  } from 'phosphor-svelte'

  // @ts-ignore
  import { UpdateTrackTags, RenameTrack } from '../../../../wailsjs/go/main/App.js'
  import type { Track } from '../../types'
  import { t } from '../../stores/i18n'
  import { toast } from '../../stores/toast'
  import { formatFileSize } from '../../utils/format'

  interface Props {
    selectedTrack: Track | null
    onTrackSaved?: (updated: Track) => void
  }

  let { selectedTrack = $bindable(null), onTrackSaved }: Props = $props()

  // Editable field state — reset whenever the selected track changes
  let editTitle    = $state('')
  let editArtist   = $state('')
  let editAlbum    = $state('')
  let editYear     = $state(0)
  let editTrackNum = $state(0)
  let editFilename = $state('')

  let isSaving   = $state(false)
  let isRenaming = $state(false)

  $effect(() => {
    if (selectedTrack) {
      editTitle    = selectedTrack.title     ?? ''
      editArtist   = selectedTrack.artist    ?? selectedTrack.artist_raw ?? ''
      editAlbum    = selectedTrack.album     ?? selectedTrack.album_raw  ?? ''
      editYear     = selectedTrack.year      ?? 0
      editTrackNum = selectedTrack.track_num ?? 0
      editFilename = selectedTrack.filename  ?? ''
    }
  })

  async function handleSave() {
    if (!selectedTrack?.id || isSaving) return
    isSaving = true
    try {
      await UpdateTrackTags(
        selectedTrack.id,
        editArtist,
        editTitle,
        editAlbum,
        editYear,
        editTrackNum,
      )
      const updated: Track = {
        ...selectedTrack,
        title:      editTitle,
        artist:     editArtist,
        artist_raw: editArtist,
        album:      editAlbum,
        album_raw:  editAlbum,
        year:       editYear,
        track_num:  editTrackNum,
      }
      selectedTrack = updated
      onTrackSaved?.(updated)
      toast.success($t('saved'))
    } catch (err) {
      toast.error(`${$t('saveError')}: ${err}`)
    } finally {
      isSaving = false
    }
  }

  async function handleRename() {
    if (!selectedTrack?.id || isRenaming || !editFilename.trim()) return
    isRenaming = true
    try {
      const newPath = await RenameTrack(selectedTrack.id, editFilename.trim())
      const pathParts = newPath.split('/')
      const newFilename = pathParts.length > 0 ? (pathParts[pathParts.length - 1] ?? editFilename.trim()) : editFilename.trim()
      const updated: Track = {
        ...selectedTrack,
        path:     newPath,
        filename: newFilename,
      }
      editFilename = newFilename
      selectedTrack = updated
      onTrackSaved?.(updated)
      toast.success($t('renameSuccess'))
    } catch (err) {
      toast.error(`${$t('renameError')}: ${err}`)
    } finally {
      isRenaming = false
    }
  }

  const isDirty = $derived(
    selectedTrack != null && (
      editTitle    !== (selectedTrack.title     ?? '') ||
      editArtist   !== (selectedTrack.artist    ?? selectedTrack.artist_raw ?? '') ||
      editAlbum    !== (selectedTrack.album     ?? selectedTrack.album_raw  ?? '') ||
      editYear     !== (selectedTrack.year      ?? 0) ||
      editTrackNum !== (selectedTrack.track_num ?? 0)
    )
  )

  const filenameChanged = $derived(
    selectedTrack != null && editFilename !== (selectedTrack.filename ?? '')
  )
</script>

<aside class="w-80 h-full bg-surface border-l border-border flex flex-col pt-6">
  <div class="px-6 mb-4">
    <h2 class="text-[11px] font-bold uppercase tracking-wider text-text-secondary opacity-60">
      {$t('details')}
    </h2>
  </div>

  {#if selectedTrack}
    <div class="flex-1 overflow-y-auto px-6 space-y-5 pb-4">

      <!-- Album art placeholder -->
      <div class="aspect-square w-full bg-background rounded-xl flex items-center justify-center border border-border group relative overflow-hidden">
        <div class="flex flex-col items-center gap-2 text-text-secondary/40 transition-transform duration-300 group-hover:scale-110">
          <MusicNoteSimpleIcon size={48} weight="thin" />
        </div>
        <div class="absolute inset-0 bg-accent/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>
      </div>

      <!-- Metadata fields -->
      <div class="space-y-3">

        <!-- Title -->
        <div>
          <label for="ctx-title" class="field-label">{$t('title')}</label>
          <input
            id="ctx-title"
            type="text"
            bind:value={editTitle}
            class="field-input"
            placeholder="Unknown title"
          />
        </div>

        <!-- Artist -->
        <div>
          <label for="ctx-artist" class="field-label">{$t('artist')}</label>
          <input
            id="ctx-artist"
            type="text"
            bind:value={editArtist}
            class="field-input"
            placeholder="Unknown artist"
          />
        </div>

        <!-- Album -->
        <div>
          <label for="ctx-album" class="field-label">{$t('album')}</label>
          <input
            id="ctx-album"
            type="text"
            bind:value={editAlbum}
            class="field-input"
            placeholder="Unknown album"
          />
        </div>

        <!-- Year + Track # -->
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label for="ctx-year" class="field-label flex items-center gap-1">
              <CalendarIcon size={11} />{$t('year')}
            </label>
            <input
              id="ctx-year"
              type="number"
              bind:value={editYear}
              min="0"
              max="9999"
              class="field-input"
              placeholder="0"
            />
          </div>
          <div>
            <label for="ctx-tracknum" class="field-label flex items-center gap-1">
              <HashIcon size={11} />{$t('trackNum')}
            </label>
            <input
              id="ctx-tracknum"
              type="number"
              bind:value={editTrackNum}
              min="0"
              class="field-input"
              placeholder="0"
            />
          </div>
        </div>

        <!-- Size (read-only) -->
        <div>
          <span class="field-label flex items-center gap-1">
            <FileAudioIcon size={11} />{$t('size')}
          </span>
          <div class="px-3 py-1.5 text-xs text-text-secondary bg-background border border-border rounded-md opacity-60 select-all">
            {formatFileSize(selectedTrack.size)}
          </div>
        </div>

        <!-- Filename + rename -->
        <div>
          <label for="ctx-filename" class="field-label flex items-center gap-1">
            <PencilSimpleIcon size={11} />{$t('filename')}
          </label>
          <div class="flex gap-2">
            <input
              id="ctx-filename"
              type="text"
              bind:value={editFilename}
              class="field-input flex-1 min-w-0 font-mono text-xs"
            />
            <button
              onclick={handleRename}
              disabled={isRenaming || !filenameChanged}
              class="shrink-0 px-3 py-1.5 text-xs font-semibold rounded-md transition-all
                     bg-accent/15 text-accent hover:bg-accent/25 active:scale-95
                     disabled:opacity-30 disabled:cursor-not-allowed"
              title={$t('rename')}
            >
              {#if isRenaming}
                <SpinnerIcon size={13} class="animate-spin" />
              {:else}
                {$t('rename')}
              {/if}
            </button>
          </div>
        </div>

        <!-- Path (read-only) -->
        <div>
          <span class="field-label flex items-center gap-1">
            <FolderIcon size={11} />{$t('path')}
          </span>
          <div class="text-[11px] font-mono break-all text-text-secondary/70 bg-background p-2 rounded border border-border leading-relaxed select-all">
            {selectedTrack.path}
          </div>
        </div>

      </div>
    </div>

    <!-- Save changes footer -->
    <div class="shrink-0 p-5 border-t border-border">
      <button
        onclick={handleSave}
        disabled={isSaving || !isDirty}
        class="w-full font-bold py-2.5 rounded-lg flex items-center justify-center gap-2 transition-all
               active:scale-[0.98] shadow-lg
               {isDirty
                 ? 'bg-accent hover:bg-accent/90 text-background shadow-accent/20 cursor-pointer'
                 : 'bg-accent/20 text-accent/50 shadow-none cursor-not-allowed'}"
      >
        {#if isSaving}
          <SpinnerIcon size={16} weight="bold" class="animate-spin" />
          {$t('saving')}
        {:else if !isDirty}
          <CheckCircleIcon size={16} weight="bold" />
          {$t('saved')}
        {:else}
          <FloppyDiskIcon size={16} weight="bold" />
          {$t('saveChanges')}
        {/if}
      </button>
    </div>

  {:else}
    <div class="flex-1 flex flex-col items-center justify-center text-text-secondary px-10 text-center gap-4">
      <div class="w-16 h-16 bg-white/5 rounded-full flex items-center justify-center">
        <TagIcon size={32} weight="thin" class="opacity-40" />
      </div>
      <p class="text-xs font-medium leading-relaxed opacity-60">{$t('selectTrackToEdit')}</p>
    </div>
  {/if}
</aside>

<style>
  .field-label {
    display: block;
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--color-text-secondary);
    opacity: 0.7;
    margin-bottom: 4px;
  }

  .field-input {
    width: 100%;
    background: var(--color-background);
    border: 1px solid var(--color-border);
    border-radius: 6px;
    padding: 6px 10px;
    font-size: 13px;
    outline: none;
    transition: border-color 0.15s, box-shadow 0.15s;
    color: var(--color-text-primary);
  }

  .field-input:focus {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-accent) 15%, transparent);
  }
</style>
