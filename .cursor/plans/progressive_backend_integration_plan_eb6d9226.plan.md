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
    content: "Phase 3: Display track list - Implement ListTracks in repository and UI, update frontend to use real data"
    status: pending
  - id: phase3-test
    content: "Phase 3 Test: Run through all validation criteria - backend method, frontend integration, data display, performance"
    status: pending
  - id: phase3-approval
    content: "Phase 3 Approval: Request user approval before proceeding to Phase 4"
    status: pending
isProject: false
---

# Progressive Backend Integration Plan

## Current State

- **Frontend**: Clean Svelte 5 + TailwindCSS implementation with UI components
- **Backend**: Minimal Wails setup with `main.go` and `app.go` (only has `Greet` method)
- **Backup**: Full implementation in `context/harmoniz-backup/` with hexagonal architecture

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

1. **Build & Run Test:**

- `wails dev` starts without errors
- No Go compilation errors
- Application window opens successfully

1. **Database Creation:**

- Database file created at `~/.harmoniz/library.db`
- File exists and is readable
- File size > 0 bytes

1. **Migration Execution:**

- Migration runs without errors
- `schema_migrations` table exists
- Version 1 recorded in `schema_migrations` table
- `tracks` table exists with correct schema

1. **Schema Verification** (using SQLite CLI or browser):

```sql
 -- Run these queries to verify:
 SELECT name FROM sqlite_master WHERE type='table' AND name='tracks';
 PRAGMA table_info(tracks);  -- Should show all columns
 SELECT version FROM schema_migrations;  -- Should return 1
```

1. **Code Quality:**

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

1. **Build & Run Test:**

- `wails dev` starts without errors
- No Go compilation errors
- Application window opens successfully

1. **Folder Dialog:**

- Click "Browse" button opens native folder picker
- Selecting a folder returns the path
- Path displays in UI after selection
- Canceling dialog doesn't crash app

1. **Scanning Functionality:**

- `ScanLibrary()` method exists and is callable from frontend
- Scan starts when folder is selected (or button clicked)
- Scan completes without errors (check terminal logs)
- Progress/status visible in UI (if implemented)

1. **Database Verification:**

- Tracks inserted into database
- Query database: `SELECT COUNT(*) FROM tracks WHERE is_deleted = 0;` returns > 0
- Sample track has correct metadata:
  ```sql
  SELECT path, title, artist_raw, size FROM tracks LIMIT 1;
  ```
- `hash_partial` populated for scanned files
- `mod_time` matches file system modification time

1. **File Format Support:**

- `.mp3` files scanned correctly
- `.flac` files scanned correctly (if available)
- `.m4a` files scanned correctly (if available)
- Non-audio files ignored

1. **Performance:**

- Scan completes in reasonable time (< 1 second per 100 files)
- No UI freezing during scan
- Memory usage stable

1. **Error Handling:**

- Corrupt files logged but don't crash scan
- Invalid paths handled gracefully
- Scan can be canceled (if implemented)

1. **Caching:**

- Re-scanning same folder skips unchanged files (check logs)
- Modified files re-processed

**APPROVAL REQUIRED**: After Phase 2 completion, verify all validation criteria and explicitly approve before proceeding to Phase 3.

---

### Phase 3: Display Track List

**Goal**: Display scanned tracks in the UI with pagination

**Files to create/copy:**

- `internal/adapters/db/track_repository.go` - Implement `ListTracks()` method
- `internal/core/domain/track.go` - Add `ListTracksResult` struct

**Changes to existing:**

- `internal/adapters/ui/app.go` - Add `ListTracks()` method
- `frontend/src/App.svelte` - Replace mock tracks with backend `ListTracks()` call
- Add pagination/infinite scroll if needed

**Test**:

- Track list displays real data from database
- Pagination works (if implemented)
- Track details show correctly

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
- **Caching**: Scanner skips files if `ModTime` matches database
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
