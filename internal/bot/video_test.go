package bot

import (
	"context"
	"testing"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/whatsapp"

	"go.mau.fi/whatsmeow/types"
)

func TestVideoLiterasiNavigationAndDelivery(t *testing.T) {
	cfg := &config.Config{
		IdleTimeoutMinutes:  30,
		UserRateLimitPer10s: 50,
		UserCooldownSeconds: 60,
		TentangBPSVideoURL:  "https://www.youtube.com/watch?v=BFj6yi8HLAg",
	}

	sender := whatsapp.NewSender(nil, nil)
	router := NewBotRouter(cfg, sender, nil)

	userJID := types.NewJID("6283843607098", types.DefaultUserServer)
	session := router.sessions.GetOrCreate(userJID, nil)

	// 1. Open Menu 4 (Video Literasi)
	pMenu4 := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: "4",
	}
	router.Dispatch(context.Background(), pMenu4)
	if session.CurrentState != StateMenu4 {
		t.Fatalf("expected state %s, got %s", StateMenu4, session.CurrentState)
	}

	// 2. Select Option 1 (Sosial Kependudukan submenu)
	pSosduk := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: "1",
	}
	router.Dispatch(context.Background(), pSosduk)
	if session.CurrentState != StateMenu4_Sosduk {
		t.Fatalf("expected state %s, got %s", StateMenu4_Sosduk, session.CurrentState)
	}

	// 3. Select Option 1 in Sosduk (Garis Kemiskinan video)
	pVideoGK := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: "1",
	}
	router.Dispatch(context.Background(), pVideoGK)

	// 4. Test Back button (99) from Sosduk submenu
	pBack := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: "99",
	}
	router.Dispatch(context.Background(), pBack)
	if session.CurrentState != StateMenu4 {
		t.Fatalf("expected state %s after back, got %s", StateMenu4, session.CurrentState)
	}

	// 5. Select Option 3 (Ekonomi submenu)
	pEkonomi := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: "3",
	}
	router.Dispatch(context.Background(), pEkonomi)
	if session.CurrentState != StateMenu4_Ekonomi {
		t.Fatalf("expected state %s, got %s", StateMenu4_Ekonomi, session.CurrentState)
	}

	// 6. Select Option 3 in Ekonomi (Memaknai Data Inflasi video)
	pVideoInflasi := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: "3",
	}
	router.Dispatch(context.Background(), pVideoInflasi)

	// 7. Reset to Root and open Menu 11 (Tentang BPS)
	session.ResetToRoot()
	pMenu11 := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: "11",
	}
	router.Dispatch(context.Background(), pMenu11)
	if session.CurrentState != StateMenu11_About {
		t.Fatalf("expected state %s, got %s", StateMenu11_About, session.CurrentState)
	}
}
