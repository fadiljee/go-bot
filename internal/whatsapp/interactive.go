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
