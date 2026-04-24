# optimus

Optimus is a small terminal emulator written in Go, using Gio for rendering.

## What it does

- Opens a GUI window titled `Optimus`.
- Starts a shell (`/bin/sh`) in a PTY.
- Sends keyboard input from the window to the shell.
- Renders a terminal cell grid with cursor, colors, and basic text attributes.

## Current status

This project is in early development.

- Core windowing and PTY plumbing are in place.
- Screen buffer logic exists (cursor movement, resize, scrolling, erase behavior, alt screen support).
- Key mappings for common keys are implemented.
- PTY output parsing in `terminal/terminal.go` is not finished yet, so full terminal emulation is still incomplete.

## Run

```bash
go run .
```

## Build

```bash
go build ./...
```

## Project layout

- `main.go` - app entry point.
- `windowing/` - Gio window event loop, drawing, keyboard handling.
- `pty/` - PTY process lifecycle and resize/read/write wrappers.
- `terminal/` - terminal buffer and emulator core.
