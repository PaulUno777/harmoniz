import { writable } from "svelte/store";
import type { Track } from "../types";
import {
  debouncedSaveSession,
  flushSessionSave,
  type LoopMode,
  type PersistedPlayback,
} from "./session";

export type { LoopMode };

interface PlaybackState {
  currentTrack: Track | null;
  isPlaying: boolean;
  currentTime: number;
  duration: number;
  volume: number;
  playlist: Track[];
  currentIndex: number;
  isMuted: boolean;
  loopMode: LoopMode;
  shuffleMode: boolean;
  shuffledPlaylist: Track[];
}

const initialState: PlaybackState = {
  currentTrack: null,
  isPlaying: false,
  currentTime: 0,
  duration: 0,
  volume: 0.8,
  playlist: [],
  currentIndex: -1,
  isMuted: false,
  loopMode: 'none',
  shuffleMode: false,
  shuffledPlaylist: [],
};

function smartShuffle(tracks: Track[], currentTrack: Track | null): Track[] {
  const arr = [...tracks]
  // Fisher-Yates
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [arr[i], arr[j]] = [arr[j], arr[i]]
  }
  // Post-process: spread same-artist tracks so they're not consecutive
  for (let i = 0; i < arr.length - 1; i++) {
    const a = arr[i], b = arr[i + 1]
    if (a.artist_raw && b.artist_raw && a.artist_raw === b.artist_raw) {
      for (let j = i + 2; j < arr.length; j++) {
        if (!arr[j].artist_raw || arr[j].artist_raw !== a.artist_raw) {
          ;[arr[i + 1], arr[j]] = [arr[j], arr[i + 1]]
          break
        }
      }
    }
  }
  // Put current track first so it continues playing
  if (currentTrack) {
    const idx = arr.findIndex(t => t.path === currentTrack.path)
    if (idx > 0) [arr[0], arr[idx]] = [arr[idx], arr[0]]
  }
  return arr
}

function createPlaybackStore() {
  const { subscribe, set: origSet, update: origUpdate } =
    writable<PlaybackState>({ ...initialState });

  let currentState: PlaybackState = { ...initialState };

  function set(value: PlaybackState) {
    currentState = value;
    origSet(value);
  }

  function update(fn: (state: PlaybackState) => PlaybackState) {
    origUpdate((state) => {
      const next = fn(state);
      currentState = next;
      return next;
    });
  }

  function getState(): PlaybackState {
    return currentState;
  }

  let audioElement: HTMLAudioElement | null = null;
  let abortController: AbortController | null = null;

  function getPersistedPlayback(): PersistedPlayback | null {
    const state = getState();
    if (!state.currentTrack?.path) return null;
    return {
      trackPath: state.currentTrack.path,
      currentTime: state.currentTime,
      volume: state.volume,
      isMuted: state.isMuted,
      loopMode: state.loopMode,
      shuffleMode: state.shuffleMode,
    };
  }

  function schedulePlaybackPersist() {
    const playback = getPersistedPlayback();
    if (!playback) return;
    debouncedSaveSession({ playback }, 2000);
  }

  function persistPlaybackNow() {
    const playback = getPersistedPlayback();
    if (playback) flushSessionSave({ playback });
  }

  function initAudio() {
    if (audioElement) return audioElement;

    abortController = new AbortController();
    const signal = abortController.signal;

    audioElement = new Audio();
    audioElement.volume = 1;

    audioElement.addEventListener(
      "timeupdate",
      () => {
        update((state) => ({
          ...state,
          currentTime: audioElement?.currentTime ?? 0,
        }));
        schedulePlaybackPersist();
      },
      { signal },
    );

    audioElement.addEventListener(
      "loadedmetadata",
      () => {
        update((state) => ({
          ...state,
          duration: audioElement?.duration ?? 0,
        }));
      },
      { signal },
    );

    audioElement.addEventListener(
      "ended",
      () => {
        next();
      },
      { signal },
    );

    audioElement.addEventListener(
      "play",
      () => {
        update((state) => ({ ...state, isPlaying: true }));
      },
      { signal },
    );

    audioElement.addEventListener(
      "pause",
      () => {
        update((state) => ({ ...state, isPlaying: false }));
      },
      { signal },
    );

    return audioElement;
  }

  // Internal helper: start playing a track at a specific index without resetting playlist.
  // Used by next() / previous() so shuffledPlaylist and playlist are preserved.
  function playTrackAt(index: number, activePlaylist: Track[]) {
    const track = activePlaylist[index]
    if (!track) return
    update(s => ({ ...s, currentTrack: track, currentIndex: index }))
    const audio = initAudio()
    const streamUrl = `/stream?path=${encodeURIComponent(track.path)}`
    audio.src = streamUrl
    audio.load()
    audio.play().catch((e) => console.error("Failed to play audio:", e))
  }

  function restore(
    track: Track,
    playlist: Track[],
    opts: Omit<PersistedPlayback, "trackPath">,
  ) {
    const audio = initAudio();
    const index = playlist.findIndex((t) => t.path === track.path);
    const activeIndex = index >= 0 ? index : 0;

    update((s) => ({
      ...s,
      currentTrack: track,
      playlist: playlist.length > 0 ? playlist : [track],
      currentIndex: activeIndex,
      isPlaying: false,
      currentTime: opts.currentTime,
      volume: opts.volume,
      isMuted: opts.isMuted,
      loopMode: opts.loopMode,
      shuffleMode: opts.shuffleMode,
      shuffledPlaylist: opts.shuffleMode
        ? smartShuffle(playlist.length > 0 ? playlist : [track], track)
        : [],
    }));

    audio.volume = opts.isMuted ? 0 : opts.volume;
    const streamUrl = `/stream?path=${encodeURIComponent(track.path)}`;
    audio.src = streamUrl;
    audio.load();

    const seekTime = opts.currentTime;
    const onLoaded = () => {
      if (seekTime > 0) {
        audio.currentTime = Math.min(seekTime, audio.duration || seekTime);
      }
      update((s) => ({
        ...s,
        currentTime: audio.currentTime,
        duration: audio.duration || 0,
      }));
      audio.removeEventListener("loadedmetadata", onLoaded);
    };
    audio.addEventListener("loadedmetadata", onLoaded);
  }

  function play(track: Track, playlist?: Track[]) {
    const audio = initAudio();

    if (playlist && playlist.length > 0) {
      const index = playlist.findIndex((t) => t.path === track.path);
      const state = getState()
      let shuffledPlaylist = state.shuffledPlaylist
      let activeIndex = index >= 0 ? index : 0
      // Rebuild shuffle with new playlist if shuffle is already on
      if (state.shuffleMode) {
        shuffledPlaylist = smartShuffle(playlist, track)
        activeIndex = 0
      }
      update((s) => ({
        ...s,
        playlist,
        currentIndex: activeIndex,
        currentTrack: track,
        shuffledPlaylist,
      }));
    } else {
      update((s) => ({
        ...s,
        currentTrack: track,
        playlist: [track],
        currentIndex: 0,
      }));
    }

    const streamUrl = `/stream?path=${encodeURIComponent(track.path)}`;
    audio.src = streamUrl;
    audio.load();
    audio.play().catch((error) => {
      console.error("Failed to play audio:", error);
    });
    persistPlaybackNow();
  }

  function pause() {
    if (audioElement) {
      audioElement.pause();
    }
  }

  function resume() {
    if (audioElement) {
      audioElement.play().catch((error) => {
        console.error("Failed to resume audio:", error);
      });
    }
  }

  function toggle() {
    const state = getState();
    if (state.isPlaying) {
      pause();
    } else {
      resume();
    }
  }

  function seek(time: number) {
    if (audioElement) {
      audioElement.currentTime = Math.max(
        0,
        Math.min(time, audioElement.duration || 0),
      );
    }
  }

  function setVolume(volume: number) {
    const clampedVolume = Math.max(0, Math.min(1, volume));
    if (audioElement) {
      audioElement.volume = clampedVolume;
    }
    update((state) => ({
      ...state,
      volume: clampedVolume,
      isMuted: clampedVolume === 0 ? true : state.isMuted,
    }));
    persistPlaybackNow();
  }

  function toggleMute() {
    const state = getState();
    if (!audioElement) return;
    if (state.isMuted) {
      audioElement.volume = state.volume || 0.5;
      update((s) => ({ ...s, isMuted: false }));
    } else {
      audioElement.volume = 0;
      update((s) => ({ ...s, isMuted: true }));
    }
    persistPlaybackNow();
  }

  const RESTART_THRESHOLD_SEC = 3;

  function next() {
    const state = getState();
    const activePlaylist = state.shuffleMode && state.shuffledPlaylist.length > 0
      ? state.shuffledPlaylist : state.playlist
    if (activePlaylist.length === 0 || state.currentIndex < 0) return;

    if (state.loopMode === 'one') { seek(0); resume(); return }

    const atEnd = state.currentIndex === activePlaylist.length - 1
    if (atEnd && state.loopMode === 'none') { stop(); return }

    playTrackAt(atEnd ? 0 : state.currentIndex + 1, activePlaylist)
  }

  function previous() {
    const state = getState();
    if (state.currentTime > RESTART_THRESHOLD_SEC && audioElement) {
      seek(0);
      return;
    }

    const activePlaylist = state.shuffleMode && state.shuffledPlaylist.length > 0
      ? state.shuffledPlaylist : state.playlist
    if (activePlaylist.length === 0 || state.currentIndex < 0) return;

    const prevIndex = state.currentIndex === 0
      ? activePlaylist.length - 1
      : state.currentIndex - 1
    playTrackAt(prevIndex, activePlaylist)
  }

  function stop() {
    if (audioElement) {
      audioElement.pause();
      audioElement.currentTime = 0;
    }
    update((state) => ({
      ...state,
      isPlaying: false,
      currentTime: 0,
    }));
  }

  function cycleLoop() {
    update(s => {
      const next: LoopMode = s.loopMode === 'none' ? 'all' : s.loopMode === 'all' ? 'one' : 'none'
      return { ...s, loopMode: next }
    })
    persistPlaybackNow();
  }

  function toggleShuffle() {
    const state = getState()
    if (state.shuffleMode) {
      const idx = state.playlist.findIndex(t => t.path === state.currentTrack?.path)
      update(s => ({ ...s, shuffleMode: false, shuffledPlaylist: [], currentIndex: idx >= 0 ? idx : 0 }))
    } else {
      const shuffled = smartShuffle(state.playlist, state.currentTrack)
      update(s => ({ ...s, shuffleMode: true, shuffledPlaylist: shuffled, currentIndex: 0 }))
    }
    persistPlaybackNow();
  }

  function cleanup() {
    if (abortController) {
      abortController.abort();
      abortController = null;
    }
    if (audioElement) {
      audioElement.pause();
      audioElement.src = "";
      audioElement = null;
    }
    currentState = { ...initialState };
    set({ ...initialState });
  }

  return {
    subscribe,
    getState,
    getPersistedPlayback,
    persistPlaybackNow,
    play,
    restore,
    pause,
    resume,
    toggle,
    seek,
    setVolume,
    toggleMute,
    next,
    previous,
    stop,
    cleanup,
    cycleLoop,
    toggleShuffle,
  };
}

export const playbackStore = createPlaybackStore();
