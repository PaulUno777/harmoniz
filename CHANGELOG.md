# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.5.0]: https://github.com/PaulUno777/harmoniz/releases/tag/v0.5.0

## [0.6.0] - 2026-06-17

### Added
- Update checking (GitHub Releases), sidebar version + Update pill
- README screenshots
- Refine the organizer and the cleaner 

### Changed
- Rescan / duplicate detection improvements
- …

[0.6.0]: https://github.com/PaulUno777/harmoniz/releases/tag/v0.6.0
