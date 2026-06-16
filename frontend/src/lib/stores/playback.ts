import { writable } from "svelte/store";
import type { Track } from "../types";

interface PlaybackState {
  currentTrack: Track | null;
  isPlaying: boolean;
  currentTime: number;
  duration: number;
  volume: number;
  playlist: Track[];
  currentIndex: number;
  isMuted: boolean;
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
};

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

  function play(track: Track, playlist?: Track[]) {
    const audio = initAudio();

    if (playlist && playlist.length > 0) {
      const index = playlist.findIndex((t) => t.path === track.path);
      update((state) => ({
        ...state,
        playlist,
        currentIndex: index >= 0 ? index : 0,
        currentTrack: track,
      }));
    } else {
      update((state) => ({
        ...state,
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
  }

  const RESTART_THRESHOLD_SEC = 3;

  function next() {
    const state = getState();
    if (state.playlist.length === 0 || state.currentIndex < 0) return;

    const nextIndex =
      state.currentIndex === state.playlist.length - 1
        ? 0
        : state.currentIndex + 1;
    const nextTrack = state.playlist[nextIndex];
    if (nextTrack) {
      play(nextTrack, state.playlist);
    }
  }

  function previous() {
    const state = getState();
    if (state.playlist.length === 0 || state.currentIndex < 0) return;

    const currentTime = state.currentTime;
    if (currentTime > RESTART_THRESHOLD_SEC && audioElement) {
      seek(0);
      return;
    }

    const prevIndex =
      state.currentIndex === 0
        ? state.playlist.length - 1
        : state.currentIndex - 1;
    const prevTrack = state.playlist[prevIndex];
    if (prevTrack) {
      play(prevTrack, state.playlist);
    }
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
    play,
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
  };
}

export const playbackStore = createPlaybackStore();
