#Optimus

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

## Features

Current implemented features include:

- PTY-backed shell session in a native Gio window
- Multi-tab support:
  - create new tab
  - switch tabs
  - close tabs
  - click tab to activate
  - click `+` to create
  - click `×` to close
- Tab titles:
  - OSC title support (`OSC 0` / `OSC 2`)
  - fallback to foreground process name (e.g. `nvim`, `bash`)
- Keyboard input:
  - printable text + control keys
  - Ctrl shortcuts mapping
  - tab/window shortcuts
- Mouse support:
  - terminal mouse protocol forwarding (including SGR mode)
  - local selection + clipboard copy/paste when not captured by app mode
- Clipboard integration:
  - copy selection
  - paste into PTY
- Rendering:
  - cell-based text renderer
  - cursor rendering
  - color attributes (ANSI 16 + 256 + truecolor via SGR)
  - Ghostty-inspired tab pill styling
- Buffer behavior:
  - cursor movement
  - erase operations
  - line/char insert-delete variants
  - scrolling regions
  - alternate screen support

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

Install:

```bash
go install .
```

