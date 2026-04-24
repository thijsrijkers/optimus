<p align="center">
  <img src="etc/transparent.png" alt="Optimus logo" width="180" />
</p>

<h1 align="center">Optimus</h1>

Native terminal emulator built in Go with Gio.
Multi-tab desktop terminal with PTY-backed sessions, ANSI parsing, and mouse/clipboard integration.

## Current Status

Optimus is an early-stage terminal emulator with a functional base.

What is already solid:

- PTY I/O pipeline
- parser-to-buffer write flow
- multi-tab terminal sessions
- core ANSI styling and mode handling
- practical Neovim usability (keyboard + mouse)

What is still evolving:

- full standards-level escape sequence coverage
- more complete xterm parity and corner-case behaviors
- richer tab/window management (drag reorder, detach, etc.)
- diagnostics and profiling tooling

## Roadmap (Short-Term)

- Improve parser/emulation completeness for broader app compatibility
- Add tab interactions (hover states, middle-click close, reordering)
- Improve resize behavior consistency across monitor scale scenarios
- Add configurable settings (font family/size, theme, keybinds)
- Add integration tests around parser + buffer state transitions

## Build and Run

Requirements:

- Go `1.25+`
- macOS/Linux environment with PTY support

Run:

```bash
go run ./main.go
```

Build:

```bash
go build ./...
```

## Installation

### macOS

To produce and install a launchable `.app` bundle (using the icon from `etc/app_icon.png`):

```bash
./etc/macos/build_app_bundle.sh
cp -R "dist/macos/Optimus.app" /Applications/
```

This creates:

- `dist/macos/Optimus.app`

Launch it from Finder (Applications) or with:

```bash
open /Applications/Optimus.app
```

Note: when launched as a macOS app, Optimus starts your shell as a login shell so your profile configuration (`~/.zprofile`, `~/.zshrc`, etc.) is loaded. This helps keep completion and TUI behavior (like Neovim) consistent.

### Linux - Windows

Coming soon.

## Keyboard Shortcuts

- `Cmd/Ctrl + T`: New tab
- `Cmd/Ctrl + W`: Close active tab
- `Cmd/Ctrl + Tab`: Next tab
- `Cmd/Ctrl + Shift + Tab`: Previous tab
- `Cmd/Ctrl + C`: Copy selection
- `Cmd/Ctrl + V`: Paste
- `Cmd/Ctrl + +` (or `=`): Zoom in
- `Cmd/Ctrl + -`: Zoom out

Notes:

- On non-macOS, copy/paste follows terminal-friendly modifier behavior from current keybinding logic.
- In terminal apps that enable mouse reporting (e.g. Neovim), Shift-modified mouse can be used to force local selection behavior.
