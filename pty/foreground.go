package pty

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// ForegroundProcessName returns the current foreground process name for this PTY.
func (p *PTY) ForegroundProcessName() (string, error) {
	pgid, err := unix.IoctlGetInt(int(p.master.Fd()), unix.TIOCGPGRP)
	if err != nil {
		return "", err
	}
	out, err := exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(pgid)).Output()
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(out))
	if name == "" {
		return "", nil
	}
	return filepath.Base(name), nil
}
