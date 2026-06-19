import { writable } from 'svelte/store'
import { domain } from '../../../wailsjs/go/models'

// @ts-ignore — Wails bindings not available in browser dev mode
import { GetConfig, UpdateConfig } from '../../../wailsjs/go/main/App'

type AppConfig = domain.AppConfig

function defaultConfig(): AppConfig {
  return domain.AppConfig.createFrom({
    cleaner: { similarity_threshold: 0.88, artist_group_threshold: 0.90 },
    organizer: { field_confidence_threshold: 0.75 },
  })
}

const _cfg = writable<AppConfig | null>(null)

export const appConfig = {
  subscribe: _cfg.subscribe,

  async load(): Promise<void> {
    try {
      const cfg = await GetConfig()
      _cfg.set(cfg)
    } catch {
      _cfg.set(defaultConfig())
    }
  },

  // Update the store immediately (for live slider display) without saving to backend.
  setLocal(cfg: AppConfig): void {
    _cfg.set(cfg)
  },

  // Persist to backend immediately. Call this on slider `onchange` (user releases).
  async save(cfg: AppConfig): Promise<void> {
    _cfg.set(cfg)
    try {
      await UpdateConfig(cfg)
    } catch (e) {
      console.warn('[settings] UpdateConfig failed:', e)
    }
  },

  async resetDefaults(): Promise<void> {
    await this.save(defaultConfig())
  },
}
