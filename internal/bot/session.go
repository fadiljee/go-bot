package bot

import (
	"sync"
	"time"

	"go.mau.fi/whatsmeow/types"
)

type UserSession struct {
	mu           sync.Mutex
	JID          types.JID
	CurrentState string
	HistoryStack []string
	IdleTimer    *time.Timer
	LastActive   time.Time
}

func (s *UserSession) PushState(newState string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentState != "" && s.CurrentState != newState {
		s.HistoryStack = append(s.HistoryStack, s.CurrentState)
	}
	s.CurrentState = newState
	s.LastActive = time.Now()
}

func (s *UserSession) PopState() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.HistoryStack) == 0 {
		s.CurrentState = StateRoot
		return StateRoot
	}
	lastIdx := len(s.HistoryStack) - 1
	prevState := s.HistoryStack[lastIdx]
	s.HistoryStack = s.HistoryStack[:lastIdx]
	s.CurrentState = prevState
	s.LastActive = time.Now()
	return prevState
}

func (s *UserSession) ResetToRoot() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentState = StateRoot
	s.HistoryStack = nil
	s.LastActive = time.Now()
}

type SessionManager struct {
	sessions sync.Map
	timeout  time.Duration
}

func NewSessionManager(timeoutMinutes int) *SessionManager {
	return &SessionManager{
		timeout: time.Duration(timeoutMinutes) * time.Minute,
	}
}

func (sm *SessionManager) GetOrCreate(jid types.JID, onTimeout func(s *UserSession)) *UserSession {
	key := jid.ToNonAD().String()
	val, loaded := sm.sessions.Load(key)
	if loaded {
		sess := val.(*UserSession)
		sess.mu.Lock()
		sess.LastActive = time.Now()
		// Reset cancelable timer (Resolves TD-04)
		if sess.IdleTimer != nil {
			sess.IdleTimer.Stop()
		}
		if onTimeout != nil {
			sess.IdleTimer = time.AfterFunc(sm.timeout, func() {
				onTimeout(sess)
			})
		}
		sess.mu.Unlock()
		return sess
	}

	sess := &UserSession{
		JID:          jid,
		CurrentState: StateRoot,
		LastActive:   time.Now(),
	}

	if onTimeout != nil {
		sess.IdleTimer = time.AfterFunc(sm.timeout, func() {
			onTimeout(sess)
		})
	}

	sm.sessions.Store(key, sess)
	return sess
}

func (sm *SessionManager) Delete(jid types.JID) {
	key := jid.ToNonAD().String()
	if val, ok := sm.sessions.LoadAndDelete(key); ok {
		sess := val.(*UserSession)
		sess.mu.Lock()
		if sess.IdleTimer != nil {
			sess.IdleTimer.Stop()
		}
		sess.mu.Unlock()
	}
}
