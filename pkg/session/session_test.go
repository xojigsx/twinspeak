package session

import (
	"sync"
	"testing"
	"time"
)

// TestNewSession tests session creation
func TestNewSession(t *testing.T) {
	model := "gemini-1.5-flash"
	session := NewSession(model)

	if session == nil {
		t.Fatal("NewSession returned nil")
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}

	if session.Model != model {
		t.Errorf("Expected model %s, got %s", model, session.Model)
	}

	if session.State != StateConnecting {
		t.Errorf("Expected initial state %s, got %s", StateConnecting, session.State)
	}

	if session.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if session.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	if session.Log == nil {
		t.Error("Log should be initialized")
	}

	if len(session.Log) != 0 {
		t.Error("Log should be empty initially")
	}
}

// TestSessionAppend tests message logging
func TestSessionAppend(t *testing.T) {
	session := NewSession("test-model")
	initialTime := session.UpdatedAt

	// Wait a bit to ensure timestamp changes
	time.Sleep(1 * time.Millisecond)

	message1 := map[string]interface{}{"type": "test", "data": "message1"}
	session.Append(message1)

	if len(session.Log) != 1 {
		t.Errorf("Expected log length 1, got %d", len(session.Log))
	}

	if session.UpdatedAt.Equal(initialTime) || session.UpdatedAt.Before(initialTime) {
		t.Error("UpdatedAt should be updated after append")
	}

	message2 := map[string]interface{}{"type": "test", "data": "message2"}
	session.Append(message2)

	if len(session.Log) != 2 {
		t.Errorf("Expected log length 2, got %d", len(session.Log))
	}

	// Verify messages are in correct order
	loggedMsg1, ok := session.Log[0].(map[string]interface{})
	if !ok {
		t.Fatal("First logged message is not a map")
	}
	if loggedMsg1["data"] != "message1" {
		t.Errorf("Expected first message data 'message1', got %v", loggedMsg1["data"])
	}

	loggedMsg2, ok := session.Log[1].(map[string]interface{})
	if !ok {
		t.Fatal("Second logged message is not a map")
	}
	if loggedMsg2["data"] != "message2" {
		t.Errorf("Expected second message data 'message2', got %v", loggedMsg2["data"])
	}
}

// TestSessionAppendConcurrency tests concurrent message logging
func TestSessionAppendConcurrency(t *testing.T) {
	session := NewSession("test-model")
	numGoroutines := 10
	messagesPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				message := map[string]interface{}{
					"goroutine": goroutineID,
					"message":   j,
				}
				session.Append(message)
			}
		}(i)
	}

	wg.Wait()

	expectedLength := numGoroutines * messagesPerGoroutine
	if len(session.Log) != expectedLength {
		t.Errorf("Expected log length %d, got %d", expectedLength, len(session.Log))
	}
}

// TestNewStore tests store creation
func TestNewStore(t *testing.T) {
	store := NewStore()

	if store == nil {
		t.Fatal("NewStore returned nil")
	}

	if store.sessions == nil {
		t.Error("Store sessions map should be initialized")
	}

	if len(store.sessions) != 0 {
		t.Error("Store should be empty initially")
	}
}

// TestStorePutGet tests storing and retrieving sessions
func TestStorePutGet(t *testing.T) {
	store := NewStore()
	session := NewSession("test-model")

	// Put session
	store.Put(session)

	// Get session
	retrieved, exists := store.Get(session.ID)
	if !exists {
		t.Error("Session should exist in store")
	}

	if retrieved == nil {
		t.Fatal("Retrieved session is nil")
	}

	if retrieved.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, retrieved.ID)
	}

	if retrieved.Model != session.Model {
		t.Errorf("Expected model %s, got %s", session.Model, retrieved.Model)
	}
}

// TestStoreGetNonExistent tests getting non-existent session
func TestStoreGetNonExistent(t *testing.T) {
	store := NewStore()
	nonExistentID := ID("non-existent-id")

	session, exists := store.Get(nonExistentID)
	if exists {
		t.Error("Non-existent session should not exist")
	}

	if session != nil {
		t.Error("Non-existent session should return nil")
	}
}

// TestStoreDelete tests session deletion
func TestStoreDelete(t *testing.T) {
	store := NewStore()
	session := NewSession("test-model")

	// Put session
	store.Put(session)

	// Verify it exists
	_, exists := store.Get(session.ID)
	if !exists {
		t.Error("Session should exist before deletion")
	}

	// Delete session
	store.Delete(session.ID)

	// Verify it's gone
	_, exists = store.Get(session.ID)
	if exists {
		t.Error("Session should not exist after deletion")
	}
}

// TestStoreDeleteNonExistent tests deleting non-existent session
func TestStoreDeleteNonExistent(t *testing.T) {
	store := NewStore()
	nonExistentID := ID("non-existent-id")

	// Should not panic or error
	store.Delete(nonExistentID)

	// Store should still be functional
	session := NewSession("test-model")
	store.Put(session)

	retrieved, exists := store.Get(session.ID)
	if !exists || retrieved == nil {
		t.Error("Store should still be functional after deleting non-existent session")
	}
}

// TestStoreConcurrency tests concurrent store operations
func TestStoreConcurrency(t *testing.T) {
	store := NewStore()
	numGoroutines := 10
	sessionsPerGoroutine := 100

	sessions := make([]*Session, numGoroutines*sessionsPerGoroutine)

	// Create sessions
	for i := 0; i < numGoroutines*sessionsPerGoroutine; i++ {
		sessions[i] = NewSession("test-model")
	}

	// First, put all sessions concurrently
	var putWg sync.WaitGroup
	putWg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(start int) {
			defer putWg.Done()
			for j := 0; j < sessionsPerGoroutine; j++ {
				store.Put(sessions[start*sessionsPerGoroutine+j])
			}
		}(i)
	}
	putWg.Wait()

	// Then, test concurrent Get and Delete operations
	var opWg sync.WaitGroup
	opWg.Add(numGoroutines * 2) // Get and Delete operations

	// Concurrent Get operations
	for i := 0; i < numGoroutines; i++ {
		go func(start int) {
			defer opWg.Done()
			for j := 0; j < sessionsPerGoroutine; j++ {
				sessionID := sessions[start*sessionsPerGoroutine+j].ID
				_, _ = store.Get(sessionID)
			}
		}(i)
	}

	// Concurrent Delete operations (delete half of the sessions)
	for i := 0; i < numGoroutines; i++ {
		go func(start int) {
			defer opWg.Done()
			for j := 0; j < sessionsPerGoroutine/2; j++ {
				sessionID := sessions[start*sessionsPerGoroutine+j].ID
				store.Delete(sessionID)
			}
		}(i)
	}

	opWg.Wait()

	// Verify that store operations completed without panics
	// The exact count may vary due to race conditions, but we can verify
	// that the store is still functional
	testSession := NewSession("test-model")
	store.Put(testSession)

	retrieved, exists := store.Get(testSession.ID)
	if !exists || retrieved == nil {
		t.Error("Store should still be functional after concurrent operations")
	}
}

// TestStateString tests state string representation
func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateConnecting, "Connecting"},
		{StateConfigured, "Configured"},
		{StateActive, "Active"},
		{StateClosing, "Closing"},
		{StateClosed, "Closed"},
		{State(999), "Unknown"}, // Invalid state
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestSessionIDUniqueness tests that session IDs are unique
func TestSessionIDUniqueness(t *testing.T) {
	numSessions := 1000
	sessions := make([]*Session, numSessions)
	idMap := make(map[ID]bool)

	for i := 0; i < numSessions; i++ {
		sessions[i] = NewSession("test-model")

		if idMap[sessions[i].ID] {
			t.Errorf("Duplicate session ID found: %s", sessions[i].ID)
		}
		idMap[sessions[i].ID] = true
	}

	if len(idMap) != numSessions {
		t.Errorf("Expected %d unique IDs, got %d", numSessions, len(idMap))
	}
}
