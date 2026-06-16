# Phase 5: Organize (Rename, Tag, Organize Library) — Plan Outline

## Purpose

Provide a **single, clearly separated** feature that allows users to:
- **Rename files** (by template or manually)
- **Set artist, title, album, year** (and other tags) on tracks
- **Organize the entire library** (e.g. move files into `Artist/Album/` structure based on metadata)

All of this is **write** operations (filesystem + tags + DB). It is **not** part of Library (read/scan/list) or Analysis (read-only suggestions).

## Feature separation

| Feature    | What it does                          | Writes? |
|-----------|----------------------------------------|--------|
| Library   | Browse, scan, list tracks              | No     |
| Analysis  | Artist clustering, duplicate detection | No     |
| **Organize** | Rename files, set tags, organize library | **Yes** (files + tags + DB) |
| Cleaner   | Resolve duplicates, prune, empty dirs  | Yes (delete/prune) |

- **Organize** = rename + tag + folder structure. Own UI flow and backend services.
- **Organizer tab (Phase 4)** = analysis views (artist suggestions, etc.). Do not use the same tab for both analysis and rename/tag; use a distinct “Organize” entry (e.g. context panel edit, or dedicated Organize tab/section for batch).

## Backend (outline)

- **Rename service** (`internal/core/services/rename/`): Pre-calculate targets, check collisions, rename via temp → final, call `UpdateTrackPath`.
- **Tag write**: Port + adapter for writing tags (e.g. `bogem/id3v2` for MP3); read already via `dhowden/tag` in scanner.
- **App methods**: e.g. `RenameTrack(oldPath, newPath)`, `UpdateTrackTags(id, artist, title, album, …)`, optional `OrganizeLibrary(root, template)`.
- **Repository**: `UpdateTrackPath(ctx, id, newPath)` — already in ports; implement in DB adapter if stub.
- **Transactions** (optional): Before-snapshots for undo; can be Phase 7.

## Frontend (outline)

- **Single-track edit**: e.g. in Library context panel — edit artist, title, album; “Save” → `UpdateTrackTags`; “Rename file” → `RenameTrack`.
- **Batch / organize**: Dedicated “Organize” tab or section — e.g. “Organize library by Artist/Album”, “Apply artist suggestion” (from Phase 4 analysis) as a tag+rename action.
- Keep **Organizer (analysis)** tab for suggestions only; **Organize** flow for all write actions.

## Reference

- `context/PLAN.md` § D (Transactional Renaming & Editing)
- `.cursor/plans/progressive_backend_integration_plan_eb6d9226.plan.md` — Phase 5 and Feature Separation table
- Backup: `internal/core/services/rename`, tag-write adapter (if present)
