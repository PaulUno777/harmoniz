---
name: Progressive Backend Integration Plan
overview: Reintroduce backend features from the backup progressively, starting with database setup and folder scanning, then displaying tracks, followed by analysis features. Each phase will be tested before proceeding.
todos:
  - id: phase1-db-setup
    content: "Phase 1: Set up database infrastructure - Copy db adapter, consolidated migration (001_init.sql), logger, domain models, and update main.go"
    status: completed
  - id: phase1-test
    content: "Phase 1 Test: Run through all validation criteria - database creation, migration execution, schema verification"
    status: completed
  - id: phase1-approval
    content: "Phase 1 Approval: Request user approval before proceeding to Phase 2"
    status: pending
  - id: phase2-scanner
    content: "Phase 2: Implement folder browsing and scanning - Copy filesystem adapter, scanner service, and wire to UI"
    status: completed
  - id: phase2-test
    content: "Phase 2 Test: Run through all validation criteria - folder dialog, scanning, database verification, performance"
    status: pending
  - id: phase2-approval
    content: "Phase 2 Approval: Request user approval before proceeding to Phase 3"
    status: pending
  - id: phase3-list-tracks
    content: "Phase 3: Display track list - ListTracks + basic UI done; add infinite scroll + virtual list for optimization"
    status: completed
  - id: phase3-test
    content: "Phase 3 Test: Run through all validation criteria - virtual list, infinite scroll, current path display, performance"
    status: completed
  - id: phase3-approval
    content: "Phase 3 Approval: Request user approval before proceeding to Phase 4"
    status: pending
isProject: false
---

# Progressive Backend Integration Plan

## Current State

- **Frontend**: Svelte 5 + TailwindCSS; library tab shows current path, browse/drag-drop, and a simple track grid (no virtual list or infinite scroll yet).
- **Backend**: Wails app with DB, scanner, `ListTracks(root, limit, offset)`, `ScanLibrary` (with 24h staleness re-sync: if `added_at` is older than 24h, tracks for that root are deleted and re-scanned).
- **Backup**: Full implementation in `context/harmoniz-backup/` with hexagonal architecture; reference for VirtualTable (`@tanstack/svelte-virtual`) and infinite-scroll approach.

## Architecture Overview

The backup follows **Hexagonal Architecture**:

- **Domain**: Pure structs (`Track`, `Transaction`, `ArtistSuggestion`)
- **Ports**: Interfaces (`TrackRepository`, `FileSystem`)
- **Services**: Business logic (`Scanner`, `Rename`, `Cleaner`, `Analysis`)
- **Adapters**: Implementations (`db/sqlite`, `fs`, `ui/app.go`)

## Integration Phases

### Phase 1: Database & Infrastructure Setup

**Goal**: Set up SQLite database with consolidated migrations and basic infrastructure

**Files to create/copy:**

- `internal/adapters/db/sqlite.go` - SQLite adapter with WAL mode
- `internal/adapters/db/track_repository.go` - Repository implementation (minimal for Phase 1)
- `internal/adapters/db/migrations/001_init.sql` - **CONSOLIDATED** migration (tracks table + indexes only)
- `internal/logger/logger.go` - Structured logging
- `internal/core/domain/track.go` - Track domain model
- `internal/core/ports/repository.go` - Repository interface

**Changes to existing:**

- `main.go` - Initialize database, run migrations
- `go.mod` - Add dependencies: `modernc.org/sqlite`

**Consolidated Migration Strategy:**

- **001_init.sql**: Contains only `tracks` table with all necessary columns and indexes
- **Excludes**: `operations`, `transactions` tables (added in later phases when needed)
- **Includes**: All track columns including `deleted_at` and `delete_reason` (for future use)

**Validation Criteria:**

1. **Build & Run**

- `wails dev` starts without errors; no Go compilation errors; application window opens.

1. **Database Creation**

- Database file created at `~/.harmoniz/library.db`
- File exists and is readable
- File size > 0 bytes

1. **Migration Execution**

- Migration runs without errors
- `schema_migrations` table exists
- Version 1 recorded in `schema_migrations` table
- `tracks` table exists with correct schema

1. **Schema Verification** (SQLite CLI or browser)

```sql
 -- Run these queries to verify:
 SELECT name FROM sqlite_master WHERE type='table' AND name='tracks';
 PRAGMA table_info(tracks);  -- Should show all columns
 SELECT version FROM schema_migrations;  -- Should return 1
```

1. **Code Quality**

- No console errors in terminal
- Logger outputs structured JSON logs
- Database connection closes properly on app exit

**APPROVAL REQUIRED**: After Phase 1 completion, you must verify all validation criteria and explicitly approve before proceeding to Phase 2.

---

### Phase 2: Folder Browsing & Scanning

**Goal**: Browse audio folder and scan files into database

**Files to create/copy:**

- `internal/adapters/fs/fs.go` - Filesystem adapter
- `internal/adapters/fs/hash.go` - Hashing utilities
- `internal/core/services/scanner/scanner.go` - Scanner service
- `internal/core/ports/filesystem.go` - FileSystem interface

**Changes to existing:**

- `internal/adapters/ui/app.go` - Add `ScanLibrary()` and `OpenFolderDialog()` methods
- `internal/adapters/db/track_repository.go` - Add `BatchUpsert()` and `GetAllPathsModTime()` methods
- `main.go` - Wire scanner service
- `frontend/src/App.svelte` - Replace mock `handleBrowse()` with real backend call
- `frontend/src/lib/types.ts` - Update `Track` interface to match backend model
- `go.mod` - Add dependency: `github.com/dhowden/tag`

**Validation Criteria:**

1. **Build & Run**

- `wails dev` starts without errors; no Go compilation errors; application window opens.

1. **Folder Dialog**

- Click "Browse" button opens native folder picker
- Selecting a folder returns the path
- Path displays in UI after selection
- Canceling dialog doesn't crash app.

1. **Scanning**

- `ScanLibrary()` method exists and is callable from frontend
- Scan starts when folder is selected (or button clicked)
- Scan completes without errors (check terminal logs)
- Progress/status visible in UI (if implemented).

1. **Database**

- Tracks inserted into database
- Query database: `SELECT COUNT(*) FROM tracks WHERE is_deleted = 0;` returns > 0
- Sample track has correct metadata:

```sql
  SELECT path, title, artist_raw, size FROM tracks LIMIT 1;


```

- `hash_partial` populated for scanned files
- `mod_time` matches file system modification time

1. **File formats**

- `.mp3` files scanned correctly
- `.flac` files scanned correctly (if available)
- `.m4a` files scanned correctly (if available)
- Non-audio files ignored.

1. **Performance**

- Scan completes in reasonable time (< 1 second per 100 files)
- No UI freezing during scan
- Memory usage stable.

1. **Error handling**

- Corrupt files logged but don't crash scan
- Invalid paths handled gracefully
- Scan can be canceled (if implemented)

1. **Caching**

- Re-scanning same folder skips unchanged files (check logs)
- Modified files re-processed

**APPROVAL REQUIRED**: After Phase 2 completion, verify all validation criteria and explicitly approve before proceeding to Phase 3.

---

### Phase 3: Display Track List (current path) with Infinite Scroll + Virtual List

**Goal**: Display the current browsed path’s tracked audio in the UI with **infinite scroll** (load more as user scrolls) and a **virtual list** (only render visible rows) for performance. Align with the approach used in `context/harmoniz-backup` (VirtualTable + `@tanstack/svelte-virtual`).

**Done so far:**

- Backend: `ListTracks(root, limit, offset)` in repository and `app.go`
- Frontend: Loads first page (e.g. 100 tracks), shows current path, displays tracks in a simple grid
- StatusBar shows total track count for current library

**Remaining work:**

1. **Virtual list**

- Add dependency: `@tanstack/svelte-virtual` (same as backup).
- Introduce a virtualized track list component that renders only visible rows (fixed row height, e.g. ~36–44px).
- Use a single scrollable container; total height = `rowCount * rowHeight`; position rows with `transform: translateY()`.

1. **Infinite scroll**

- Keep a single “tracks” array that accumulates pages.
- When the user scrolls near the bottom (e.g. within 1–2 screens), call `ListTracks(currentLibraryPath, pageSize, offset)` and append the next page.
- Use `total` from the first/current response to know when to stop requesting.
- Avoid duplicate requests (e.g. guard with “loading more” or “has no more” flags).

1. **UI**

- Current path stays visible (already in top bar).
- Track list shows: title, artist, album, size (and optionally filename, bitrate).
- Row click selects track and opens context panel (already in place).
- Optional: sticky header row (e.g. #, Title, Artist, Album, Size) for table-style layout.

**Files to create:**

- `frontend/src/lib/components/library/VirtualTrackList.svelte` (or similar) – virtual list + infinite scroll, consumes `tracks`, `total`, `loading`, `onLoadMore(offset)`.

**Changes to existing:**

- `frontend/package.json` – add `@tanstack/svelte-virtual`.
- `frontend/src/App.svelte` – use virtual list component; replace single `loadTracks()` with initial load + “load more” when scrolling (infinite scroll); pass current path and track data.

**Reference (backup):**

- `context/harmoniz-backup/frontend/src/lib/components/organizer/VirtualTable.svelte` – uses `createVirtualizer` from `@tanstack/svelte-virtual`, fixed row height 36px, overscan 10.
- Backup loads one big page (e.g. 10000) into store; we will use incremental loading instead for true infinite scroll.

**Phase 3 Validation Criteria (verify before approval):**

1. **Current path display**

- Browsed folder path is visible in the UI (e.g. top bar or content area).
- After scan or drop, the path reflects the current library root.

1. **Track list shows current path’s audio only**

- List shows only tracks under the current library path (no other roots).
- Total count in StatusBar (or UI) matches backend `ListTracks(..., total)` for that root.

1. **Virtual list**

- A virtual list is used (e.g. `@tanstack/svelte-virtual`): only a window of rows is rendered (DOM node count much smaller than total track count).
- With 500+ tracks, only on the order of tens of rows in DOM (e.g. visible + overscan).
- Scrolling through a large list is smooth (no obvious jank).

1. **Infinite scroll**

- Initial load fetches first page (e.g. limit 100–200).
- Scrolling near the bottom loads the next page and appends to the list.
- When all pages are loaded (offset + limit ≥ total), no further requests are made.
- No duplicate requests for the same range (guarded by loading / hasMore state).

1. **Data and interaction**

- Each row shows at least: title, artist, album, size (or equivalent).
- Clicking a row selects the track and updates the context panel (if implemented).
- Empty state: when there are no tracks, a clear message is shown.

1. **Build and stability**

- `wails dev` runs without errors; no console errors in normal flow.
- Changing library (browse/drop) clears or replaces the list and shows the new library’s tracks.

**APPROVAL REQUIRED**: After Phase 3 completion, verify all criteria above and explicitly approve before proceeding to Phase 4.

---

### Phase 4: Analysis Features (Future)

**Goal**: Add artist clustering and duplicate detection

**Files to create/copy:**

- `internal/core/services/analysis/clustering.go`
- `internal/core/services/analysis/deduplication.go`
- `internal/core/domain/suggestion.go`
- `internal/core/jobs/manager.go` - Job management for long-running tasks

**Changes to existing:**

- `internal/adapters/ui/app.go` - Add analysis methods
- Frontend components for analysis results

---

### Phase 5+: Additional Features (Future)

- Rename service
- Cleaner service
- Transaction system
- Undo/Redo functionality

## Implementation Strategy

1. **Copy files selectively** - Only copy what's needed for current phase
2. **Adapt imports** - Update module paths from `harmonizr` to `harmoniz`
3. **Test incrementally** - After each phase, test thoroughly before proceeding
4. **Keep main.go simple** - Use it as entry point, wire dependencies there
5. **Maintain frontend compatibility** - Ensure TypeScript types match Go structs

## Key Dependencies to Add

```go
require (
    github.com/dhowden/tag v0.0.0-20240417053706-3d75831295e8  // Audio metadata
    modernc.org/sqlite v1.45.0                                 // Pure Go SQLite
    github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342  // String metrics (for fuzzy matching)
    golang.org/x/text v0.34.0                                  // Text normalization
)
```

## Notes

- **Database location**: `~/.harmoniz/library.db` (configurable)
- **Concurrency**: Scanner uses worker pool pattern (`runtime.NumCPU() * 4`)
- **Caching**: Scanner skips files if `ModTime` matches database. **Staleness re-sync**: `tracks.added_at` records when a track was first added; if the latest `added_at` for a root is older than 24h, `ScanLibrary` deletes all tracks for that root and re-scans (rsync-style).
- **Error handling**: Scanner logs errors but continues (never crashes)
- **File formats**: Supports `.mp3`, `.flac`, `.m4a`, `.ogg`, `.wav`

## Testing Checklist (Per Phase)

- `wails dev` runs without errors
- Database file created successfully
- Migrations run without errors
- Backend methods callable from frontend
- UI updates correctly with real data
- No console errors
- Performance acceptable (< 1s for typical operations)
