# Phase 4 UI Plan: Analysis Features (Artist Clustering + Duplicate Detection)

## Goal

Expose Phase 4 backend analysis (artist clustering and duplicate detection) in the app with clear UX: run analysis for the **current library**, show results in dedicated views, and reuse patterns from `context/harmoniz-backup` where appropriate.

## Feature separation (important)

- **Phase 4 (this plan)** = **Analysis only** (read-only): show artist suggestions and duplicate groups; **no** renaming, **no** tag writing, **no** file moves.
- **Organize (Phase 5)** = **Separate feature**: rename files, set artist/title/tags, organize library. Its UI and backend are distinct from analysis; do not mix “Organizer” analysis views with rename/tag/organize actions.
- **Cleaner (Phase 6)** = Resolve duplicates (delete), prune, empty folders — also separate from both Analysis and Organize.

## Current State

- **Tabs**: Library, Organizer, Cleaner, Settings. Only Library and Settings have real content; Organizer/Cleaner currently show the same library content when a path is set.
- **Backend**: `AnalyzeArtists(root)` and `DetectDuplicates(root)` are exposed from `app.go`; Wails bindings will be in `frontend/wailsjs/go/main/App.js` after `wails build` / `wails dev`.
- **Backup reference**: 
  - **Cleaner**: Duplicate groups in accordion cards (`DuplicateCluster.svelte`), “Run Analysis” when empty, BulkActions (clean folders, auto-keep).
  - **Analysis store**: `duplicateGroups`, `artistSuggestions`, `loading`, `setDuplicates()`, `setArtistSuggestions()`.
  - **App**: Single “Analyze” button in the top bar that runs both analyses with `Promise.all`, then updates the store and shows a toast.
  - **Artist suggestions**: Stored but not rendered in the backup; we will add a dedicated UI.

## UX Principles

1. **Scope by current library** – Analysis always uses `currentLibraryPath`; empty state when no library is selected.
2. **Single “Run analysis” action** – One action runs both artist clustering and duplicate detection for the current root; loading state and errors are clear.
3. **Dedicated views** – Organizer = artist suggestions; Cleaner = duplicate groups (no mixing).
4. **Clear empty/loading/error states** – Empty: “Run analysis” CTA; Loading: spinner + message; Error: inline or small banner.
5. **Scannable, minimal UI** – Cards/sections, key info (original → suggested, score; filename, size, path), no clutter.
6. **Future-proof** – Placeholders for “Apply” / “Resolve” actions later (e.g. accept suggestion, delete duplicate) without implementing them in this phase.

## Architecture

### 1. Analysis state

**Option A – Analysis store (recommended, matches backup)**  
- **File**: `frontend/src/lib/stores/analysis.ts` (or `.svelte.ts` if using runes).
- **State**:  
  - `duplicateGroups: Track[][]`  
  - `artistSuggestions: ArtistSuggestion[]`  
  - `loading: boolean`  
  - `error: string | null`  
- **Methods**: `setDuplicates(groups)`, `setArtistSuggestions(suggestions)`, `setLoading(bool)`, `setError(msg)`, optional `reset()`.

**Option B – State in App.svelte**  
- Keep `analysisDuplicates`, `analysisSuggestions`, `analysisLoading`, `analysisError` in App and pass as props to Organizer/Cleaner. Simpler but less reusable; store is better if we add more analysis later.

**Recommendation**: Use a small **analysis store** (Option A) so Organizer and Cleaner can subscribe independently and we can add toasts or other consumers later.

### 2. Types

- **File**: `frontend/src/lib/types.ts`
- **Add**:
  - `ArtistSuggestion`: `{ original: string; suggested: string; score: number; reason: string; confidence_level: string }`.
  - Reuse existing `Track` for duplicate groups; backend returns `Track[][]`.

### 3. App.svelte changes

- **Imports**: `AnalyzeArtists`, `DetectDuplicates` from `../wailsjs/go/main/App.js`.
- **Handler**: e.g. `async function handleRunAnalysis()`:
  - If `!currentLibraryPath`: show message or focus “Browse” (no backend call).
  - Set analysis store `loading = true`, `error = null`.
  - `Promise.all([AnalyzeArtists(currentLibraryPath), DetectDuplicates(currentLibraryPath)])`.
  - Map backend response to frontend types (e.g. `artist_raw` → `artist` if needed).
  - Call `analysis.setDuplicates(duplicates ?? [])`, `analysis.setArtistSuggestions(suggestions ?? [])`.
  - On catch: `analysis.setError(e.message)` (and optionally clear results).
  - In finally: `analysis.setLoading(false)`.
- **Routing**:
  - When `activeTab === 'organizer'`: render **OrganizerView** (see below).
  - When `activeTab === 'cleaner'`: render **CleanerView** (see below).
  - When `activeTab === 'library'`: keep current **LibraryContent**.
- **Top bar**: Add a “Run analysis” (or “Analyze”) button, e.g. next to Browse, disabled when `!currentLibraryPath` or `analysis.loading`. Optional: only show when `activeTab === 'organizer' || activeTab === 'cleaner'`, or always show for consistency.

### 4. Organizer tab – Artist suggestions

- **Purpose**: Show artist name clustering results and a CTA to run analysis.
- **Component**: `frontend/src/lib/components/analysis/OrganizerView.svelte` (or under `layout/` if you prefer).
- **Layout**:
  - **Header**: Title “Artist suggestions” (or use i18n), optional short description (“Similar artist names detected”).
  - **Run analysis**: If no library: show empty state “Select a library to analyze” + link to Browse. If library and no results yet: show “Run analysis” button (and optional “Analyzing…” when `analysis.loading`). If error: show `analysis.error` inline.
  - **List**: When `analysis.artistSuggestions.length > 0`: list of **ArtistSuggestionCard** (see below). Optional: group by confidence (High / Medium / Low) or show a small badge per card.
- **Empty state (after run, 0 suggestions)**: “No similar artist names found” – positive message.
- **Component**: `frontend/src/lib/components/analysis/ArtistSuggestionCard.svelte`
  - **Props**: `suggestion: ArtistSuggestion`.
  - **Content**: `original` → `suggested` (e.g. “guns n roses” → “Guns N' Roses”), `score` (e.g. “92% match”), `confidence_level` badge, optional `reason`.
  - **Actions**: Placeholder button “Use suggested” / “Apply” (no-op or console.log for now); optional “Dismiss”.
  - **Style**: Card (border, padding), hover state; match existing design (e.g. surface, border, text hierarchy).

### 5. Cleaner tab – Duplicate groups

- **Purpose**: Show duplicate file groups and a CTA to run analysis.
- **Component**: `frontend/src/lib/components/analysis/CleanerView.svelte`.
- **Layout**:
  - **Header**: “Duplicate tracks” (or i18n), short description (“Exact duplicate files in this library”).
  - **Run analysis**: Same pattern as Organizer (no library → empty; library + no results → “Run analysis”; loading; error).
  - **List**: When `analysis.duplicateGroups.length > 0`: list of **DuplicateGroupCard** (see below). Optional: summary line “N duplicate groups” above the list.
  - **Empty state (after run, 0 groups)**: “No duplicates found” – positive message.
- **Component**: `frontend/src/lib/components/analysis/DuplicateGroupCard.svelte` (inspiration: backup `DuplicateCluster.svelte`)
  - **Props**: `group: Track[]`, optional `onResolve?: (trackId, action) => void` for future use.
  - **Content**:
    - **Header**: Collapsible; first track’s `filename` (or title), “and N duplicates”; optional “Save X MB” (e.g. `(group.length - 1) * group[0].size`); optional short hash preview (`hash_partial?.slice(0, 8)`).
    - **Body (when expanded)**: One row per track: path (truncated), size (formatted), optional bitrate; “Delete” (or “Remove”) placeholder button that calls `onResolve(track.id, 'delete')` if provided.
  - **Style**: Border, rounded, surface background; expand/collapse icon; match backup look without external deps (no clsx if not in project – use class strings).
- **Future**: Bulk actions (“Keep highest bitrate”, “Keep oldest”) can be a bar above the list or in the header; for this phase, optional placeholder buttons in the card footer (no-op).

### 6. Shared / layout

- **Run analysis button**: Can live in TopBar (e.g. “Analyze” next to Browse) or inside Organizer/Cleaner headers. Recommendation: **TopBar** when `activeTab` is organizer or cleaner, so one place to trigger analysis; disable when no library or loading.
- **Loading**: In Organizer/Cleaner, show a spinner and “Analyzing library…” when `analysis.loading`; optionally disable Run analysis and tab switch during load.
- **Errors**: One place (e.g. small banner at top of Organizer/Cleaner content or under the Run button) showing `analysis.error` with optional “Dismiss”.
- **i18n**: Add keys for “Run analysis”, “Analyzing…”, “Artist suggestions”, “Duplicate tracks”, “No similar artist names found”, “No duplicates found”, “Select a library to analyze”, “Use suggested”, “Dismiss”, “Save X MB”, “Delete”, etc., in `frontend/src/lib/stores/i18n.ts`.

### 7. File list (implementation order)

| Step | File | Action |
|------|------|--------|
| 1 | `frontend/src/lib/types.ts` | Add `ArtistSuggestion` interface. |
| 2 | `frontend/src/lib/stores/analysis.ts` | Create analysis store (duplicateGroups, artistSuggestions, loading, error, setters). |
| 3 | `frontend/src/App.svelte` | Import AnalyzeArtists, DetectDuplicates; add handleRunAnalysis(); route organizer → OrganizerView, cleaner → CleanerView; pass currentLibraryPath and/or analysis store. |
| 4 | `frontend/src/lib/components/layout/TopBar.svelte` | Add “Run analysis” button (visible when tab is organizer or cleaner), disabled when !currentLibraryPath or analysis.loading; call onRunAnalysis. |
| 5 | `frontend/src/lib/components/analysis/ArtistSuggestionCard.svelte` | New; props suggestion; display original → suggested, score, confidence; placeholder Apply/Dismiss. |
| 6 | `frontend/src/lib/components/analysis/OrganizerView.svelte` | New; empty state / Run analysis / error / list of ArtistSuggestionCard. |
| 7 | `frontend/src/lib/components/analysis/DuplicateGroupCard.svelte` | New; collapsible group, header (filename, N duplicates, save MB), rows (path, size, Delete placeholder). |
| 8 | `frontend/src/lib/components/analysis/CleanerView.svelte` | New; empty state / Run analysis / error / list of DuplicateGroupCard. |
| 9 | `frontend/src/lib/stores/i18n.ts` | Add translation keys for analysis UI. |

### 8. Wails bindings

- After backend is built, run `wails dev` (or `wails build`) so that `frontend/wailsjs/go/main/App.js` (and `.d.ts`) include:
  - `AnalyzeArtists(root: string): Promise<ArtistSuggestion[]>`
  - `DetectDuplicates(root: string): Promise<Track[][]>`
- Map backend field names to frontend if needed (e.g. `artist_raw` → `artist` in Track; backend already returns `original`, `suggested`, `score`, `reason`, `confidence_level` for suggestions).

### 9. Validation / acceptance

- [ ] Select a library; switch to Organizer; click “Run analysis” → loading, then artist suggestions or “No similar artist names found”.
- [ ] Switch to Cleaner; click “Run analysis” (or use same run) → duplicate groups or “No duplicates found”.
- [ ] No library selected → Run analysis disabled or empty state with “Select a library”.
- [ ] Error from backend → message visible; loading stops.
- [ ] Duplicate group expands/collapses; each track shows path and size; Delete is a placeholder.
- [ ] Artist card shows original → suggested, score, confidence; Apply/Dismiss placeholders.
- [ ] No console errors; layout works with sidebar and context panel.

### 10. Optional (later) — keep features separated

- Toast on “Analysis complete” (e.g. “Found N artist suggestions and M duplicate groups”).
- **Apply artist suggestion** → implemented in **Phase 5 (Organize)** as a tag-write/rename action, not in Phase 4.
- **Resolve duplicate** (delete file, update DB) → implemented in **Phase 6 (Cleaner)**.
- Bulk actions in Cleaner (keep best bitrate, keep oldest) → Phase 6.

---

## Summary

- **Analysis store** holds duplicate groups, artist suggestions, loading, and error.
- **App.svelte** runs `AnalyzeArtists(currentLibraryPath)` and `DetectDuplicates(currentLibraryPath)` on “Run analysis”, updates store, and routes Organizer/Cleaner to new views.
- **OrganizerView** shows artist suggestion cards and empty/run/error states.
- **CleanerView** shows duplicate group cards (accordion-style) and empty/run/error states.
- **TopBar** gets “Run analysis” when on Organizer or Cleaner.
- Types and i18n extended; Wails bindings generated by build. This completes the Phase 4 UI plan with good UX inspired by the backup and aligned with the current app structure.
