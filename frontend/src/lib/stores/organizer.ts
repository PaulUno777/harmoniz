import { writable, derived } from 'svelte/store'
import type { OrganizerSuggestion, FilenamePreview } from '../types'

export type OrganizerStatus = 'idle' | 'analyzing' | 'ready' | 'applying'

const suggestions = writable<OrganizerSuggestion[]>([])
const status = writable<OrganizerStatus>('idle')
const error = writable<string | null>(null)
const filenameTemplate = writable('{artist} - {title}.{ext}')
const filenamePreviews = writable<FilenamePreview[]>([])
const lastAppliedCount = writable(0)

const withSuggestions = derived(suggestions, ($s) =>
  $s.filter(s => Object.keys(s.fields).length > 0)
)

const highConfidenceCount = derived(suggestions, ($s) =>
  $s.filter(s => s.score >= 75).length
)

const completeCount = derived(suggestions, ($s) =>
  $s.filter(s => s.issues.length === 0 && Object.keys(s.fields).length === 0).length
)

export const organizer = {
  suggestions,
  status,
  error,
  filenameTemplate,
  filenamePreviews,
  lastAppliedCount,
  withSuggestions,
  highConfidenceCount,
  completeCount,

  startAnalysis() {
    status.set('analyzing')
    error.set(null)
    filenamePreviews.set([])
  },

  setResults(suggs: OrganizerSuggestion[]) {
    suggestions.set(suggs ?? [])
    status.set('ready')
  },

  setError(msg: string) {
    error.set(msg)
    status.set('idle')
  },

  startApplying() {
    status.set('applying')
  },

  setApplied(count: number) {
    lastAppliedCount.set(count)
    status.set('ready')
  },

  setPreviews(previews: FilenamePreview[]) {
    filenamePreviews.set(previews ?? [])
  },

  /** Update a single track's suggestion in place after applying it. */
  markApplied(trackID: number) {
    suggestions.update($s => $s.map(s =>
      s.track_id === trackID
        ? { ...s, fields: {}, issues: [], score: 0 }
        : s
    ))
  },

  reset() {
    suggestions.set([])
    status.set('idle')
    error.set(null)
    filenamePreviews.set([])
    lastAppliedCount.set(0)
  },
}
