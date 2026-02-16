# Phase 2: Folder Browsing & Scanning - COMPLETE

## Files Created

### Filesystem Adapter
- ✅ `internal/adapters/fs/fs.go` - Filesystem adapter with SafeMove
- ✅ `internal/adapters/fs/hash.go` - Hashing utilities (partial and full hash)
- ✅ `internal/core/ports/filesystem.go` - Filesystem interface

### Scanner Service
- ✅ `internal/core/services/scanner/scanner.go` - Scanner service with worker pool

### Updated Files
- ✅ `internal/adapters/db/track_repository.go` - Added `BatchUpsert()` and `GetAllPathsModTime()` methods, plus stub implementations for other interface methods
- ✅ `app.go` - Added `ScanLibrary()`, `OpenFolderDialog()`, and `ListTracks()` methods
- ✅ `main.go` - Wired scanner service
- ✅ `frontend/src/App.svelte` - Replaced mock `handleBrowse()` with real backend calls, added scan and load tracks functionality
- ✅ `frontend/src/lib/types.ts` - Updated `Track` interface to match backend model
- ✅ `go.mod` - Added dependency: `github.com/dhowden/tag`

## What Was Done

1. **Filesystem Adapter**: Created adapter for file operations with cross-device move support
2. **Hashing**: Implemented partial hash (first 1MB + last 1MB for large files) and full hash
3. **Scanner Service**: 
   - Worker pool pattern (`runtime.NumCPU() * 4` workers)
   - Smart caching (skips files if `ModTime` matches database)
   - Batch upsert (500 tracks per batch)
   - Supports `.mp3`, `.flac`, `.m4a`, `.ogg`, `.wav`
4. **Repository Methods**: Implemented `BatchUpsert()` and `GetAllPathsModTime()` for scanning
5. **Frontend Integration**: 
   - Real folder dialog via `OpenFolderDialog()`
   - Automatic scanning when folder selected
   - Track list loads from database after scan
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

#### ✅ Build & Run Test
- [ ] `wails dev` starts without errors
- [ ] No Go compilation errors
- [ ] Application window opens successfully

#### ✅ Folder Dialog
- [ ] Click "Browse" button opens native folder picker
- [ ] Selecting a folder returns the path
- [ ] Path displays in UI after selection
- [ ] Canceling dialog doesn't crash app

#### ✅ Scanning Functionality
- [ ] `ScanLibrary()` method exists and is callable from frontend
- [ ] Scan starts when folder is selected (or button clicked)
- [ ] Scan completes without errors (check terminal logs)
- [ ] Progress/status visible in UI (loading indicator shows)

#### ✅ Database Verification
- [ ] Tracks inserted into database
- [ ] Query database: `SELECT COUNT(*) FROM tracks WHERE is_deleted = 0;` returns > 0
- [ ] Sample track has correct metadata:
  ```sql
  SELECT path, title, artist_raw, size FROM tracks LIMIT 1;
  ```
- [ ] `hash_partial` populated for scanned files
- [ ] `mod_time` matches file system modification time

#### ✅ File Format Support
- [ ] `.mp3` files scanned correctly
- [ ] `.flac` files scanned correctly (if available)
- [ ] `.m4a` files scanned correctly (if available)
- [ ] Non-audio files ignored

#### ✅ Performance
- [ ] Scan completes in reasonable time (< 1 second per 100 files)
- [ ] No UI freezing during scan
- [ ] Memory usage stable

#### ✅ Error Handling
- [ ] Corrupt files logged but don't crash scan (check terminal logs)
- [ ] Invalid paths handled gracefully
- [ ] Scan can be canceled (if implemented)

#### ✅ Caching
- [ ] Re-scanning same folder skips unchanged files (check logs for "Cache hit" messages)
- [ ] Modified files re-processed

#### ✅ Frontend Display
- [ ] Track list displays after scan completes
- [ ] Track titles display correctly
- [ ] Artist names display correctly
- [ ] Album names display correctly
- [ ] File sizes display correctly (formatted as MB/KB)
- [ ] Empty state shows when no tracks found

## Expected Terminal Output

When scanning, you should see logs like:
```json
{"level":"INFO","msg":"Starting scan","root":"/path/to/music"}
{"level":"INFO","msg":"Cache loaded","entries":0}
{"level":"INFO","msg":"Spawning workers","count":16}
{"level":"INFO","msg":"Scan completed","duration":"2.5s"}
```

## Next Steps

Once all validation criteria are met and you approve Phase 2, we will proceed to **Phase 3: Display Track List** (Note: ListTracks is already implemented, but we can enhance pagination and filtering if needed).

## Notes

- Scanner uses worker pool for concurrent processing
- Files are cached based on modification time - unchanged files are skipped
- Batch inserts improve performance (500 tracks per transaction)
- Frontend automatically loads tracks after scan completes
- Drag & drop also triggers scan automatically
