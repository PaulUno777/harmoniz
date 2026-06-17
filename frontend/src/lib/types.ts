export interface Track {
  id?: number
  title: string
  artist_raw?: string
  artist?: string     // display alias for artist_raw (mapped on load)
  album_raw?: string
  album?: string      // display alias for album_raw (mapped on load)
  year: number
  path: string
  filename?: string
  size: number        // bytes
  mod_time?: number
  track_num?: number
  bitrate?: number
  hash_partial?: string
  status?: string
}

/** Phase 4 analysis: similar artist name suggestion from clustering */
export interface ArtistSuggestion {
  original: string
  suggested: string
  score: number
  reason: string
  confidence_level: string
}

/** Phase 5b organizer: per-field suggestion with confidence and source info */
export interface FieldSuggestion {
  value: string
  confidence: number  // 0.0–1.0
  // Backend models come from Wails bindings where this may be a plain string.
  source: 'path' | 'filename' | 'neighbor' | string
}

/** Phase 5b organizer: all suggestions for a single track */
export interface OrganizerSuggestion {
  track_id: number
  track: Track
  score: number  // 0–100 (confidence of suggestions; 0 = no suggestions or already complete)
  fields: Record<string, FieldSuggestion>  // keys: "artist" "title" "album" "year" "track_num"
  issues: string[]
}

/** Phase 5b organizer: before/after filename pair for template preview */
export interface FilenamePreview {
  track_id: number
  before: string
  after: string
}

/** Phase 4 analysis: a group of byte-identical duplicate tracks with a recommended keeper */
export interface DuplicateGroup {
  tracks: Track[]
  recommended_keep_id: number
}

export type TabId = 'library' | 'organizer' | 'cleaner' | 'settings'
