---
name: Floating player component
overview: Create a modern floating player component that appears above the status bar when music is playing. Includes play/pause, volume control, progress bar, and time displays (current/total/remaining).
todos:
  - id: "1"
    content: Create playback store (frontend/src/lib/stores/playback.ts) with audio element and state management
    status: completed
  - id: "2"
    content: Add formatTime utility function to frontend/src/lib/utils/format.ts
    status: completed
  - id: "3"
    content: Create FloatingPlayer component with track info, controls, progress bar, and volume
    status: completed
  - id: "4"
    content: Integrate FloatingPlayer into App.svelte above StatusBar
    status: completed
  - id: "5"
    content: Connect track selection in LibraryContent/TrackItem to playback store
    status: completed
  - id: "6"
    content: Add styling, animations, and polish the floating effect
    status: completed
isProject: false
---

# Floating Player Component

## Overview

Create a floating music player component that appears above the status bar when music is playing. The player will have a modern design with a floating effect (backdrop blur, shadow, rounded corners) and include all essential playback controls.

## Architecture

### 1. Playback Store (`frontend/src/lib/stores/playback.ts`)

Create a new Svelte store to manage playback state:

- **State:**
  - `currentTrack: Track | null` - Currently playing track
  - `isPlaying: boolean` - Playback state
  - `currentTime: number` - Current playback time in seconds
  - `duration: number` - Total track duration in seconds
  - `volume: number` - Volume level (0-1)
  - `playlist: Track[]` - Queue of tracks to play
  - `currentIndex: number` - Index in playlist
- **Methods:**
  - `play(track: Track, playlist?: Track[])` - Start playing a track
  - `pause()` - Pause playback
  - `resume()` - Resume playback
  - `toggle()` - Toggle play/pause
  - `seek(time: number)` - Seek to specific time
  - `setVolume(volume: number)` - Set volume (0-1)
  - `next()` - Play next track
  - `previous()` - Play previous track
- **Audio Element:**
  - Use HTML5 `<audio>` element
  - Handle `timeupdate`, `loadedmetadata`, `ended` events
  - Update state reactively

### 2. FloatingPlayer Component (`frontend/src/lib/components/layout/FloatingPlayer.svelte`)

**Props:**

- `activeTab: TabId` - Current active tab (only show on "library")
- `isVisible: boolean` - Whether player should be visible (derived from playback store)

**Layout:**

- Fixed position above status bar (bottom: 32px or similar)
- Centered horizontally with max-width constraint
- Modern floating design:
  - Backdrop blur (`backdrop-blur-md` or `backdrop-blur-lg`)
  - Semi-transparent background (`bg-surface/90` or `bg-background/95`)
  - Rounded corners (`rounded-2xl` or `rounded-3xl`)
  - Shadow (`shadow-2xl shadow-black/20`)
  - Border (`border border-border/50`)

**Components:**

1. **Track Info Section (Left)**
  - Track title (truncated)
  - Artist name (truncated)
  - Small album art placeholder or icon
2. **Playback Controls (Center)**
  - Previous track button
  - Play/Pause button (larger, prominent)
  - Next track button
  - Progress bar (seekable)
  - Time display: `currentTime / totalTime` and `-timeLeft`
3. **Volume Control (Right)**
  - Volume icon (with mute state)
  - Volume slider (horizontal)
  - Click icon to toggle mute

**Styling:**

- Use Phosphor icons for all buttons
- Smooth transitions and hover effects
- Responsive layout (flexbox)
- Ensure it doesn't overlap with status bar

### 3. Integration Points

**App.svelte:**

- Import `FloatingPlayer` component
- Import playback store
- Add `<FloatingPlayer>` component above `<StatusBar>` in the footer section
- Pass `activeTab` prop
- Conditionally render based on `playbackStore.currentTrack !== null`

**TrackItem.svelte / LibraryContent.svelte:**

- When track is clicked/selected, call `playbackStore.play(track, tracks)` to start playback
- This will automatically show the floating player

**StatusBar.svelte:**

- Ensure status bar has proper spacing (padding-bottom) to accommodate floating player above it

### 4. Time Formatting Utility

Create or update `frontend/src/lib/utils/format.ts`:

- Add `formatTime(seconds: number): string` function
- Format as `MM:SS` or `H:MM:SS` for longer tracks
- Handle NaN/undefined gracefully

### 5. Visual Design Details

**Floating Effect:**

- Use CSS transforms for subtle animations
- Consider adding a subtle pulse/glow effect when playing
- Smooth slide-in/slide-out animations

**Color Scheme:**

- Match existing theme colors
- Use accent color for active states (play button, progress bar)
- Ensure good contrast for readability

**Responsive Behavior:**

- On smaller screens, stack elements vertically or hide less critical info
- Ensure controls remain accessible

## File Structure

```
frontend/src/lib/
├── stores/
│   └── playback.ts (new)
├── components/
│   └── layout/
│       └── FloatingPlayer.svelte (new)
└── utils/
    └── format.ts (update - add formatTime)
```

## Implementation Steps

1. Create playback store with audio element management
2. Create FloatingPlayer component with all controls
3. Add time formatting utility
4. Integrate into App.svelte
5. Connect track selection to playback
6. Style and polish with animations
7. Test playback functionality

## Notes

- The player should only appear when `activeTab === "library"` AND `currentTrack !== null`
- Use Svelte 5 runes (`$state`, `$derived`) for reactivity
- Ensure audio element is properly cleaned up on component destroy
- Handle edge cases (no duration, loading state, etc.)

