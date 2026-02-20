<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    PlayIcon,
    PauseIcon,
    SkipBackIcon,
    SkipForwardIcon,
    SpeakerHighIcon,
    SpeakerXIcon,
    FileAudioIcon,
  } from "phosphor-svelte";
  import { playbackStore } from "../../stores/playback";
  import { formatTime } from "../../utils/format";
  import type { Track } from "../../types";
  import type { TabId } from "../../types";

  interface Props {
    activeTab: TabId;
  }

  let { activeTab }: Props = $props();

  type PlaybackSnapshot = {
    currentTrack: Track | null;
    isPlaying: boolean;
    currentTime: number;
    duration: number;
    volume: number;
    isMuted: boolean;
  };

  let playbackState = $state<PlaybackSnapshot>({
    currentTrack: null,
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    volume: 1,
    isMuted: false,
  });

  let isDraggingProgress = $state(false);

  let unsubscribe: (() => void) | null = null;

  onMount(() => {
    unsubscribe = playbackStore.subscribe((state) => {
      playbackState = {
        currentTrack: state.currentTrack,
        isPlaying: state.isPlaying,
        currentTime: state.currentTime,
        duration: state.duration,
        volume: state.volume,
        isMuted: state.isMuted,
      };
    });
  });

  onDestroy(() => {
    if (unsubscribe) {
      unsubscribe();
    }
  });

  const isVisible = $derived(
    activeTab === "library" && playbackState.currentTrack !== null,
  );

  const timeLeft = $derived(
    playbackState.duration > 0
      ? playbackState.duration - playbackState.currentTime
      : 0,
  );

  const progressPercent = $derived(
    playbackState.duration > 0
      ? (playbackState.currentTime / playbackState.duration) * 100
      : 0,
  );

  function handlePlayPause() {
    playbackStore.toggle();
  }

  function handlePrevious() {
    playbackStore.previous();
  }

  function handleNext() {
    playbackStore.next();
  }

  function handleProgressChange(e: Event) {
    const target = e.target as HTMLInputElement;
    const value = parseFloat(target.value);
    playbackStore.seek(value);
  }

  function handleProgressMouseDown() {
    isDraggingProgress = true;
  }

  function handleProgressMouseUp() {
    isDraggingProgress = false;
  }

  function handleVolumeChange(e: Event) {
    const target = e.target as HTMLInputElement;
    const value = parseFloat(target.value) / 100;
    playbackStore.setVolume(value);
  }

  function handleVolumeClick() {
    playbackStore.toggleMute();
  }
</script>

{#if isVisible && playbackState.currentTrack}
  <div
    class="absolute bottom-4 left-0 right-0 z-50 pointer-events-none transition-all duration-300"
    style="animation: slideUp 0.3s ease-out;"
  >
    <div class="max-w-5xl mx-auto px-8">
      <div
        class="bg-surface/95 backdrop-blur-lg border border-border/50 rounded-2xl shadow-2xl shadow-black/20 p-4 pointer-events-auto transition-all duration-300
             {playbackState.isPlaying
          ? 'ring-1 ring-accent/30 shadow-accent/10'
          : ''}"
      >
        <!-- 3-column grid keeps center controls truly centered -->
        <div class="grid grid-cols-[1fr_auto_1fr] items-center gap-4">
          <!-- Track Info (Left) -->
          <div
            class="flex items-center gap-3 min-w-0 max-w-[320px] overflow-hidden justify-self-start"
          >
            <div
              class="w-12 h-12 bg-background rounded-lg flex items-center justify-center border border-border shrink-0
                   {playbackState.isPlaying ? 'animate-pulse' : ''}"
            >
              <FileAudioIcon size={24} weight="duotone" class="text-accent" />
            </div>
            <div class="min-w-0 flex-1">
              <div
                class="font-bold text-text-primary truncate text-sm"
                title={playbackState.currentTrack.title}
              >
                {playbackState.currentTrack.title}
              </div>
              <div
                class="text-xs text-text-secondary truncate"
                title={playbackState.currentTrack.artist || "Unknown Artist"}
              >
                {playbackState.currentTrack.artist || "Unknown Artist"}
              </div>
            </div>
          </div>

          <!-- Playback Controls (Center) -->
          <div
            class="w-[min(520px,100%)] flex flex-col items-center gap-2 min-w-0 justify-self-center"
          >
            <div class="flex items-center gap-2">
              <button
                onclick={handlePrevious}
                class="p-2 hover:bg-white/5 rounded-lg text-text-secondary hover:text-accent transition-all"
                title="Previous track"
              >
                <SkipBackIcon size={20} weight="fill" />
              </button>

              <button
                onclick={handlePlayPause}
                class="w-12 h-12 bg-accent hover:bg-accent-hover text-background rounded-full flex items-center justify-center transition-all shadow-lg shadow-accent/30 hover:scale-105 active:scale-95"
                title={playbackState.isPlaying ? "Pause" : "Play"}
              >
                {#if playbackState.isPlaying}
                  <PauseIcon size={24} weight="fill" />
                {:else}
                  <PlayIcon size={24} weight="fill" />
                {/if}
              </button>

              <button
                onclick={handleNext}
                class="p-2 hover:bg-white/5 rounded-lg text-text-secondary hover:text-accent transition-all"
                title="Next track"
              >
                <SkipForwardIcon size={20} weight="fill" />
              </button>
            </div>

            <!-- Progress Bar -->
            <div class="w-full flex items-center gap-2">
              <span
                class="text-[10px] font-mono text-text-muted tabular-nums shrink-0"
              >
                {formatTime(playbackState.currentTime)}
              </span>
              <input
                type="range"
                min="0"
                max={playbackState.duration || 0}
                value={playbackState.currentTime}
                oninput={handleProgressChange}
                onmousedown={handleProgressMouseDown}
                onmouseup={handleProgressMouseUp}
                class="flex-1 h-1 bg-border rounded-full appearance-none cursor-pointer accent-accent
                     {isDraggingProgress ? '' : 'hover:accent-accent-hover'}"
                style="background: linear-gradient(to right, var(--color-accent) 0%, var(--color-accent) {progressPercent}%, var(--color-border) {progressPercent}%, var(--color-border) 100%);"
              />
              <div class="flex items-center gap-1 shrink-0">
                <span
                  class="text-[10px] font-mono text-text-muted tabular-nums"
                >
                  {formatTime(playbackState.duration)}
                </span>
                <span
                  class="text-[10px] font-mono text-text-muted tabular-nums"
                >
                  / -{formatTime(timeLeft)}
                </span>
              </div>
            </div>
          </div>

          <!-- Volume Control (Right) -->
          <div class="flex items-center gap-2 min-w-0 justify-self-end">
            <button
              onclick={handleVolumeClick}
              class="p-2 hover:bg-white/5 rounded-lg text-text-secondary hover:text-accent transition-all shrink-0"
              title={playbackState.isMuted ? "Unmute" : "Mute"}
            >
              {#if playbackState.isMuted || playbackState.volume === 0}
                <SpeakerXIcon size={20} weight="fill" />
              {:else}
                <SpeakerHighIcon size={20} weight="fill" />
              {/if}
            </button>
            <input
              type="range"
              min="0"
              max="100"
              value={playbackState.isMuted ? 0 : playbackState.volume * 100}
              oninput={handleVolumeChange}
              class="w-20 h-1 bg-border rounded-full appearance-none cursor-pointer accent-accent hover:accent-accent-hover"
              style="background: linear-gradient(to right, var(--color-accent) 0%, var(--color-accent) {playbackState.isMuted
                ? 0
                : playbackState.volume *
                  100}%, var(--color-border) {playbackState.isMuted
                ? 0
                : playbackState.volume * 100}%, var(--color-border) 100%);"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translate(-50%, 20px);
    }
    to {
      opacity: 1;
      transform: translate(-50%, 0);
    }
  }

  input[type="range"] {
    -webkit-appearance: none;
    appearance: none;
  }

  input[type="range"]::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--color-accent);
    cursor: pointer;
    border: 2px solid var(--color-background);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  }

  input[type="range"]::-webkit-slider-thumb:hover {
    background: var(--color-accent-hover);
    transform: scale(1.1);
  }

  input[type="range"]::-moz-range-thumb {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--color-accent);
    cursor: pointer;
    border: 2px solid var(--color-background);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  }

  input[type="range"]::-moz-range-thumb:hover {
    background: var(--color-accent-hover);
    transform: scale(1.1);
  }
</style>
