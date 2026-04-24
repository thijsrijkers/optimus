package pty

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/creack/pty"
)

type PTY struct {
	master *os.File
	cmd    *exec.Cmd
}

func New(shell string, cols, rows int) (*PTY, error) {
	shell = filepath.Clean(shell)
	cmd := exec.Command(shell, "-l")

	home, _ := os.UserHomeDir()
	if home != "" {
		cmd.Dir = home
	}
	cmd.Env = withEnvDefaults(os.Environ(), map[string]string{
		"SHELL":        shell,
		"HOME":         home,
		"PWD":          home,
		"PATH":         "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin",
		"LANG":         "en_US.UTF-8",
		"LC_CTYPE":     "en_US.UTF-8",
		"TERM":         "xterm-256color",
		"COLORTERM":    "truecolor",
		"TERM_PROGRAM": "optimus",
	})

	master, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	})

	if err != nil {
		return nil, err
	}

	return &PTY{master: master, cmd: cmd}, nil
}

func withEnvDefaults(base []string, defaults map[string]string) []string {
	env := make([]string, 0, len(base)+len(defaults))
	seen := make(map[string]struct{}, len(base))

	for _, entry := range base {
		env = append(env, entry)
		eq := -1
		for i := 0; i < len(entry); i++ {
			if entry[i] == '=' {
				eq = i
				break
			}
		}
		if eq <= 0 {
			continue
		}
		seen[entry[:eq]] = struct{}{}
	}

	for key, val := range defaults {
		if val == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		env = append(env, key+"="+val)
	}

	return env
}

func (p *PTY) Write(data []byte) (int, error) {
	return p.master.Write(data)
}

func (p *PTY) Resize(cols, rows int) error {
	return pty.Setsize(p.master, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	})
}

func (p *PTY) Read(buf []byte) (int, error) {
	return p.master.Read(buf)
}

func (p *PTY) Close() error {
	p.cmd.Process.Kill()
	return p.master.Close()
}
