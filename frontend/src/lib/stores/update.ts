import { writable, derived, get } from "svelte/store";
import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime.js";
import { CheckForUpdates, GetAppVersion } from "../../../wailsjs/go/main/App.js";
import { loadSession, saveSession } from "./session";

export type UpdateStatus = "idle" | "checking" | "ready" | "error";

export interface UpdateState {
  currentVersion: string;
  latestVersion: string;
  updateAvailable: boolean;
  downloadUrl: string;
  releasePageUrl: string;
  status: UpdateStatus;
  lastCheckedAt: string | null;
  errorMessage: string | null;
}

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

async function init() {
  try {
    const version = await GetAppVersion();
    setState({ currentVersion: version ?? "" });
  } catch (e) {
    console.warn("[update] failed to load app version:", e);
  }
}

async function check(_silent = false) {
  setState({ status: "checking", errorMessage: null });
  try {
    const result = await CheckForUpdates();
    setState({
      currentVersion: result.currentVersion ?? "",
      latestVersion: result.latestVersion ?? "",
      updateAvailable: !!result.updateAvailable,
      downloadUrl: result.downloadUrl ?? "",
      releasePageUrl: result.releasePageUrl ?? "",
      status: "ready",
      lastCheckedAt: new Date().toISOString(),
      errorMessage: null,
    });
  } catch (e) {
    const message = e instanceof Error ? e.message : String(e);
    setState({
      status: "error",
      errorMessage: message,
      lastCheckedAt: new Date().toISOString(),
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
};
