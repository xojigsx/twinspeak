package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type ID string

type Session struct {
	ID               ID
	Model            string
	State            State
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ResumptionHandle string
	Log              []any
	mu               sync.Mutex
}

func NewSession(model string) *Session {
	now := time.Now()
	return &Session{
		ID:        ID(uuid.New().String()),
		Model:     model,
		State:     StateConnecting,
		CreatedAt: now,
		UpdatedAt: now,
		Log:       make([]any, 0),
	}
}

func (s *Session) Append(message any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Log = append(s.Log, message)
	s.UpdatedAt = time.Now()
}

type Store struct {
	sessions map[ID]*Session
	mu       sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[ID]*Session),
	}
}

func (s *Store) Put(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
}

func (s *Store) Get(id ID) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[id]
	return session, exists
}

func (s *Store) Delete(id ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}
