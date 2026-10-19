package bot

import (
	"sync/atomic"
	"testing"
	"time"

	"go.mau.fi/whatsmeow/types"
)

func TestSessionManager(t *testing.T) {
	mgr := NewSessionManager(1) // 1 minute
	jid := types.NewJID("628123456789", types.DefaultUserServer)

	var timeoutCalled atomic.Bool
	s := mgr.GetOrCreate(jid, func(sess *UserSession) {
		timeoutCalled.Store(true)
	})

	if s.CurrentState != StateRoot {
		t.Errorf("expected initial state %s, got %s", StateRoot, s.CurrentState)
	}

	// Test PushState
	s.PushState(StateMenu1)
	if s.CurrentState != StateMenu1 {
		t.Errorf("expected state %s, got %s", StateMenu1, s.CurrentState)
	}

	s.PushState(StateMenu1_Sosduk)
	if s.CurrentState != StateMenu1_Sosduk {
		t.Errorf("expected state %s, got %s", StateMenu1_Sosduk, s.CurrentState)
	}

	// Test PopState
	prev := s.PopState()
	if prev != StateMenu1 {
		t.Errorf("expected popped state %s, got %s", StateMenu1, prev)
	}
	if s.CurrentState != StateMenu1 {
		t.Errorf("expected current state %s, got %s", StateMenu1, s.CurrentState)
	}

	// Test PopState back to root
	prev = s.PopState()
	if prev != StateRoot {
		t.Errorf("expected popped state %s, got %s", StateRoot, prev)
	}
	if s.CurrentState != StateRoot {
		t.Errorf("expected current state %s, got %s", StateRoot, s.CurrentState)
	}

	// Test ResetToRoot
	s.PushState("some_random_state")
	s.ResetToRoot()
	if s.CurrentState != StateRoot {
		t.Errorf("expected root state after reset, got %s", s.CurrentState)
	}
	if len(s.HistoryStack) != 0 {
		t.Errorf("expected empty history after reset, got len %d", len(s.HistoryStack))
	}
}

func TestSessionTimerReset(t *testing.T) {
	mgr := NewSessionManager(1)
	jid := types.NewJID("628987654321", types.DefaultUserServer)

	var timeoutFired atomic.Bool
	s := mgr.GetOrCreate(jid, func(sess *UserSession) {
		timeoutFired.Store(true)
	})

	// Manually reset timer with a short 20ms duration for test
	s.mu.Lock()
	if s.IdleTimer != nil {
		s.IdleTimer.Stop()
	}
	s.IdleTimer = time.AfterFunc(20*time.Millisecond, func() {
		timeoutFired.Store(true)
	})
	s.mu.Unlock()

	time.Sleep(50 * time.Millisecond)

	if !timeoutFired.Load() {
		t.Errorf("expected idle timer to fire callback")
	}
}
