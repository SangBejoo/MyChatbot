package infrastructure

import (
	"sync"
	"time"
)

// UserSession tracks active operations per user
type UserSession struct {
	ChatID      int64
	IsProcessing bool
	LastClick   time.Time
	mu          sync.Mutex
}

// SessionManager manages user sessions globally
type SessionManager struct {
	sessions map[int64]*UserSession
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[int64]*UserSession),
	}
}

// GetOrCreateSession returns or creates a user session
func (sm *SessionManager) GetOrCreateSession(chatID int64) *UserSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[chatID]
	if !exists {
		session = &UserSession{ChatID: chatID}
		sm.sessions[chatID] = session
	}
	return session
}

// IsAllowedClick checks if the click is allowed (debouncing)
// Returns true if allowed, false if spam/duplicate
func (us *UserSession) IsAllowedClick() bool {
	us.mu.Lock()
	defer us.mu.Unlock()

	// If already processing, deny
	if us.IsProcessing {
		return false
	}

	// If last click was within 2 seconds, deny (debounce)
	if time.Since(us.LastClick) < 2*time.Second {
		return false
	}

	us.LastClick = time.Now()
	return true
}

// StartProcessing marks session as processing
func (us *UserSession) StartProcessing() {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.IsProcessing = true
}

// FinishProcessing marks session as done
func (us *UserSession) FinishProcessing() {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.IsProcessing = false
}
