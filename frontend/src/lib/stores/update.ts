import { writable, derived, get } from "svelte/store";
import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime.js";
import {
  CheckForUpdates,
  GetAppVersion,
} from "../../../wailsjs/go/main/App.js";
import { loadSession, saveSession, type PersistedUpdateCheck } from "./session";

type UpdateStatus = "idle" | "checking" | "ready" | "error";

interface UpdateState {
  currentVersion: string;
  latestVersion: string;
  updateAvailable: boolean;
  downloadUrl: string;
  releasePageUrl: string;
  status: UpdateStatus;
  lastCheckedAt: string | null;
  errorMessage: string | null;
}

/** Minimum interval between successful GitHub release checks. */
const CHECK_TTL_MS = 60 * 60 * 1000;
/** Shorter retry window after a failed check (network/GitHub errors). */
const CHECK_ERROR_TTL_MS = 15 * 60 * 1000;

const initialState: UpdateState = {
  currentVersion: "",
  latestVersion: "",
  updateAvailable: false,
  downloadUrl: "",
  releasePageUrl: "",
  status: "idle",
  lastCheckedAt: null,
  errorMessage: null,
};

const state = writable<UpdateState>({ ...initialState });
const dismissedVersion = writable<string | undefined>(
  loadSession().dismissedUpdateVersion,
);

let checkInFlight: Promise<void> | null = null;

export const showUpdatePrompt = derived(
  [state, dismissedVersion],
  ([$state, $dismissed]) =>
    $state.updateAvailable &&
    !!$state.latestVersion &&
    $dismissed !== $state.latestVersion,
);

function setState(partial: Partial<UpdateState>) {
  state.update((s) => ({ ...s, ...partial }));
}

function formatError(e: unknown): string {
  if (e instanceof Error) return e.message;
  if (typeof e === "string") return e;
  try {
    return JSON.stringify(e);
  } catch {
    return String(e);
  }
}

export function isCheckCacheFresh(
  lastCheckedAt: string | null,
  status?: UpdateStatus,
): boolean {
  if (!lastCheckedAt) return false;
  const age = Date.now() - new Date(lastCheckedAt).getTime();
  if (age < 0) return false;
  const ttl = status === "error" ? CHECK_ERROR_TTL_MS : CHECK_TTL_MS;
  return age < ttl;
}

function loadCachedCheck(): PersistedUpdateCheck | undefined {
  return loadSession().updateCheckCache;
}

function applyCachedCheck(cached: PersistedUpdateCheck) {
  setState({
    latestVersion: cached.latestVersion,
    updateAvailable: cached.updateAvailable,
    downloadUrl: cached.downloadUrl,
    releasePageUrl: cached.releasePageUrl,
    status: cached.status,
    lastCheckedAt: cached.checkedAt,
    errorMessage: cached.errorMessage ?? null,
  });
}

function persistCheck(partial: Omit<PersistedUpdateCheck, "checkedAt">) {
  const checkedAt = new Date().toISOString();
  saveSession({
    updateCheckCache: {
      checkedAt,
      ...partial,
    },
  });
  return checkedAt;
}

async function init() {
  try {
    const version = await GetAppVersion();
    setState({ currentVersion: version ?? "" });
  } catch (e) {
    console.warn("[update] failed to load app version:", e);
  }

  const cached = loadCachedCheck();
  if (cached) {
    applyCachedCheck(cached);
  }
}

async function check(options: { silent?: boolean; force?: boolean } = {}) {
  const { silent = false, force = false } = options;

  const cached = loadCachedCheck();
  if (!force && cached && isCheckCacheFresh(cached.checkedAt, cached.status)) {
    if (!silent) {
      console.info("[update] using cached check", {
        checkedAt: cached.checkedAt,
      });
    }
    applyCachedCheck(cached);
    return;
  }

  if (checkInFlight) {
    return checkInFlight;
  }

  checkInFlight = runCheck(silent).finally(() => {
    checkInFlight = null;
  });
  return checkInFlight;
}

async function runCheck(silent: boolean) {
  const current = get(state).currentVersion;
  if (!silent) {
    console.info("[update] checking for updates", { current });
  }
  setState({ status: "checking", errorMessage: null });
  try {
    const result = await CheckForUpdates();
    const checkedAt = persistCheck({
      latestVersion: result.latestVersion ?? "",
      updateAvailable: !!result.updateAvailable,
      downloadUrl: result.downloadUrl ?? "",
      releasePageUrl: result.releasePageUrl ?? "",
      status: "ready",
      errorMessage: null,
    });
    console.info("[update] check succeeded", {
      current: result.currentVersion,
      latest: result.latestVersion,
      updateAvailable: result.updateAvailable,
      downloadUrl: result.downloadUrl,
    });
    setState({
      currentVersion: result.currentVersion ?? "",
      latestVersion: result.latestVersion ?? "",
      updateAvailable: !!result.updateAvailable,
      downloadUrl: result.downloadUrl ?? "",
      releasePageUrl: result.releasePageUrl ?? "",
      status: "ready",
      lastCheckedAt: checkedAt,
      errorMessage: null,
    });
  } catch (e) {
    const message = formatError(e);
    const checkedAt = persistCheck({
      latestVersion: "",
      updateAvailable: false,
      downloadUrl: "",
      releasePageUrl: "",
      status: "error",
      errorMessage: message,
    });
    console.error("[update] check failed", { message, error: e });
    setState({
      status: "error",
      errorMessage: message,
      lastCheckedAt: checkedAt,
    });
  }
}

function openDownload() {
  const s = get(state);
  const url = s.downloadUrl || s.releasePageUrl;
  if (url) BrowserOpenURL(url);
}

function dismissUpdate() {
  const s = get(state);
  if (s.latestVersion) {
    saveSession({ dismissedUpdateVersion: s.latestVersion });
    dismissedVersion.set(s.latestVersion);
  }
}

function isDismissed(): boolean {
  const s = get(state);
  return !!s.latestVersion && get(dismissedVersion) === s.latestVersion;
}

export const updateStore = {
  subscribe: state.subscribe,
  init,
  check,
  openDownload,
  dismissUpdate,
  isDismissed,
  isCheckCacheFresh,
};
