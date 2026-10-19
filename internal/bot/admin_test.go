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
		AdminPassword:         "supertim1903",
		AdminWhitelistNumbers: []string{"6285860125305"},
		AdminMaxRetry:         3,
		AdminLockoutMinutes:   60,
	}

	router := NewBotRouter(cfg, nil, nil)

	// 1. Test whitelist check
	if !router.isAdminWhitelisted("6285860125305") {
		t.Error("expected 6285860125305 to be whitelisted")
	}
	if !router.isAdminWhitelisted("+6285860125305") {
		t.Error("expected +6285860125305 to be whitelisted after normalization")
	}
	if !router.isAdminWhitelisted("085860125305") {
		t.Error("expected 085860125305 to be whitelisted after 08->62 conversion")
	}
	if router.isAdminWhitelisted("6281111111111") {
		t.Error("did not expect 6281111111111 to be whitelisted")
	}

	// 2. Test lockout logic
	phone := "6285860125305"
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
		Text: "supertim1903",
	})
	// State should reset to root
	if session.CurrentState != StateRoot {
		t.Errorf("expected session to reset to root, got %s", session.CurrentState)
	}
	_ = err
}
