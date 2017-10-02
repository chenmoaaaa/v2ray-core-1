package mux

import (
	"io"
	"sync"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/transport/ray"
)

type SessionManager struct {
	sync.RWMutex
	sessions map[uint16]*Session
	count    uint16
	closed   bool
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		count:    0,
		sessions: make(map[uint16]*Session, 32),
	}
}

func (m *SessionManager) Size() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.sessions)
}

func (m *SessionManager) Count() int {
	m.RLock()
	defer m.RUnlock()

	return int(m.count)
}

func (m *SessionManager) Allocate() *Session {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return nil
	}

	m.count++
	s := &Session{
		ID:     m.count,
		parent: m,
	}
	m.sessions[s.ID] = s
	return s
}

func (m *SessionManager) Add(s *Session) {
	m.Lock()
	defer m.Unlock()

	m.sessions[s.ID] = s
}

func (m *SessionManager) Remove(id uint16) {
	m.Lock()
	defer m.Unlock()

	delete(m.sessions, id)
}

func (m *SessionManager) Get(id uint16) (*Session, bool) {
	m.RLock()
	defer m.RUnlock()

	if m.closed {
		return nil, false
	}

	s, found := m.sessions[id]
	return s, found
}

func (m *SessionManager) CloseIfNoSession() bool {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return true
	}

	if len(m.sessions) != 0 {
		return false
	}

	m.closed = true
	return true
}

func (m *SessionManager) Close() {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return
	}

	m.closed = true

	for _, s := range m.sessions {
		s.input.Close()
		s.output.Close()
	}

	m.sessions = make(map[uint16]*Session)
}

type Session struct {
	input        ray.InputStream
	output       ray.OutputStream
	parent       *SessionManager
	ID           uint16
	transferType protocol.TransferType
}

func (s *Session) Close() {
	s.output.Close()
	s.input.Close()
	s.parent.Remove(s.ID)
}

func (s *Session) NewReader(reader io.Reader) buf.Reader {
	if s.transferType == protocol.TransferTypeStream {
		return NewStreamReader(reader)
	}
	return NewPacketReader(reader)
}
