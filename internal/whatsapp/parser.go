package whatsapp

import (
	"encoding/json"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type ParsedMessage struct {
	JID       types.JID
	SenderJID types.JID
	Name      string
	IsFromMe  bool
	IsGroup   bool
	Text      string
	Raw       *waE2E.Message
}

type nativeFlowParams struct {
	ID string `json:"id"`
}

// ParseIncomingMessage extracts text, IDs from buttons, or interactive single select lists
func ParseIncomingMessage(evt *events.Message) *ParsedMessage {
	if evt == nil || evt.Message == nil {
		return nil
	}

	info := evt.Info
	msg := evt.Message

	parsed := &ParsedMessage{
		JID:       info.Chat,
		SenderJID: info.Sender,
		Name:      info.PushName,
		IsFromMe:  info.IsFromMe,
		IsGroup:   info.IsGroup,
		Raw:       msg,
	}

	// Unwrap viewOnceMessage / ephemeralMessage if present
	if msg.GetViewOnceMessage() != nil && msg.GetViewOnceMessage().GetMessage() != nil {
		msg = msg.GetViewOnceMessage().GetMessage()
	}
	if msg.GetEphemeralMessage() != nil && msg.GetEphemeralMessage().GetMessage() != nil {
		msg = msg.GetEphemeralMessage().GetMessage()
	}

	// 1. Native Flow Interactive Response (Single Select List / Quick Reply)
	if inter := msg.GetInteractiveResponseMessage(); inter != nil {
		if nativeFlow := inter.GetNativeFlowResponseMessage(); nativeFlow != nil {
			var params nativeFlowParams
			if err := json.Unmarshal([]byte(nativeFlow.GetParamsJSON()), &params); err == nil && params.ID != "" {
				parsed.Text = strings.TrimSpace(params.ID)
				return parsed
			}
		}
		if body := inter.GetBody(); body != nil && body.GetText() != "" {
			parsed.Text = strings.TrimSpace(body.GetText())
			return parsed
		}
	}

	// 2. WhatsApp Template Button Reply
	if btn := msg.GetTemplateButtonReplyMessage(); btn != nil {
		if id := btn.GetSelectedID(); id != "" {
			parsed.Text = strings.TrimSpace(id)
			return parsed
		}
		if text := btn.GetSelectedDisplayText(); text != "" {
			parsed.Text = strings.TrimSpace(text)
			return parsed
		}
	}

	// 3. WhatsApp Buttons Response
	if btn := msg.GetButtonsResponseMessage(); btn != nil {
		if id := btn.GetSelectedButtonID(); id != "" {
			parsed.Text = strings.TrimSpace(id)
			return parsed
		}
		if text := btn.GetSelectedDisplayText(); text != "" {
			parsed.Text = strings.TrimSpace(text)
			return parsed
		}
	}

	// 4. WhatsApp Legacy List Response
	if list := msg.GetListResponseMessage(); list != nil {
		if row := list.GetSingleSelectReply(); row != nil && row.GetSelectedRowID() != "" {
			parsed.Text = strings.TrimSpace(row.GetSelectedRowID())
			return parsed
		}
		if title := list.GetTitle(); title != "" {
			parsed.Text = strings.TrimSpace(title)
			return parsed
		}
	}

	// 5. Standard Text Message
	if text := msg.GetConversation(); text != "" {
		parsed.Text = strings.TrimSpace(text)
		return parsed
	}
	if ext := msg.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
		parsed.Text = strings.TrimSpace(ext.GetText())
		return parsed
	}

	return parsed
}
