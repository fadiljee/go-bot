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
	var secList []singleSelectSection
	for _, sec := range sections {
		var rows []ListRow
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
	var nativeButtons []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton
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

// BuildBodyText formats text body with clean numbered list
func BuildBodyText(intro string, rows []ListRow, footer string) string {
	var sb strings.Builder

	if intro != "" {
		sb.WriteString(intro)
		sb.WriteString("\n\n")
	}

	for _, r := range rows {
		sb.WriteString(fmt.Sprintf("*%s.* %s\n", r.ID, r.Title))
		if r.Description != "" {
			sb.WriteString(fmt.Sprintf("   _%s_\n", r.Description))
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
		if header := inter.GetHeader(); header != nil && header.GetTitle() != "" {
			parts = append(parts, fmt.Sprintf("*%s*", strings.TrimSpace(header.GetTitle())))
		}
		if body := inter.GetBody(); body != nil && body.GetText() != "" {
			parts = append(parts, strings.TrimSpace(body.GetText()))
		}

		// If there are quick reply buttons, extract options if not already listed in body
		if nativeFlow := inter.GetNativeFlowMessage(); nativeFlow != nil {
			var buttonTexts []string
			bodyText := ""
			if inter.GetBody() != nil {
				bodyText = inter.GetBody().GetText()
			}
			for _, b := range nativeFlow.GetButtons() {
				if b.GetName() == "quick_reply" {
					var p quickReplyParams
					if err := json.Unmarshal([]byte(b.GetButtonParamsJSON()), &p); err == nil && p.DisplayText != "" {
						if !strings.Contains(bodyText, fmt.Sprintf("*%s*", p.ID)) && !strings.Contains(bodyText, fmt.Sprintf("(%s)", p.ID)) {
							buttonTexts = append(buttonTexts, fmt.Sprintf("• Ketik *%s* untuk %s", p.ID, p.DisplayText))
						}
					}
				}
			}
			if len(buttonTexts) > 0 {
				parts = append(parts, "*Pilihan Navigasi:*\n"+strings.Join(buttonTexts, "\n"))
			}
		}

		if footer := inter.GetFooter(); footer != nil && footer.GetText() != "" {
			parts = append(parts, fmt.Sprintf("_%s_", strings.TrimSpace(footer.GetText())))
		}
		return strings.Join(parts, "\n\n")
	}

	if list := unwrapped.GetListMessage(); list != nil {
		var parts []string
		if list.GetTitle() != "" {
			parts = append(parts, fmt.Sprintf("*%s*", strings.TrimSpace(list.GetTitle())))
		}
		if list.GetDescription() != "" {
			parts = append(parts, strings.TrimSpace(list.GetDescription()))
		}
		if list.GetFooterText() != "" {
			parts = append(parts, fmt.Sprintf("_%s_", strings.TrimSpace(list.GetFooterText())))
		}
		return strings.Join(parts, "\n\n")
	}

	if btn := unwrapped.GetButtonsMessage(); btn != nil {
		var parts []string
		if btn.GetContentText() != "" {
			parts = append(parts, strings.TrimSpace(btn.GetContentText()))
		}
		if btn.GetFooterText() != "" {
			parts = append(parts, fmt.Sprintf("_%s_", strings.TrimSpace(btn.GetFooterText())))
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
