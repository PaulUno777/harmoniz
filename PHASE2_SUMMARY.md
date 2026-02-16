# Phase 2: Folder Browsing & Scanning - COMPLETE

## Files Created

### Filesystem Adapter
- ✅ `internal/adapters/fs/fs.go` - Filesystem adapter with SafeMove
- ✅ `internal/adapters/fs/hash.go` - Hashing utilities (partial and full hash)
- ✅ `internal/core/ports/filesystem.go` - Filesystem interface

### Scanner Service
- ✅ `internal/core/services/scanner/scanner.go` - Scanner service with worker pool, plus `PrepareLibrarySync()` for staleness check (24h re-sync)

### Updated Files
- ✅ `internal/adapters/db/track_repository.go` - Added `BatchUpsert()`, `GetAllPathsModTime()`, `LatestAddedAtForRoot()`, `DeleteTracksForRoot()`; migration 002 adds `added_at` to tracks
- ✅ `app.go` - Added `ScanLibrary()` (uses PrepareLibrarySync), `OpenFolderDialog()`, `ListTracks()`
- ✅ `main.go` - Wired scanner service
- ✅ `frontend/src/App.svelte` - Real `handleBrowse()`, drag-and-drop folder, scan + load tracks, current path display
- ✅ `frontend/src/lib/types.ts` - `Track` interface aligned with backend
- ✅ `go.mod` - Added `github.com/dhowden/tag`

## What Was Done

1. **Filesystem Adapter**: Created adapter for file operations with cross-device move support
2. **Hashing**: Implemented partial hash (first 1MB + last 1MB for large files) and full hash
3. **Scanner Service**: 
   - Worker pool pattern (`runtime.NumCPU() * 4` workers)
   - Smart caching (skips files if `ModTime` matches database)
   - Batch upsert (500 tracks per batch)
   - Supports `.mp3`, `.flac`, `.m4a`, `.ogg`, `.wav`
4. **Repository Methods**: `BatchUpsert()`, `GetAllPathsModTime()` for scanning; `LatestAddedAtForRoot()` and `DeleteTracksForRoot()` for 24h staleness re-sync.
5. **Staleness re-sync**: `ScanLibrary()` calls `PrepareLibrarySync(root, 24h)`. If the most recent `added_at` for that root is older than 24h (or no tracks), all tracks for that root are deleted and a full scan runs. Otherwise scan is skipped (log: "Library still fresh, skipping re-sync").
6. **Frontend Integration**:
   - Real folder dialog via `OpenFolderDialog()`
   - Drag-and-drop folder on empty state (same action as Browse)
   - Automatic scan when folder selected or dropped
   - Track list loads from DB after scan; current path and total count shown
   - Loading states and error handling

## Testing Instructions

### Step 1: Install Dependencies
```bash
cd /Users/paulinnzodoum/workspace/personal/harmoniz
go mod tidy
```

### Step 2: Build & Run
```bash
wails dev
```

### Step 3: Validation Checklist

Use this list to validate Phase 2 before approval. Check each item after testing.

#### 1. Build & Run
- [ ] `wails dev` starts without errors
- [ ] No Go compilation errors
- [ ] Application window opens successfully

#### 2. Folder Dialog
- [ ] Click "Browse" (top bar or empty state) opens native folder picker
- [ ] Selecting a folder returns the path and path displays in UI
- [ ] Canceling dialog does not crash the app

#### 3. Drag & Drop (optional but implemented)
- [ ] Dragging a folder onto the empty-state drop zone shows visual feedback
- [ ] Dropping the folder sets the library path and triggers scan (same as Browse)

#### 4. Scanning
- [ ] `ScanLibrary()` is callable from frontend (runs when folder selected or dropped)
- [ ] Scan starts and completes without errors (check terminal logs)
- [ ] Loading/status visible in UI (e.g. "Scanning library..." or similar)
- [ ] After scan, track list loads and displays (first page)

#### 5. Database
- [ ] Tracks are inserted: `SELECT COUNT(*) FROM tracks WHERE is_deleted = 0;` returns > 0
- [ ] Sample row has expected columns, e.g.:
  ```sql
  SELECT path, title, artist_raw, size, hash_partial, mod_time FROM tracks LIMIT 1;
  ```
- [ ] `hash_partial` is populated for scanned files
- [ ] `mod_time` matches file modification time; `added_at` is set (migration 002)

#### 6. File Formats
- [ ] `.mp3` files are scanned
- [ ] `.flac` / `.m4a` (if available) are scanned
- [ ] `.ogg` / `.wav` are supported (scanner includes them)
- [ ] Non-audio files are ignored

#### 7. Performance
- [ ] Scan completes in reasonable time (e.g. &lt; ~1 s per 100 files on typical hardware)
- [ ] UI does not freeze during scan
- [ ] Memory usage remains stable

#### 8. Error Handling
- [ ] Corrupt or unreadable files are logged but do not stop the scan
- [ ] Invalid or missing paths are handled without crashing
- [ ] (Optional) Scan cancel is not required for Phase 2 sign-off

#### 9. Caching & Staleness
- [ ] Re-scanning the **same** folder with **no file changes** skips unchanged files (e.g. second scan faster; no "Cache hit" log—scanner skips silently)
- [ ] If you re-scan after 24h (or clear DB), a full re-sync runs (tracks for that root deleted then re-scanned); otherwise log may show "Library still fresh, skipping re-sync"
- [ ] Modifying a file on disk and re-scanning causes that file to be re-processed

#### 10. Frontend Display
- [ ] Current library path is visible (e.g. in top bar)
- [ ] Track list appears after scan (titles, artist, album, size)
- [ ] StatusBar or UI shows total track count for current library
- [ ] Empty state when no tracks (e.g. "No tracks found" / try scanning)
- [ ] Selecting a track updates the context panel (if implemented)

## Expected Terminal Output

When scanning, you should see logs similar to:
```json
{"level":"INFO","msg":"ScanLibrary called","root":"/path/to/music"}
{"level":"INFO","msg":"Library stale or empty, re-syncing from disk","root":"/path/to/music"}
{"level":"INFO","msg":"Starting scan","root":"/path/to/music"}
{"level":"INFO","msg":"Cache loaded","entries":0}
{"level":"INFO","msg":"Spawning workers","count":16}
{"level":"INFO","msg":"Scan completed","duration":"2.5s"}
```

If you scan the same folder again within 24h and no files changed:
```json
{"level":"INFO","msg":"Library still fresh, skipping re-sync","root":"/path/to/music"}
```

## Next Steps

After all validation criteria above are checked and you approve Phase 2, proceed to **Phase 3: Display Track List** — add **infinite scroll** and a **virtual list** (e.g. `@tanstack/svelte-virtual`) so the current path’s tracks scale to large libraries without jank. See `.cursor/plans/progressive_backend_integration_plan_eb6d9226.plan.md` for Phase 3 scope and validation.

## Notes

- Scanner uses a worker pool (`runtime.NumCPU() * 4`).
- Caching: files are skipped when `mod_time` matches DB; no "Cache hit" log.
- Batch inserts: 500 tracks per transaction.
- Staleness: if latest `added_at` for the root is &gt; 24h (or no tracks), all tracks for that root are deleted and a full scan runs.
- Frontend: browse and drag-and-drop both set path and trigger scan; tracks load from `ListTracks()` after scan.
