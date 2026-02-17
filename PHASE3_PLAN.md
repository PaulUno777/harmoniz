# Phase 3: Infinite Scroll Implementation Plan (No Virtualization)

## Decision: Skip Virtualization

**Why no virtualization (`@tanstack/svelte-virtual`)?**

1. **Offline app**: No network latency, so incremental loading is fast
2. **Incremental loading**: Loading 100 tracks at a time means max ~500-1000 DOM nodes, which is perfectly fine for modern browsers
3. **Design compatibility**: Current grid layout with hover effects works better without virtualization
4. **Simplicity**: Less complexity, easier to maintain, fewer edge cases

**When virtualization would be needed:**
- Loading 10,000+ tracks at once
- Network latency causing slow loads
- Very complex row rendering (heavy components)

**Our approach**: Simple infinite scroll that loads more when user scrolls to ~80% of content.

---

## Implementation Plan

### 1. State Management (App.svelte)

**New state variables:**
```typescript
let isLoadingMore = $state(false);  // Separate from initial loading
let hasMoreTracks = $state(true);   // Whether more tracks exist
const PAGE_SIZE = 100;               // Tracks per page
```

**Modified `loadTracks()` function:**
- Rename to `loadInitialTracks()` - resets everything, loads first page
- New `loadMoreTracks()` - appends next page to existing array
- Track current offset: `let currentOffset = $state(0)`

### 2. Scroll Detection (LibraryContent.svelte)

**Add scroll listener:**
- Use `onMount` + `onDestroy` to attach/detach scroll listener
- Calculate scroll percentage: `(scrollTop + clientHeight) / scrollHeight`
- When scroll reaches **80%**, call `onLoadMore()` callback
- Use `requestAnimationFrame` or throttling to avoid excessive checks

**Scroll threshold:**
- Trigger at **80%** scroll position (user's preference)
- Add small debounce/throttle (e.g., 100ms) to prevent rapid-fire requests

### 3. Loading More Logic

**Flow:**
1. User scrolls to ~80%
2. Check: `!isLoadingMore && hasMoreTracks && tracks.length < totalTrackCount`
3. Set `isLoadingMore = true`
4. Call `ListTracks(path, PAGE_SIZE, currentOffset)`
5. Append new tracks: `tracks = [...tracks, ...newTracks]`
6. Update `currentOffset += PAGE_SIZE`
7. Check if more exist: `hasMoreTracks = tracks.length < totalTrackCount`
8. Set `isLoadingMore = false`

**Edge cases:**
- Prevent duplicate requests (guard with `isLoadingMore`)
- Stop when `tracks.length >= totalTrackCount`
- Handle errors gracefully (show message, allow retry)

### 4. UI Updates

**Loading indicator:**
- Show at bottom of track list when `isLoadingMore === true`
- Simple spinner or "Loading more..." text
- Only visible when loading more (not initial load)

**Empty state:**
- Keep existing empty state logic
- Show when `tracks.length === 0 && !isLoadingTracks`

**Initial loading:**
- Keep existing full-screen loading state
- Only show when `isLoadingTracks === true` (first load)

### 5. Reset Logic

**When library changes:**
- Reset `tracks = []`
- Reset `currentOffset = 0`
- Reset `hasMoreTracks = true`
- Reset `isLoadingMore = false`
- Call `loadInitialTracks()`

**When scan completes:**
- Same reset + reload first page

---

## File Changes

### `frontend/src/App.svelte`

**Changes:**
1. Add state: `isLoadingMore`, `hasMoreTracks`, `currentOffset`, `PAGE_SIZE`
2. Split `loadTracks()` → `loadInitialTracks()` (reset + first page)
3. Add `loadMoreTracks()` (append next page)
4. Pass `onLoadMore` callback to `LibraryContent`
5. Pass `hasMoreTracks` and `isLoadingMore` to `LibraryContent`

### `frontend/src/lib/components/layout/LibraryContent.svelte`

**Changes:**
1. Add props: `isLoadingMore`, `hasMoreTracks`, `onLoadMore`
2. Add scroll container ref: `let scrollContainer: HTMLDivElement`
3. Add scroll listener in `onMount` / `onDestroy`
4. Calculate scroll percentage and call `onLoadMore()` at 80%
5. Add loading indicator at bottom when `isLoadingMore === true`
6. Use throttling/debouncing for scroll handler

---

## Implementation Details

### Scroll Detection Pattern

```typescript
let scrollContainer: HTMLDivElement;
let scrollTimeout: number | null = null;

function handleScroll() {
  if (scrollTimeout) return; // Throttle
  
  scrollTimeout = window.setTimeout(() => {
    const { scrollTop, scrollHeight, clientHeight } = scrollContainer;
    const scrollPercent = (scrollTop + clientHeight) / scrollHeight;
    
    if (scrollPercent >= 0.8 && hasMoreTracks && !isLoadingMore) {
      onLoadMore();
    }
    
    scrollTimeout = null;
  }, 100);
}

onMount(() => {
  scrollContainer.addEventListener('scroll', handleScroll);
});

onDestroy(() => {
  scrollContainer?.removeEventListener('scroll', handleScroll);
  if (scrollTimeout) clearTimeout(scrollTimeout);
});
```

### Load More Function

```typescript
async function loadMoreTracks() {
  if (isLoadingMore || !hasMoreTracks || tracks.length >= totalTrackCount) {
    return;
  }
  
  isLoadingMore = true;
  try {
    const result = await ListTracks(currentLibraryPath, PAGE_SIZE, tracks.length);
    if (result?.Tracks?.length) {
      const newTracks = result.Tracks.map((t) => ({
        ...t,
        artist: t.artist_raw || "",
        album: t.album_raw || "",
      }));
      tracks = [...tracks, ...newTracks];
      hasMoreTracks = tracks.length < (result.Total ?? 0);
    } else {
      hasMoreTracks = false;
    }
  } catch (error) {
    console.error("Failed to load more tracks:", error);
    // Optionally show error message to user
  } finally {
    isLoadingMore = false;
  }
}
```

---

## Validation Criteria

After implementation, verify:

1. **Initial load**
   - [ ] First 100 tracks load correctly
   - [ ] Total count displayed correctly

2. **Scroll detection**
   - [ ] Scrolling to ~80% triggers load more
   - [ ] No duplicate requests (check network/console)
   - [ ] Scroll handler is throttled (no excessive calls)

3. **Loading more**
   - [ ] Next page loads and appends to list
   - [ ] Loading indicator appears at bottom
   - [ ] Tracks accumulate correctly (no duplicates)

4. **Completion**
   - [ ] When all tracks loaded, no more requests
   - [ ] `hasMoreTracks` becomes false
   - [ ] Loading indicator disappears

5. **Reset behavior**
   - [ ] Changing library resets and loads first page
   - [ ] Scanning resets and loads first page
   - [ ] State variables reset correctly

6. **Performance**
   - [ ] Smooth scrolling with 500+ tracks
   - [ ] No UI freezing during load
   - [ ] Memory usage reasonable

7. **Error handling**
   - [ ] Failed loads don't crash app
   - [ ] User can retry (scroll again)
   - [ ] Error logged to console

---

## Benefits of This Approach

✅ **Simple**: No complex virtualization library  
✅ **Maintainable**: Easy to understand and debug  
✅ **Design-friendly**: Works with current grid layout  
✅ **Performant**: Good enough for offline app with incremental loading  
✅ **User-friendly**: Smooth infinite scroll experience  

---

## Next Steps After Approval

1. Implement state management in `App.svelte`
2. Add scroll detection to `LibraryContent.svelte`
3. Add loading indicator UI
4. Test with various library sizes
5. Verify all validation criteria

---

## Notes

- **Page size**: Keep at 100 (good balance)
- **Scroll threshold**: 80% (can be adjusted if needed)
- **Throttle**: 100ms (prevents excessive checks)
- **Future optimization**: If performance becomes an issue with 2000+ tracks, consider virtualization then
