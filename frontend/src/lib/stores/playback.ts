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

function createPlaybackStore() {
  const { subscribe, set, update } = writable<PlaybackState>({
    currentTrack: null,
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    volume: 1,
    playlist: [],
    currentIndex: -1,
    isMuted: false,
  });

  let audioElement: HTMLAudioElement | null = null;

  function initAudio() {
    if (audioElement) return audioElement;

    audioElement = new Audio();
    audioElement.volume = 1;

    audioElement.addEventListener("timeupdate", () => {
      update((state) => ({
        ...state,
        currentTime: audioElement?.currentTime || 0,
      }));
    });

    audioElement.addEventListener("loadedmetadata", () => {
      update((state) => ({
        ...state,
        duration: audioElement?.duration || 0,
      }));
    });

    audioElement.addEventListener("ended", () => {
      next();
    });

    audioElement.addEventListener("play", () => {
      update((state) => ({ ...state, isPlaying: true }));
    });

    audioElement.addEventListener("pause", () => {
      update((state) => ({ ...state, isPlaying: false }));
    });

    return audioElement;
  }

  function play(track: Track, playlist?: Track[]) {
    const audio = initAudio();

    if (playlist && playlist.length > 0) {
      const index = playlist.findIndex((t) => t.path === track.path);
      update((state) => ({
        ...state,
        playlist: playlist,
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

    // Stream via Go backend (supports Range for seeking)
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
    update((state) => {
      if (state.isPlaying) {
        pause();
      } else {
        resume();
      }
      return state;
    });
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
    update((state) => {
      if (audioElement) {
        if (state.isMuted) {
          audioElement.volume = state.volume || 0.5;
          return { ...state, isMuted: false };
        } else {
          audioElement.volume = 0;
          return { ...state, isMuted: true };
        }
      }
      return state;
    });
  }

  function next() {
    update((state) => {
      if (state.playlist.length === 0 || state.currentIndex < 0) return state;

      const nextIndex = (state.currentIndex + 1) % state.playlist.length;
      const nextTrack = state.playlist[nextIndex];

      if (nextTrack) {
        play(nextTrack, state.playlist);
      }

      return state;
    });
  }

  function previous() {
    update((state) => {
      if (state.playlist.length === 0 || state.currentIndex < 0) return state;

      const prevIndex =
        state.currentIndex === 0
          ? state.playlist.length - 1
          : state.currentIndex - 1;
      const prevTrack = state.playlist[prevIndex];

      if (prevTrack) {
        play(prevTrack, state.playlist);
      }

      return state;
    });
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
    if (audioElement) {
      audioElement.pause();
      audioElement.src = "";
      audioElement = null;
    }
    set({
      currentTrack: null,
      isPlaying: false,
      currentTime: 0,
      duration: 0,
      volume: 1,
      playlist: [],
      currentIndex: -1,
      isMuted: false,
    });
  }

  return {
    subscribe,
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
