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
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ListSection struct {
	Title string    `json:"title"`
	Rows  []ListRow `json:"rows"`
}

type QuickReplyButton struct {
	ID          string
	DisplayText string
}

type singleSelectParams struct {
	Title    string        `json:"title"`
	Sections []ListSection `json:"sections"`
}

type quickReplyParams struct {
	DisplayText string `json:"display_text"`
	ID          string `json:"id"`
}

// BuildInteractiveList constructs a single-bubble interactive message with Single Select List
func BuildInteractiveList(header, body, footer, buttonText string, sections []ListSection) *waE2E.Message {
	paramJSON, _ := json.Marshal(singleSelectParams{
		Title:    buttonText,
		Sections: sections,
	})

	interactiveMsg := &waE2E.InteractiveMessage{
		Header: &waE2E.InteractiveMessage_Header{
			Title:              proto.String(header),
			HasMediaAttachment: proto.Bool(false),
		},
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(body),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String(footer),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					{
						Name: proto.String("single_select"),
						ButtonParamsJSON: proto.String(string(paramJSON)),
					},
				},
			},
		},
	}

	return &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				MessageContextInfo: &waE2E.MessageContextInfo{
					DeviceListMetadata:        &waE2E.DeviceListMetadata{},
					DeviceListMetadataVersion: proto.Int32(2),
				},
				InteractiveMessage: interactiveMsg,
			},
		},
	}
}

// BuildQuickReplyButtons constructs a message with quick reply buttons at bottom
func BuildQuickReplyButtons(header, body, footer string, buttons []QuickReplyButton) *waE2E.Message {
	var nativeButtons []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton
	for _, btn := range buttons {
		pJSON, _ := json.Marshal(quickReplyParams{
			DisplayText: btn.DisplayText,
			ID:          btn.ID,
		})
		nativeButtons = append(nativeButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("quick_reply"),
			ButtonParamsJSON: proto.String(string(pJSON)),
		})
	}

	var headerObj *waE2E.InteractiveMessage_Header
	if header != "" {
		headerObj = &waE2E.InteractiveMessage_Header{
			Title:              proto.String(header),
			HasMediaAttachment: proto.Bool(false),
		}
	}

	interactiveMsg := &waE2E.InteractiveMessage{
		Header: headerObj,
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(body),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String(footer),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: nativeButtons,
			},
		},
	}

	return &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				MessageContextInfo: &waE2E.MessageContextInfo{
					DeviceListMetadata:        &waE2E.DeviceListMetadata{},
					DeviceListMetadataVersion: proto.Int32(2),
				},
				InteractiveMessage: interactiveMsg,
			},
		},
	}
}

// BuildBodyText formats text body with numbered list fallback
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

	return sb.String()
}
