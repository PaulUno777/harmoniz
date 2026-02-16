export interface Track {
  id?: number
  title: string
  artist_raw?: string
  artist?: string // Maps to artist_raw (computed in frontend)
  album_raw?: string
  album?: string // Maps to album_raw (computed in frontend)
  year: number
  path: string
  filename?: string
  size: number // Size in bytes
  mod_time?: number
  track_num?: number
  hash_partial?: string
  status?: string
}

export type TabId = 'library' | 'organizer' | 'cleaner' | 'settings'
