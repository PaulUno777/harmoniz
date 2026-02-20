# Phase 4: Analysis Features - Detailed Plan

## Purpose & Goal

**Purpose**: Add intelligent analysis capabilities to help users identify and resolve data quality issues in their music library:
1. **Artist Clustering**: Detect similar artist names (e.g., "Guns N' Roses" vs "guns n roses") and suggest normalization
2. **Duplicate Detection**: Find exact duplicate files using a multi-stage funnel approach (size → partial hash → full hash)

**Goal**: Enable users to discover inconsistencies and duplicates in their library, providing actionable suggestions for cleanup.

## Architecture Overview

Following **Hexagonal Architecture** pattern (same as backup):

- **Domain**: Pure structs (`ArtistSuggestion`)
- **Ports**: Repository interface methods (`GetDuplicateCandidates`, `StreamUniqueArtists`)
- **Services**: Business logic (`ClusteringService`, `DeduplicationService`)
- **Adapters**: UI adapter (`app.go`) exposes methods to frontend

## Feature Details

### 1. Artist Clustering

**Problem**: Music libraries often have inconsistent artist names:
- "Guns N' Roses" vs "guns n roses" vs "Guns and Roses"
- "AC/DC" vs "ACDC"
- Case variations, punctuation differences, etc.

**Solution**: Use fuzzy string matching (Jaro-Winkler algorithm) to detect similar artist names and suggest a canonical form.

**Algorithm**:
1. **Bucketing**: Group artists by first character (normalized) to reduce comparison space
2. **Pairwise Comparison**: Within each bucket, compare all pairs using Jaro-Winkler distance
3. **Threshold**: Score > 0.9 suggests similarity
4. **Length Filter**: Skip comparisons if length difference > 3 characters

**Output**: Array of `ArtistSuggestion` objects:
```go
type ArtistSuggestion struct {
    Original        string  // e.g., "guns n roses"
    Suggested      string  // e.g., "Guns N' Roses"
    Score           float64 // Jaro-Winkler score (0.0-1.0)
    Reason          string  // e.g., "Similar Name"
    ConfidenceLevel string  // "High", "Medium", "Low"
}
```

### 2. Duplicate Detection

**Problem**: Users may have multiple copies of the same audio file (different paths, same content).

**Solution**: Multi-stage funnel approach for performance:
1. **Stage 1 (SQL)**: Group by file size (fast, filters out most non-duplicates)
2. **Stage 2 (RAM)**: Group by partial hash (first 1MB + last 1MB) - already computed during scan
3. **Stage 3 (Disk I/O)**: Compute full hash only for candidates that match size + partial hash

**Why This Works**:
- Most files have unique sizes → Stage 1 eliminates ~99% of comparisons
- Partial hash collision is rare → Stage 2 eliminates most remaining false positives
- Full hash is expensive (disk I/O) → Only computed when necessary

**Output**: Array of duplicate groups (each group is an array of `Track` objects):
```go
[[]domain.Track] // Each inner array represents a group of duplicate files
```

**Example**:
```go
[
    [Track{Path: "/music/song1.mp3"}, Track{Path: "/music/copy/song1.mp3"}], // Group 1: 2 duplicates
    [Track{Path: "/music/track2.mp3"}, Track{Path: "/backup/track2.mp3"}, Track{Path: "/old/track2.mp3"}] // Group 2: 3 duplicates
]
```

## Implementation Plan

### Step 1: Add Dependencies

**File**: `go.mod`

Add required packages:
```go
require (
    github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342  // Jaro-Winkler string matching
    golang.org/x/text v0.34.0                                  // Text normalization (already present)
)
```

**Note**: `golang.org/x/text` is already in dependencies, but we may need to update version.

### Step 2: Create Domain Model

**File**: `internal/core/domain/suggestion.go`

```go
package domain

type ArtistSuggestion struct {
    Original        string  `json:"original"`
    Suggested       string  `json:"suggested"`
    Score           float64 `json:"score"`            // Jaro-Winkler score (0.0 - 1.0)
    Reason          string  `json:"reason"`           // e.g., "Similar Name"
    ConfidenceLevel string  `json:"confidence_level"` // "High", "Medium", "Low"
}
```

### Step 3: Add Repository Methods

**File**: `internal/core/ports/repository.go`

Methods already exist in interface:
- `GetDuplicateCandidates() ([]domain.Track, error)` ✅
- `StreamUniqueArtists(ctx context.Context) (<-chan string, error)` ✅

**File**: `internal/adapters/db/track_repository.go`

**Implement**:
1. `GetDuplicateCandidates()` - SQL query to find tracks with same size
2. `StreamUniqueArtists(ctx)` - Stream unique normalized artist names

**SQL for duplicates**:
```sql
SELECT t1.* FROM tracks t1
JOIN (SELECT size FROM tracks WHERE is_deleted = 0 GROUP BY size HAVING COUNT(*) > 1) t2
ON t1.size = t2.size
WHERE t1.is_deleted = 0
ORDER BY t1.size DESC
```

**SQL for artists**:
```sql
SELECT DISTINCT artist_norm FROM tracks 
WHERE is_deleted = 0 AND artist_norm != ''
```

### Step 4: Create Clustering Service

**File**: `internal/core/services/analysis/clustering.go`

**Key Functions**:
- `NewClusteringService(repo ports.TrackRepository) *ClusteringService`
- `AnalyzeArtists(ctx context.Context) ([]domain.ArtistSuggestion, error)`

**Algorithm Flow**:
1. Stream unique artists from repository
2. Bucket by first character (normalized)
3. For each bucket with 2+ artists:
   - Compare all pairs (O(N²) but N is small per bucket)
   - Filter by length difference (max 3 chars)
   - Calculate Jaro-Winkler score
   - If score > 0.9, create suggestion
4. Return all suggestions

**Performance Notes**:
- Skip buckets with > 500 artists (log warning)
- Use goroutine-safe channel streaming
- Context cancellation support

### Step 5: Create Deduplication Service

**File**: `internal/core/services/analysis/deduplication.go`

**Key Functions**:
- `NewDeduplicationService(repo ports.TrackRepository) *DeduplicationService`
- `DetectDuplicates() ([][]domain.Track, error)`

**Algorithm Flow**:
1. Get duplicate candidates (same size) from repository
2. Group by size in memory
3. For each size group:
   - Sub-group by partial hash
   - For groups with 2+ tracks:
     - Compute full hash for tracks missing it
     - Group by full hash
     - If group has 2+ tracks → duplicate group found
4. Return all duplicate groups

**Performance Notes**:
- Warn if size group > 500 tracks
- Compute full hash only when needed (lazy)
- Use existing `fs.ComputeFullHash()` function

### Step 6: Wire Services in App

**File**: `app.go`

**Changes**:
1. Add service fields to `App` struct:
   ```go
   type App struct {
       ctx        context.Context
       scanner    *scanner.Service
       clustering *analysis.ClusteringService
       deduper    *analysis.DeduplicationService
   }
   ```

2. Update `NewApp()` to initialize services:
   ```go
   func NewApp(scannerService *scanner.Service, repo ports.TrackRepository) *App {
       return &App{
           scanner:    scannerService,
           clustering: analysis.NewClusteringService(repo),
           deduper:    analysis.NewDeduplicationService(repo),
       }
   }
   ```

3. Add public methods:
   ```go
   // AnalyzeArtists analyzes artist names and returns clustering suggestions
   func (a *App) AnalyzeArtists() ([]domain.ArtistSuggestion, error) {
       return a.clustering.AnalyzeArtists(a.ctx)
   }

   // DetectDuplicates finds duplicate files using multi-stage funnel
   func (a *App) DetectDuplicates() ([][]domain.Track, error) {
       return a.deduper.DetectDuplicates()
   }
   ```

**File**: `main.go`

**Changes**:
- Update `NewApp()` call to pass repository:
  ```go
  app := NewApp(scannerService, dbAdapter)
  ```

### Step 7: Frontend Integration (Optional for Phase 4)

**Note**: Phase 4 focuses on backend implementation. Frontend UI can be added in a follow-up phase.

**Future Frontend Work**:
- Add "Analysis" tab or section in UI
- Display artist suggestions in a list/table
- Display duplicate groups with comparison view
- Allow user to accept/reject suggestions
- Show progress indicators for long-running analysis

## Files to Create

1. ✅ `internal/core/domain/suggestion.go` - ArtistSuggestion domain model
2. ✅ `internal/core/services/analysis/clustering.go` - Clustering service
3. ✅ `internal/core/services/analysis/deduplication.go` - Deduplication service

## Files to Modify

1. ✅ `go.mod` - Add `github.com/xrash/smetrics` dependency
2. ✅ `internal/adapters/db/track_repository.go` - Implement `GetDuplicateCandidates()` and `StreamUniqueArtists()`
3. ✅ `app.go` - Wire analysis services and expose methods
4. ✅ `main.go` - Pass repository to `NewApp()`

## Validation Criteria

### 1. Build & Run
- [ ] `go mod tidy` runs without errors
- [ ] `wails dev` starts without errors
- [ ] No Go compilation errors

### 2. Repository Methods
- [ ] `GetDuplicateCandidates()` returns tracks with same size
- [ ] `StreamUniqueArtists()` streams unique normalized artists
- [ ] Methods handle empty results gracefully
- [ ] Context cancellation works for streaming

### 3. Clustering Service
- [ ] `AnalyzeArtists()` returns suggestions for similar artist names
- [ ] Jaro-Winkler scores are between 0.0 and 1.0
- [ ] Suggestions have reasonable confidence levels
- [ ] Large buckets (>500) are skipped with warning
- [ ] Performance acceptable (< 5 seconds for 10k artists)

### 4. Deduplication Service
- [ ] `DetectDuplicates()` returns groups of duplicate tracks
- [ ] Only exact duplicates (same full hash) are grouped
- [ ] Full hash computed only when needed
- [ ] Performance acceptable (< 10 seconds for 10k tracks)
- [ ] Handles missing `hash_full` gracefully

### 5. App Integration
- [ ] `AnalyzeArtists()` method callable from frontend
- [ ] `DetectDuplicates()` method callable from frontend
- [ ] Methods return correct JSON-serializable types
- [ ] Error handling works correctly

### 6. Testing
- [ ] Test with library containing duplicate artists
- [ ] Test with library containing duplicate files
- [ ] Test with empty library (no errors)
- [ ] Test with large library (performance acceptable)

## Performance Considerations

### Clustering
- **Bucketing**: Reduces O(N²) comparisons by grouping by first character
- **Length Filter**: Skips comparisons where length difference > 3
- **Large Bucket Skip**: Prevents O(N²) explosion for common first letters (e.g., "The")
- **Streaming**: Uses channels to avoid loading all artists into memory

### Deduplication
- **Multi-Stage Funnel**: 
  - Stage 1 (SQL): Fast size-based filtering
  - Stage 2 (RAM): Partial hash grouping (already computed)
  - Stage 3 (Disk): Full hash only for candidates
- **Lazy Full Hash**: Computed only when partial hash matches
- **Batch Processing**: Processes size groups sequentially

## Notes from Backup Implementation

1. **Normalization**: Artist names are already normalized in `artist_norm` column during scan
2. **Hash Strategy**: Partial hash (first 1MB + last 1MB) is computed during scan; full hash computed on-demand
3. **Error Handling**: Services log errors but don't crash; return partial results if possible
4. **Context Support**: Both services support context cancellation for long-running operations
5. **Memory Efficiency**: Streaming used for artist clustering to avoid loading all artists into memory

## Dependencies Reference

From backup `go.mod`:
```go
require (
    github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342  // Jaro-Winkler
    golang.org/x/text v0.34.0                                  // Text normalization
)
```

## Next Steps After Phase 4

- **Phase 5**: Frontend UI for analysis results
- **Phase 6**: Apply suggestions (rename artists, delete duplicates)
- **Phase 7**: Transaction system for undo/redo
