package tabs

import (
	"sync"

	"optimus/core"
	"optimus/pty"
)

type Session struct {
	id       int
	terminal *core.Terminal
	pty      *pty.PTY
	mu       sync.Mutex
}

func newSession(id int, shell string, cols, rows int, invalidate func()) (*Session, error) {
	p, err := pty.New(shell, cols, rows)
	if err != nil {
		return nil, err
	}
	s := &Session{
		id:       id,
		terminal: core.New(cols, rows),
		pty:      p,
	}
	go s.readLoop(invalidate)
	return s, nil
}

func (s *Session) readLoop(invalidate func()) {
	buf := make([]byte, 4096)
	for {
		n, err := s.pty.Read(buf)
		if err != nil {
			return
		}
		s.mu.Lock()
		s.terminal.Write(buf[:n])
		s.mu.Unlock()
		if invalidate != nil {
			invalidate()
		}
	}
}

func (s *Session) ID() int { return s.id }

func (s *Session) Title() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t := s.terminal.Title(); t != "" {
		return t
	}
	if app, err := s.pty.ForegroundProcessName(); err == nil && app != "" {
		return app
	}
	return ""
}

func (s *Session) Lock() { s.mu.Lock() }

func (s *Session) Unlock() { s.mu.Unlock() }

func (s *Session) Terminal() *core.Terminal { return s.terminal }

func (s *Session) WriteInput(data []byte) {
	if len(data) == 0 {
		return
	}
	_, _ = s.pty.Write(data)
}

func (s *Session) Resize(cols, rows int) {
	s.mu.Lock()
	s.terminal.Resize(cols, rows)
	s.mu.Unlock()
	_ = s.pty.Resize(cols, rows)
}

func (s *Session) Close() {
	_ = s.pty.Close()
}
