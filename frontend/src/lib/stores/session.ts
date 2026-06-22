import type { FilterState } from "../components/layout/FilterPanel.svelte";
import type { TabId } from "../types";

export type Locale = "en" | "fr";

export type LoopMode = "none" | "one" | "all";

export interface PersistedPlayback {
  trackPath: string;
  currentTime: number;
  volume: number;
  isMuted: boolean;
  loopMode: LoopMode;
  shuffleMode: boolean;
}

export interface PersistedUpdateCheck {
  checkedAt: string;
  latestVersion: string;
  updateAvailable: boolean;
  downloadUrl: string;
  releasePageUrl: string;
  status: "ready" | "error";
  errorMessage?: string | null;
}

export interface PersistedSession {
  locale: Locale;
  libraryPath: string;
  activeTab: TabId;
  searchQuery: string;
  filters: FilterState;
  listOffset: number;
  selectedTrackPath: string | null;
  playback: PersistedPlayback | null;
  dismissedUpdateVersion?: string;
  updateCheckCache?: PersistedUpdateCheck;
}

const STORAGE_KEY = "harmoniz.session";

const DEFAULT_FILTERS: FilterState = {
  yearMin: 0,
  yearMax: 0,
  sizeMin: 0,
  sizeMax: 0,
};

const defaultSession: PersistedSession = {
  locale: "en",
  libraryPath: "",
  activeTab: "library",
  searchQuery: "",
  filters: { ...DEFAULT_FILTERS },
  listOffset: 0,
  selectedTrackPath: null,
  playback: null,
};

const VALID_TABS: TabId[] = ["library", "organizer", "cleaner", "settings"];
const VALID_LOCALES: Locale[] = ["en", "fr"];
const VALID_LOOP_MODES: LoopMode[] = ["none", "one", "all"];

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null;
}

function parseFilters(raw: unknown): FilterState {
  if (!isRecord(raw)) return { ...DEFAULT_FILTERS };
  return {
    yearMin: typeof raw.yearMin === "number" ? raw.yearMin : 0,
    yearMax: typeof raw.yearMax === "number" ? raw.yearMax : 0,
    sizeMin: typeof raw.sizeMin === "number" ? raw.sizeMin : 0,
    sizeMax: typeof raw.sizeMax === "number" ? raw.sizeMax : 0,
  };
}

function parsePlayback(raw: unknown): PersistedPlayback | null {
  if (!isRecord(raw) || typeof raw.trackPath !== "string" || !raw.trackPath) {
    return null;
  }
  const loopMode = raw.loopMode;
  return {
    trackPath: raw.trackPath,
    currentTime: typeof raw.currentTime === "number" ? raw.currentTime : 0,
    volume: typeof raw.volume === "number" ? raw.volume : 0.8,
    isMuted: raw.isMuted === true,
    loopMode:
      typeof loopMode === "string" &&
      VALID_LOOP_MODES.includes(loopMode as LoopMode)
        ? (loopMode as LoopMode)
        : "none",
    shuffleMode: raw.shuffleMode === true,
  };
}

function parseSession(raw: unknown): PersistedSession | null {
  if (!isRecord(raw)) return null;

  const locale = raw.locale;
  const activeTab = raw.activeTab;

  return {
    locale:
      typeof locale === "string" && VALID_LOCALES.includes(locale as Locale)
        ? (locale as Locale)
        : "en",
    libraryPath: typeof raw.libraryPath === "string" ? raw.libraryPath : "",
    activeTab:
      typeof activeTab === "string" && VALID_TABS.includes(activeTab as TabId)
        ? (activeTab as TabId)
        : "library",
    searchQuery: typeof raw.searchQuery === "string" ? raw.searchQuery : "",
    filters: parseFilters(raw.filters),
    listOffset:
      typeof raw.listOffset === "number" ? Math.max(0, raw.listOffset) : 0,
    selectedTrackPath:
      typeof raw.selectedTrackPath === "string" ? raw.selectedTrackPath : null,
    playback: parsePlayback(raw.playback),
    dismissedUpdateVersion:
      typeof raw.dismissedUpdateVersion === "string"
        ? raw.dismissedUpdateVersion
        : undefined,
    updateCheckCache: parseUpdateCheckCache(raw.updateCheckCache),
  };
}

function parseUpdateCheckCache(raw: unknown): PersistedUpdateCheck | undefined {
  if (!isRecord(raw) || typeof raw.checkedAt !== "string") return undefined;
  const status = raw.status;
  if (status !== "ready" && status !== "error") return undefined;
  return {
    checkedAt: raw.checkedAt,
    latestVersion:
      typeof raw.latestVersion === "string" ? raw.latestVersion : "",
    updateAvailable: raw.updateAvailable === true,
    downloadUrl: typeof raw.downloadUrl === "string" ? raw.downloadUrl : "",
    releasePageUrl:
      typeof raw.releasePageUrl === "string" ? raw.releasePageUrl : "",
    status,
    errorMessage:
      typeof raw.errorMessage === "string" ? raw.errorMessage : null,
  };
}

let cachedSession: PersistedSession = { ...defaultSession };
let saveTimer: ReturnType<typeof setTimeout> | null = null;
let pendingPartial: Partial<PersistedSession> = {};

export function loadSession(): PersistedSession {
  if (typeof localStorage === "undefined") {
    cachedSession = { ...defaultSession };
    return cachedSession;
  }

  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      cachedSession = { ...defaultSession };
      return cachedSession;
    }
    const parsed = parseSession(JSON.parse(raw));
    cachedSession = parsed ?? { ...defaultSession };
    return cachedSession;
  } catch {
    cachedSession = { ...defaultSession };
    return cachedSession;
  }
}

export function saveSession(partial: Partial<PersistedSession>): void {
  cachedSession = {
    ...cachedSession,
    ...partial,
    filters: partial.filters
      ? { ...cachedSession.filters, ...partial.filters }
      : cachedSession.filters,
  };

  if (typeof localStorage === "undefined") return;

  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(cachedSession));
  } catch (e) {
    console.warn("[session] Failed to save:", e);
  }
}

export function debouncedSaveSession(
  partial: Partial<PersistedSession>,
  delayMs = 500,
): void {
  pendingPartial = { ...pendingPartial, ...partial };
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(() => {
    saveTimer = null;
    const toSave = pendingPartial;
    pendingPartial = {};
    saveSession(toSave);
  }, delayMs);
}

export function flushSessionSave(partial?: Partial<PersistedSession>): void {
  if (partial) pendingPartial = { ...pendingPartial, ...partial };
  if (saveTimer) {
    clearTimeout(saveTimer);
    saveTimer = null;
  }
  if (Object.keys(pendingPartial).length > 0) {
    const toSave = pendingPartial;
    pendingPartial = {};
    saveSession(toSave);
    return;
  }
  if (typeof localStorage !== "undefined") {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(cachedSession));
    } catch (e) {
      console.warn("[session] Failed to flush:", e);
    }
  }
}

export function clearLibrarySession(): void {
  saveSession({
    libraryPath: "",
    listOffset: 0,
    selectedTrackPath: null,
    playback: null,
  });
}
