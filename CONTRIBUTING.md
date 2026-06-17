# Contributing to Harmoniz

Thank you for your interest in contributing!

## Development setup

### Prerequisites

- [Go](https://go.dev/dl/) 1.24+
- [Node.js](https://nodejs.org/) 20+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2.11+

### Getting started

```bash
git clone https://github.com/PaulUno777/harmoniz.git
cd harmoniz
wails dev
```

### Running checks locally

```bash
# Backend tests
go test ./...

# Frontend type check
cd frontend && npm ci && npm run check
```

### Building locally

```bash
wails build
```

Platform-specific builds:

```bash
wails build -platform darwin/universal   # macOS
wails build -platform windows/amd64 -nsis  # Windows installer
wails build -platform linux/amd64          # Linux
```

## Git workflow

1. Fork the repository and create a feature branch from `main` (e.g. `feat/organizer-preview`).
2. Make your changes with clear, focused commits.
3. Run `go test ./...` and `npm run check` in `frontend/` before opening a PR.
4. Open a pull request against `main` with a short description of what changed and why.
5. Update `CHANGELOG.md` under `[Unreleased]` if the change is user-facing.

## Releases

Releases are automated via GitHub Actions when a semver tag is pushed:

```bash
git tag v0.5.0
git push origin v0.5.0
```

The release workflow builds unsigned binaries for macOS, Windows, and Linux and publishes them to [GitHub Releases](https://github.com/PaulUno777/harmoniz/releases).

## Code style

- Match existing patterns in the file you are editing.
- Keep changes minimal and focused on the task at hand.
- Go: standard `gofmt` formatting.
- Frontend: follow existing Svelte 5 and Tailwind conventions.

## Questions

Open a [GitHub issue](https://github.com/PaulUno777/harmoniz/issues) for bugs, feature requests, or questions.
