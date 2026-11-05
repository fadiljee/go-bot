package whatsapp

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

type ListRow struct {
	ID          string `json:"id"`
	RowID       string `json:"row_id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ListSection struct {
	Title string    `json:"title"`
	Rows  []ListRow `json:"rows"`
}

type QuickReplyButton struct {
	ID          string `json:"id"`
	DisplayText string `json:"display_text"`
}

type singleSelectSection struct {
	Title string    `json:"title"`
	Rows  []ListRow `json:"rows"`
}

type singleSelectParams struct {
	Title    string                `json:"title"`
	Sections []singleSelectSection `json:"sections"`
}

type quickReplyParams struct {
	DisplayText string `json:"display_text"`
	ID          string `json:"id"`
}

// BuildInteractiveList constructs a native WhatsApp List Message with Dropdown Drawer (single_select)
func BuildInteractiveList(header, body, footer, buttonText string, sections []ListSection) *waE2E.Message {
	secList := make([]singleSelectSection, 0, len(sections))
	for _, sec := range sections {
		rows := make([]ListRow, 0, len(sec.Rows))
		for _, r := range sec.Rows {
			row := r
			if row.RowID == "" {
				row.RowID = row.ID
			}
			rows = append(rows, row)
		}
		secList = append(secList, singleSelectSection{
			Title: sec.Title,
			Rows:  rows,
		})
	}

	payload := singleSelectParams{
		Title:    buttonText,
		Sections: secList,
	}

	jsonBytes, _ := json.Marshal(payload)

	btn := &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		Name:             proto.String("single_select"),
		ButtonParamsJSON: proto.String(string(jsonBytes)),
	}

	msgVersion := int32(1)
	interactiveMsg := &waE2E.InteractiveMessage{
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(body),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String(footer),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				MessageVersion: &msgVersion,
				Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					btn,
				},
			},
		},
	}

	if header != "" {
		interactiveMsg.Header = &waE2E.InteractiveMessage_Header{
			Title: proto.String(header),
		}
	}

	return &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				MessageContextInfo: &waE2E.MessageContextInfo{
					DeviceListMetadataVersion: proto.Int32(2),
				},
				InteractiveMessage: interactiveMsg,
			},
		},
	}
}

// BuildQuickReplyButtons constructs a native WhatsApp Quick Reply Buttons Message (vertical stacked full-width)
func BuildQuickReplyButtons(header, body, footer string, buttons []QuickReplyButton) *waE2E.Message {
	nativeButtons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(buttons))
	for _, b := range buttons {
		p := quickReplyParams{
			DisplayText: b.DisplayText,
			ID:          b.ID,
		}
		jsonBytes, _ := json.Marshal(p)
		nativeButtons = append(nativeButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("quick_reply"),
			ButtonParamsJSON: proto.String(string(jsonBytes)),
		})
	}

	msgVersion := int32(1)
	interactiveMsg := &waE2E.InteractiveMessage{
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(body),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String(footer),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				MessageVersion: &msgVersion,
				Buttons:        nativeButtons,
			},
		},
	}

	if header != "" {
		interactiveMsg.Header = &waE2E.InteractiveMessage_Header{
			Title: proto.String(header),
		}
	}

	return &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				MessageContextInfo: &waE2E.MessageContextInfo{
					DeviceListMetadataVersion: proto.Int32(2),
				},
				InteractiveMessage: interactiveMsg,
			},
		},
	}
}

// UnpackInteractiveMessage extracts the inner InteractiveMessage regardless of ViewOnce packaging
func UnpackInteractiveMessage(msg *waE2E.Message) *waE2E.InteractiveMessage {
	if msg == nil {
		return nil
	}
	if msg.GetInteractiveMessage() != nil {
		return msg.GetInteractiveMessage()
	}
	if msg.GetViewOnceMessage() != nil && msg.GetViewOnceMessage().GetMessage() != nil {
		return msg.GetViewOnceMessage().GetMessage().GetInteractiveMessage()
	}
	if msg.GetViewOnceMessageV2() != nil && msg.GetViewOnceMessageV2().GetMessage() != nil {
		return msg.GetViewOnceMessageV2().GetMessage().GetInteractiveMessage()
	}
	return nil
}

// NumberToEmoji formats numeric IDs to clean numbered format without emojis
func NumberToEmoji(id string) string {
	return fmt.Sprintf("%s.", id)
}

func isRedundantDesc(title, desc string) bool {
	t := strings.ToLower(strings.TrimSpace(title))
	d := strings.ToLower(strings.TrimSpace(desc))
	if t == d || strings.Contains(t, d) || strings.Contains(d, t) {
		return true
	}
	return false
}

// BuildBodyText formats text body with clean numbered list
func BuildBodyText(intro string, rows []ListRow, footer string) string {
	var sb strings.Builder

	if intro != "" {
		sb.WriteString(intro)
		sb.WriteString("\n\n")
	}

	for _, r := range rows {
		fmt.Fprintf(&sb, "%s  %s\n", NumberToEmoji(r.ID), strings.TrimSpace(r.Title))
		desc := strings.TrimSpace(r.Description)
		if desc != "" && !isRedundantDesc(r.Title, desc) {
			fmt.Fprintf(&sb, "   • %s\n", desc)
		}
	}

	if footer != "" {
		sb.WriteString("\n")
		sb.WriteString(footer)
	}

	return strings.TrimSpace(sb.String())
}

// ExtractTextFromMessage converts any interactive message (List, Buttons, or Native Flow)
// into clean, fully-readable plain text compatible with 100% of WhatsApp clients (including iOS iPhone).
func ExtractTextFromMessage(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}

	unwrapped := msg
	if msg.GetViewOnceMessage() != nil && msg.GetViewOnceMessage().GetMessage() != nil {
		unwrapped = msg.GetViewOnceMessage().GetMessage()
	} else if msg.GetViewOnceMessageV2() != nil && msg.GetViewOnceMessageV2().GetMessage() != nil {
		unwrapped = msg.GetViewOnceMessageV2().GetMessage()
	}

	if inter := unwrapped.GetInteractiveMessage(); inter != nil {
		var parts []string
		bodyText := ""
		if inter.GetBody() != nil {
			bodyText = strings.TrimSpace(inter.GetBody().GetText())
		}

		if header := inter.GetHeader(); header != nil && header.GetTitle() != "" {
			hTitle := strings.TrimSpace(header.GetTitle())
			if !strings.Contains(strings.ToLower(bodyText), strings.ToLower(hTitle)) {
				parts = append(parts, fmt.Sprintf("*%s*", hTitle))
			}
		}

		if bodyText != "" {
			parts = append(parts, bodyText)
		}

		// If there are quick reply buttons, extract options if not already listed in body
		if nativeFlow := inter.GetNativeFlowMessage(); nativeFlow != nil {
			var buttonTexts []string
			for _, b := range nativeFlow.GetButtons() {
				if b.GetName() == "quick_reply" {
					var p quickReplyParams
					if err := json.Unmarshal([]byte(b.GetButtonParamsJSON()), &p); err == nil && p.DisplayText != "" {
						if !strings.Contains(bodyText, fmt.Sprintf("*%s*", p.ID)) && !strings.Contains(bodyText, fmt.Sprintf("(%s)", p.ID)) {
							dText := strings.TrimSpace(p.DisplayText)
							dText = strings.TrimSuffix(dText, fmt.Sprintf("(%s)", p.ID))
							dText = strings.TrimSpace(dText)
							buttonTexts = append(buttonTexts, fmt.Sprintf("• Ketik *%s* untuk %s", p.ID, dText))
						}
					}
				}
			}
			if len(buttonTexts) > 0 {
				parts = append(parts, "*Pilihan Navigasi:*\n"+strings.Join(buttonTexts, "\n"))
			}
		}

		if footer := inter.GetFooter(); footer != nil && footer.GetText() != "" {
			fText := strings.TrimSpace(footer.GetText())
			if !strings.Contains(bodyText, fText) && !strings.Contains(bodyText, "Statistik Bermakna") {
				if fText == "PST BPS Kabupaten Bangka Barat" {
					parts = append(parts, "```Statistik Bermakna, Statistik Berdampak```")
				} else {
					parts = append(parts, fText)
				}
			}
		}
		return strings.Join(parts, "\n\n")
	}

	if list := unwrapped.GetListMessage(); list != nil {
		var parts []string
		bodyText := ""
		if list.GetDescription() != "" {
			bodyText = strings.TrimSpace(list.GetDescription())
		}
		if list.GetTitle() != "" {
			tTitle := strings.TrimSpace(list.GetTitle())
			if !strings.Contains(strings.ToLower(bodyText), strings.ToLower(tTitle)) {
				parts = append(parts, fmt.Sprintf("*%s*", tTitle))
			}
		}
		if bodyText != "" {
			parts = append(parts, bodyText)
		}
		if list.GetFooterText() != "" {
			fText := strings.TrimSpace(list.GetFooterText())
			if !strings.Contains(bodyText, fText) {
				parts = append(parts, fText)
			}
		}
		return strings.Join(parts, "\n\n")
	}

	if btn := unwrapped.GetButtonsMessage(); btn != nil {
		var parts []string
		bodyText := ""
		if btn.GetContentText() != "" {
			bodyText = strings.TrimSpace(btn.GetContentText())
			parts = append(parts, bodyText)
		}
		if btn.GetFooterText() != "" {
			fText := strings.TrimSpace(btn.GetFooterText())
			if !strings.Contains(bodyText, fText) {
				parts = append(parts, fText)
			}
		}
		return strings.Join(parts, "\n\n")
	}

	if conv := unwrapped.GetConversation(); conv != "" {
		return conv
	}
	if ext := unwrapped.GetExtendedTextMessage(); ext != nil {
		return ext.GetText()
	}

	return ""
}
