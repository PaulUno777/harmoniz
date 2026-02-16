export interface Track {
  title: string
  artist: string
  album: string
  year: number
  path: string
  size: string
}

export type TabId = 'library' | 'organizer' | 'cleaner' | 'settings'
