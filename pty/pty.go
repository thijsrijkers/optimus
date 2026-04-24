package pty

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type PTY struct {
	master *os.File
	cmd    *exec.Cmd
}

func New(shell string, cols, rows int) (*PTY, error) {
	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"TERM_PROGRAM=optimus",
	)

	master, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	})

	if err != nil {
		return nil, err
	}

	return &PTY{master: master, cmd: cmd}, nil
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
