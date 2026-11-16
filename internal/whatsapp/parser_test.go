package whatsapp

import (
	"testing"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func TestParseIncomingMessage_UnsupportedMedia(t *testing.T) {
	baseInfo := types.MessageInfo{
		MessageSource: types.MessageSource{
			Chat:   types.NewJID("6281234567890", types.DefaultUserServer),
			Sender: types.NewJID("6281234567890", types.DefaultUserServer),
		},
	}

	// Voice note / AudioMessage should be flagged as unsupported media
	vnEvt := &events.Message{
		Info:    baseInfo,
		Message: &waE2E.Message{AudioMessage: &waE2E.AudioMessage{}},
	}
	parsedVN := ParseIncomingMessage(vnEvt)
	if parsedVN == nil || !parsedVN.IsUnsupportedMedia {
		t.Fatalf("expected AudioMessage to set IsUnsupportedMedia=true, got %+v", parsedVN)
	}
	if parsedVN.Text != "" {
		t.Fatalf("expected empty Text for AudioMessage, got %q", parsedVN.Text)
	}

	// ProtocolMessage should remain ignored (IsUnsupportedMedia=false)
	protoEvt := &events.Message{
		Info:    baseInfo,
		Message: &waE2E.Message{ProtocolMessage: &waE2E.ProtocolMessage{}},
	}
	parsedProto := ParseIncomingMessage(protoEvt)
	if parsedProto == nil || parsedProto.IsUnsupportedMedia {
		t.Fatalf("expected ProtocolMessage to keep IsUnsupportedMedia=false, got %+v", parsedProto)
	}
}
