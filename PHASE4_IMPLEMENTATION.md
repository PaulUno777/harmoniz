# Phase 4: Analysis Features – Implementation Summary

## Done

### 1. Domain & config
- **`internal/core/domain/suggestion.go`** – `ArtistSuggestion` (original, suggested, score, reason, confidence_level).
- **`internal/core/services/analysis/config.go`** – `ClusteringConfig` and `DedupConfig` with defaults and optional overrides.

### 2. Repository (root-scoped)
- **`internal/core/ports/repository.go`** – `GetDuplicateCandidates(ctx, root)` and `StreamUniqueArtists(ctx, root)`; empty `root` = whole DB.
- **`internal/adapters/db/track_repository.go`**:
  - `GetDuplicateCandidates(ctx, root)`: JOIN on sizes with `COUNT(*) > 1`, same path filter for subquery and outer query; returns full track rows via shared `scanTracks(rows)`.
  - `StreamUniqueArtists(ctx, root)`: `SELECT DISTINCT artist_norm` with path filter; streams on a channel with context cancellation.
  - **`scanTracks(rows)`** – shared helper for duplicate-candidates and consistent null handling.

### 3. Clustering service
- **`internal/core/services/analysis/clustering.go`**:
  - `NewClusteringService(repo)` / `NewClusteringServiceWithConfig(repo, cfg)`.
  - `AnalyzeArtists(ctx, root)`:
    - Streams artists, buckets by first rune (non-letter/non-number → `?`).
    - For each bucket: pairwise Jaro-Winkler with length filter; threshold and max bucket size from config.
    - Returns `[]domain.ArtistSuggestion` with confidence from score (High/Medium/Low).
  - Context checked in the read loop and before heavy work.

### 4. Deduplication service
- **`internal/core/services/analysis/deduplication.go`**:
  - `NewDeduplicationService(repo)` / `NewDeduplicationServiceWithConfig(repo, cfg)`.
  - `DetectDuplicates(ctx, root)`:
    - Stage 1: `GetDuplicateCandidates(ctx, root)` (same size).
    - Stage 2: group by partial hash in memory; skip empty partial hash.
    - Stage 3: `groupByFullHash(ctx, group)` – compute full hash only when missing, then group by full hash; duplicate groups = len > 1.
  - Context checked in the full-hash loop.
  - Config: `MaxSizeGroupLog` to warn on large size groups.

### 5. App wiring
- **`app.go`** – App holds `clustering` and `deduper`; `NewApp(scanner, repo)` builds them; `AnalyzeArtists(root)` and `DetectDuplicates(root)` call services with `a.ctx` (fallback to `context.Background()` if nil).
- **`main.go`** – `NewApp(scannerService, dbAdapter)`.

### 6. Dependencies
- **`go.mod`** – `require github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673`.

### 7. Tests
- **`internal/core/services/analysis/clustering_test.go`** – `TestClusteringConfigDefaults`, `TestDedupConfigDefaults`.

## One-time setup

From the repo root, with network:

```bash
go mod tidy
```

Then build and run:

```bash
go build ./...
wails dev
```

After `wails dev`, bindings in `frontend/wailsjs/go/main/App.js` (and types) will include:
- `AnalyzeArtists(root: string) => Promise<ArtistSuggestion[]>`
- `DetectDuplicates(root: string) => Promise<Track[][]>`

Use the current library path as `root` when calling from the UI so analysis is scoped to the current library.

## Design notes

- **Root scoping**: All analysis is scoped by library root when `root` is set; empty string = entire DB (e.g. for “Analyze all”).
- **Config**: Clustering (threshold, max length diff, max bucket size, Jaro-Winkler params) and dedup (max size group to log) are configurable for tuning and tests.
- **Context**: Streaming and full-hash loops respect `ctx.Done()` for cancellation.
- **Dedup**: No mutation of slice elements when computing full hash; we assign `t.HashFull` on the slice element (in-place) only for grouping; no DB write of full hash in this phase.
- **Errors**: Wrapped with `fmt.Errorf(..., %w)`; logging at appropriate levels (Warn for large buckets/groups, Error for hash failures).
