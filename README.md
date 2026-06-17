<p align="center">
  <img src="build/appicon.png" alt="Harmoniz" width="128" height="128">
</p>

<h1 align="center">Harmoniz</h1>

<p align="center">
  A high-performance, transactional audio library manager built with Go, Wails, and Svelte.<br>
  Organize, tag, and deduplicate 100k+ files safely.
</p>

<p align="center">
  <a href="https://github.com/PaulUno777/harmoniz/actions/workflows/ci.yml"><img src="https://github.com/PaulUno777/harmoniz/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/PaulUno777/harmoniz/releases"><img src="https://img.shields.io/github/v/release/PaulUno777/harmoniz?label=release" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
</p>

## Features

- **Fast library scanning** — SQLite-backed catalog designed for large collections (100k+ tracks)
- **Virtualized browser** — Search, filter, and browse your library without UI lag
- **Artist clustering** — Detect similar artist names and suggest normalization
- **Duplicate detection** — Multi-stage funnel (size → partial hash → full hash)
- **File organizer** — Rename files with templates and write metadata tags
- **In-app playback** — Floating player for previewing tracks
- **Themes & i18n** — Dark/light mode with English and French UI

## Screenshots

> Add screenshots to [`docs/screenshots/`](docs/screenshots/) and embed them here before the first public release.

## Download

Pre-built binaries are on the [Releases](https://github.com/PaulUno777/harmoniz/releases) page.

> **Important:** download the **platform binary** below — **not** "Source code (zip)". The source archive is for developers; it does not contain a runnable app.

| Platform | Download | Install |
|----------|----------|---------|
| macOS | [![Download macOS](https://img.shields.io/badge/Download-macOS-blue)](https://github.com/PaulUno777/harmoniz/releases/latest/download/harmoniz.app.zip) | See [macOS install](#macos) below |
| Windows | [![Download Windows](https://img.shields.io/badge/Download-Windows-blue)](https://github.com/PaulUno777/harmoniz/releases/latest/download/harmoniz-amd64-installer.exe) | Run the installer (WebView2 embedded) |
| Linux | [![Download Linux](https://img.shields.io/badge/Download-Linux-blue)](https://github.com/PaulUno777/harmoniz/releases/latest/download/harmoniz-linux-amd64.tar.gz) | `tar -xzf harmoniz-linux-amd64.tar.gz && ./harmoniz` |

> Binaries are currently **unsigned**. macOS and Windows may show security warnings on first launch.

### macOS

1. Download [harmoniz.app.zip](https://github.com/PaulUno777/harmoniz/releases/latest/download/harmoniz.app.zip)
2. Double-click to unzip — you get `harmoniz.app`
3. Drag `harmoniz.app` to **Applications**
4. Open from Applications
   - If macOS blocks it (unsigned app): **Right-click** → **Open** → **Open**

### Windows

1. Download [harmoniz-amd64-installer.exe](https://github.com/PaulUno777/harmoniz/releases/latest/download/harmoniz-amd64-installer.exe)
2. Run the installer and follow the prompts
3. Launch Harmoniz from the Start menu

### Linux

1. Download [harmoniz-linux-amd64.tar.gz](https://github.com/PaulUno777/harmoniz/releases/latest/download/harmoniz-linux-amd64.tar.gz)
2. Extract and run:

```bash
tar -xzf harmoniz-linux-amd64.tar.gz
chmod +x harmoniz
./harmoniz
```

Requires GTK 3 and WebKit2GTK.

## Quick start

1. Download and install for your platform (see [Download](#download) above).
2. Launch Harmoniz.
3. Click **Add library** and select a folder containing your audio files.
4. Wait for the scan to complete, then browse, analyze, and organize your collection.

Application data is stored in `~/.harmoniz/` (library database and settings).

## Development

### Prerequisites

- Go 1.24+
- Node.js 20+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2.11+

### Live development

```bash
wails dev
```

This starts a Vite dev server with hot reload. Go methods are callable from the embedded webview (and from http://localhost:34115 in a browser during dev).

### Run checks

```bash
cd frontend && npm ci && npm run build
go test ./...
cd frontend && npm run check
```

### Build locally

```bash
wails build
```

Platform-specific builds:

```bash
wails build -platform darwin/universal
wails build -platform windows/amd64 -nsis
wails build -platform linux/amd64
```

Output is written to `build/bin/`.

## Releasing

Releases are automated when a semver tag is pushed to `main`:

```bash
git tag v0.5.0
git push origin v0.5.0
```

GitHub Actions builds macOS, Windows, and Linux artifacts and publishes a GitHub Release. See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and the pull request process.

## License

[MIT](LICENSE) — Copyright © 2026 Paulin NZODOUM
