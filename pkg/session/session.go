// Package session provides session management for WebSocket connections.
package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ID represents a unique session identifier.
type ID string

// Session represents a WebSocket session with its state and message log.
type Session struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ID               ID
	Model            string
	ResumptionHandle string
	Log              []any
	mu               sync.Mutex
	State            State
}

// NewSession creates a new session with the specified model.
func NewSession(model string) *Session {
	now := time.Now()
	return &Session{
		ID:        ID(uuid.New().String()),
		Model:     model,
		State:     StateConnecting,
		CreatedAt: now,
		UpdatedAt: now,
		Log:       []any{},
	}
}

// Append adds a message to the session log.
func (s *Session) Append(message any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Log = append(s.Log, message)
	s.UpdatedAt = time.Now()
}

// Store manages multiple sessions with thread-safe operations.
type Store struct {
	sessions map[ID]*Session
	mu       sync.RWMutex
}

// NewStore creates a new session store.
func NewStore() *Store {
	return &Store{
		sessions: make(map[ID]*Session),
	}
}

// Put stores a session in the store.
func (s *Store) Put(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
}

// Get retrieves a session by ID. Returns the session and true if found.
func (s *Store) Get(id ID) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[id]
	return session, exists
}

// Delete removes a session from the store.
func (s *Store) Delete(id ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}
