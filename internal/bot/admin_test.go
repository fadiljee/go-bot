package bot

import (
	"context"
	"testing"
	"time"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/whatsapp"
	"go.mau.fi/whatsmeow/types"
)

func TestAdminWhitelistAndLockout(t *testing.T) {
	cfg := &config.Config{
		AdminTriggerKeyword:   "login admin - 1903",
		AdminPassword:         "test_admin_pass_123",
		AdminWhitelistNumbers: []string{"6281234567890"},
		AdminMaxRetry:         3,
		AdminLockoutMinutes:   60,
	}

	router := NewBotRouter(cfg, nil, nil)

	// 1. Test whitelist check
	if !router.isAdminWhitelisted("6281234567890") {
		t.Error("expected 6281234567890 to be whitelisted")
	}
	if !router.isAdminWhitelisted("+6281234567890") {
		t.Error("expected +6281234567890 to be whitelisted after normalization")
	}
	if !router.isAdminWhitelisted("081234567890") {
		t.Error("expected 081234567890 to be whitelisted after 08->62 conversion")
	}
	if router.isAdminWhitelisted("6281111111111") {
		t.Error("did not expect 6281111111111 to be whitelisted")
	}

	// 2. Test lockout logic
	phone := "6281234567890"
	session := &UserSession{
		JID:          types.NewJID(phone, types.DefaultUserServer),
		CurrentState: StateAdminWaitPass,
	}

	// Attempt 1: wrong pass
	_ = router.HandleAdminAuth(context.Background(), session, &whatsapp.ParsedMessage{
		JID:  session.JID,
		Text: "wrongpass1",
	})
	if locked, _ := router.isAdminLockedOut(phone); locked {
		t.Error("user should not be locked out on 1st failure")
	}

	// Attempt 2: wrong pass
	_ = router.HandleAdminAuth(context.Background(), session, &whatsapp.ParsedMessage{
		JID:  session.JID,
		Text: "wrongpass2",
	})
	if locked, _ := router.isAdminLockedOut(phone); locked {
		t.Error("user should not be locked out on 2nd failure")
	}

	// Attempt 3: wrong pass -> triggers lockout
	_ = router.HandleAdminAuth(context.Background(), session, &whatsapp.ParsedMessage{
		JID:  session.JID,
		Text: "wrongpass3",
	})
	locked, until := router.isAdminLockedOut(phone)
	if !locked {
		t.Error("user should be locked out on 3rd failure")
	}
	if until.Before(time.Now().Add(50 * time.Minute)) {
		t.Errorf("expected lockout until ~60 min from now, got %v", until)
	}

	// Attempt 4: even if correct password is submitted during lockout, it should be rejected
	err := router.HandleAdminAuth(context.Background(), session, &whatsapp.ParsedMessage{
		JID:  session.JID,
		Text: "test_admin_pass_123",
	})
	// State should reset to root
	if session.CurrentState != StateRoot {
		t.Errorf("expected session to reset to root, got %s", session.CurrentState)
	}
	_ = err
}

func TestAdminSessionBackAndExit(t *testing.T) {
	cfg := &config.Config{
		AdminTriggerKeyword: "login admin",
		AdminPassword:       "pass123",
		IdleTimeoutMinutes:  30,
	}
	router := NewBotRouter(cfg, nil, nil)

	phone := "6289999999999"
	jid := types.NewJID(phone, types.DefaultUserServer)

	// 1. In StateAdminLoggedIn, typing 99 should keep state as StateAdminLoggedIn
	sess := router.sessions.GetOrCreate(jid, nil)
	sess.CurrentState = StateAdminLoggedIn

	router.Dispatch(context.Background(), &whatsapp.ParsedMessage{
		JID:  jid,
		Text: "99",
	})

	if sess.CurrentState != StateAdminLoggedIn {
		t.Errorf("expected state to remain StateAdminLoggedIn after typing 99, got %s", sess.CurrentState)
	}

	// 2. In StateAdminManagePakar, typing 99 should transition back to StateAdminLoggedIn
	sess.CurrentState = StateAdminManagePakar

	router.Dispatch(context.Background(), &whatsapp.ParsedMessage{
		JID:  jid,
		Text: "99",
	})

	if sess.CurrentState != StateAdminLoggedIn {
		t.Errorf("expected state to transition to StateAdminLoggedIn after typing 99 from ManagePakar, got %s", sess.CurrentState)
	}

	// 3. In StateAdminLoggedIn, typing 0 should exit admin mode to StateRoot
	router.Dispatch(context.Background(), &whatsapp.ParsedMessage{
		JID:  jid,
		Text: "0",
	})

	if sess.CurrentState != StateRoot {
		t.Errorf("expected state to reset to StateRoot after typing 0 in admin mode, got %s", sess.CurrentState)
	}
}
