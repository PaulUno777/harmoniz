# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.0] - 2026-06-17

### Added

- **Algorithm configuration panel** — adjust Cleaner and Organizer precision from the UI (similarity thresholds, artist grouping, field confidence)
- **Persistent app config** — algorithm settings saved to disk via `GetConfig` / `UpdateConfig`; services reload when values change
- **Fuzzy duplicate detection** — groups similar titles and filenames (Jaro–Winkler), with `kind` labels: `exact`, `similar_title`, `similar_filename`
- **Duplicate quality scoring** — prefers tracks with richer metadata and penalizes obvious copy filenames (e.g. `song (1).mp3`)
- **Status bar file context** — shows the selected track filename in the footer
- Unit tests for deduplication, clustering config, and settings persistence

### Changed

- Organizer suggestions use the configurable field-confidence threshold instead of a hard-coded value
- Sidebar and status bar layout refinements; update prompt accessible from the footer
- Updated application icon

## [0.6.0] - 2026-06-17

### Added

- In-app update checking against GitHub Releases (`GetAppVersion`, `CheckForUpdates`)
- Sidebar version label and Update pill; settings page update UI with download and dismiss
- README screenshots for Library, Organizer, Cleaner, and Settings

### Changed

- Library rescan workflow and duplicate detection pipeline improvements
- Cleaner and Organizer refinements

## [0.5.0] - 2026-06-17

### Added

- Desktop app for macOS, Windows, and Linux (Wails + Go + Svelte)
- Library scanning with SQLite-backed catalog (100k+ track scale)
- Virtualized library browser with search and filtering
- Artist clustering analysis (fuzzy name matching, normalization suggestions)
- Duplicate detection (size → partial hash → full hash funnel)
- File organizer with rename templates and metadata writing
- In-app audio playback with floating player
- Dark/light theme support and internationalization (EN/FR)
- Automated CI (tests + type checks) and tag-based multi-platform releases

### Changed

- Production builds use Info-level logging and no DevTools inspector on startup

[0.7.0]: https://github.com/PaulUno777/harmoniz/releases/tag/v0.7.0
[0.6.0]: https://github.com/PaulUno777/harmoniz/releases/tag/v0.6.0
[0.5.0]: https://github.com/PaulUno777/harmoniz/releases/tag/v0.5.0
