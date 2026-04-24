package tabs

import "fmt"

type TabMeta struct {
	ID      int
	Title   string
	Active  bool
	Closing bool
}

type Manager struct {
	shell      string
	cols       int
	rows       int
	invalidate func()
	nextID     int
	active     int
	sessions   []*Session
}

func New(shell string, cols, rows int, invalidate func()) (*Manager, error) {
	m := &Manager{shell: shell, cols: cols, rows: rows, invalidate: invalidate, nextID: 1}
	if err := m.NewTab(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) Active() *Session {
	if len(m.sessions) == 0 {
		return nil
	}
	if m.active < 0 || m.active >= len(m.sessions) {
		m.active = 0
	}
	return m.sessions[m.active]
}

func (m *Manager) NewTab() error {
	id := m.nextID
	m.nextID++
	s, err := newSession(id, m.shell, m.cols, m.rows, m.invalidate)
	if err != nil {
		return err
	}
	m.sessions = append(m.sessions, s)
	m.active = len(m.sessions) - 1
	if m.invalidate != nil {
		m.invalidate()
	}
	return nil
}

func (m *Manager) CloseActive() {
	m.CloseAt(m.active)
}

func (m *Manager) CloseAt(index int) {
	if index < 0 || index >= len(m.sessions) {
		return
	}
	if len(m.sessions) <= 1 {
		return
	}
	idx := index
	m.sessions[idx].Close()
	m.sessions = append(m.sessions[:idx], m.sessions[idx+1:]...)
	if idx >= len(m.sessions) {
		m.active = len(m.sessions) - 1
	} else if idx < m.active {
		m.active--
	}
	if m.active < 0 {
		m.active = 0
	}
	if m.invalidate != nil {
		m.invalidate()
	}
}

func (m *Manager) Next() {
	if len(m.sessions) <= 1 {
		return
	}
	m.active = (m.active + 1) % len(m.sessions)
	if m.invalidate != nil {
		m.invalidate()
	}
}

func (m *Manager) ActivateAt(index int) {
	if index < 0 || index >= len(m.sessions) {
		return
	}
	m.active = index
	if m.invalidate != nil {
		m.invalidate()
	}
}

func (m *Manager) Prev() {
	if len(m.sessions) <= 1 {
		return
	}
	m.active = (m.active - 1 + len(m.sessions)) % len(m.sessions)
	if m.invalidate != nil {
		m.invalidate()
	}
}

func (m *Manager) ResizeAll(cols, rows int) {
	m.cols, m.rows = cols, rows
	for _, s := range m.sessions {
		s.Resize(cols, rows)
	}
}

func (m *Manager) CloseAll() {
	for _, s := range m.sessions {
		s.Close()
	}
	m.sessions = nil
}

func (m *Manager) List() []TabMeta {
	out := make([]TabMeta, 0, len(m.sessions))
	for i, s := range m.sessions {
		title := s.Title()
		if title == "" {
			title = fmt.Sprintf("Tab %d", s.ID())
		}
		out = append(out, TabMeta{
			ID:     s.ID(),
			Title:  title,
			Active: i == m.active,
		})
	}
	return out
}
