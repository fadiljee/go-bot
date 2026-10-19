package whatsapp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

var (
	botNumberMu sync.RWMutex
	botNumber   string
)

// SetBotNumber configures the current bot phone number for Click-to-Chat links
func SetBotNumber(number string) {
	botNumberMu.Lock()
	defer botNumberMu.Unlock()
	botNumber = number
}

// GetBotNumber returns the current bot phone number
func GetBotNumber() string {
	botNumberMu.RLock()
	defer botNumberMu.RUnlock()
	return botNumber
}

// GenerateClickToChatLink creates a wa.me direct action link for a command or number
func GenerateClickToChatLink(command string) string {
	num := GetBotNumber()
	if num == "" {
		return ""
	}
	return fmt.Sprintf("https://wa.me/%s?text=%s", num, url.QueryEscape(command))
}

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
						Name:             proto.String("single_select"),
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

// BuildQuickReplyButtons constructs a message with quick reply buttons and wa.me direct links
func BuildQuickReplyButtons(header, body, footer string, buttons []QuickReplyButton) *waE2E.Message {
	botNum := GetBotNumber()
	var bodyWithButtons strings.Builder
	bodyWithButtons.WriteString(body)

	if len(buttons) > 0 {
		bodyWithButtons.WriteString("\n\n📋 *Pilihan Cepat:*")
		for _, btn := range buttons {
			bodyWithButtons.WriteString(fmt.Sprintf("\n• *%s*", btn.DisplayText))
			if botNum != "" && btn.ID != "" {
				bodyWithButtons.WriteString(fmt.Sprintf("\n  👉 https://wa.me/%s?text=%s", botNum, url.QueryEscape(btn.ID)))
			}
		}
	}

	finalBody := bodyWithButtons.String()

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
			Text: proto.String(finalBody),
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

// FormatFooterWithLinks appends wa.me click-to-chat links for navigation lines in footer
func FormatFooterWithLinks(footer string) string {
	botNum := GetBotNumber()
	if botNum == "" || footer == "" {
		return footer
	}

	lines := strings.Split(footer, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, line)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "*0.") || strings.HasPrefix(trimmed, "*0 ") {
			result = append(result, fmt.Sprintf("👉 https://wa.me/%s?text=0", botNum))
		} else if strings.HasPrefix(trimmed, "*99.") || strings.HasPrefix(trimmed, "*99 ") {
			result = append(result, fmt.Sprintf("👉 https://wa.me/%s?text=99", botNum))
		} else if strings.HasPrefix(trimmed, "*") {
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) >= 1 {
				numPart := strings.Trim(parts[0], "*.:")
				if numPart != "" && len(numPart) <= 4 {
					allDigits := true
					for _, c := range numPart {
						if c < '0' || c > '9' {
							allDigits = false
							break
						}
					}
					if allDigits && numPart != "0" && numPart != "99" {
						result = append(result, fmt.Sprintf("👉 https://wa.me/%s?text=%s", botNum, url.QueryEscape(numPart)))
					}
				}
			}
		}
	}
	return strings.Join(result, "\n")
}

// BuildBodyText formats text body with numbered list and one-tap click-to-chat links
func BuildBodyText(intro string, rows []ListRow, footer string) string {
	var sb strings.Builder
	botNum := GetBotNumber()

	if intro != "" {
		sb.WriteString(intro)
		sb.WriteString("\n\n")
	}

	for _, r := range rows {
		sb.WriteString(fmt.Sprintf("*%s.* %s\n", r.ID, r.Title))
		if r.Description != "" {
			sb.WriteString(fmt.Sprintf("   _%s_\n", r.Description))
		}
		if botNum != "" && r.ID != "" {
			sb.WriteString(fmt.Sprintf("   👉 https://wa.me/%s?text=%s\n", botNum, url.QueryEscape(r.ID)))
		}
		sb.WriteString("\n")
	}

	if footer != "" {
		sb.WriteString(FormatFooterWithLinks(footer))
	}

	return strings.TrimSpace(sb.String())
}
