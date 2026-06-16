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

export type TabId = 'library' | 'organizer' | 'cleaner' | 'settings'
