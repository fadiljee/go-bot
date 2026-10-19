package whatsapp

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type CachedMedia struct {
	URL           string
	DirectPath    string
	MediaKey      []byte
	FileSHA256    []byte
	FileEncSHA256 []byte
	FileLength    uint64
}

type Sender struct {
	client     *whatsmeow.Client
	mediaMu    sync.RWMutex
	mediaCache map[string]*CachedMedia
	httpClient *http.Client
}

func NewSender(client *whatsmeow.Client) *Sender {
	if client != nil && client.Store != nil && client.Store.ID != nil {
		SetBotNumber(client.Store.ID.User)
	}
	return &Sender{
		client:     client,
		mediaCache: make(map[string]*CachedMedia),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendTextMessage sends plain text message
func (s *Sender) SendTextMessage(ctx context.Context, jid types.JID, text string) error {
	log.Printf("📤 [SENDER] Mengirim pesan teks ke %s...", jid.String())
	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}
	resp, err := s.client.SendMessage(ctx, jid, msg)
	if err != nil {
		log.Printf("❌ [SENDER ERROR] Gagal mengirim pesan teks ke %s: %v", jid.String(), err)
		return err
	}
	log.Printf("✅ [SENDER OK] Pesan teks terkirim ke %s (ID: %s)", jid.String(), resp.ID)
	return nil
}

// SendInteractiveMessage sends single-select or quick-reply interactive message, with reliable plain-text delivery
func (s *Sender) SendInteractiveMessage(ctx context.Context, jid types.JID, msg *waE2E.Message) error {
	if s.client != nil && s.client.Store != nil && s.client.Store.ID != nil {
		SetBotNumber(s.client.Store.ID.User)
	}

	var textToSend string

	if msg != nil {
		var interactive *waE2E.InteractiveMessage
		if msg.GetViewOnceMessage() != nil && msg.GetViewOnceMessage().GetMessage() != nil {
			interactive = msg.GetViewOnceMessage().GetMessage().GetInteractiveMessage()
		} else {
			interactive = msg.GetInteractiveMessage()
		}

		if interactive != nil {
			if interactive.GetBody() != nil {
				textToSend = interactive.GetBody().GetText()
			}
		}
	}

	if textToSend == "" && msg.GetConversation() != "" {
		textToSend = msg.GetConversation()
	}

	if textToSend != "" {
		// Kirim teks terformat resmi yang 100% didukung dan dijamin tampil di semua WhatsApp client
		return s.SendTextMessage(ctx, jid, textToSend)
	}

	log.Printf("📤 [SENDER] Mengirim pesan ke %s...", jid.String())
	resp, err := s.client.SendMessage(ctx, jid, msg)
	if err != nil {
		log.Printf("❌ [SENDER ERROR] Gagal mengirim pesan ke %s: %v", jid.String(), err)
		return err
	}
	log.Printf("✅ [SENDER OK] Pesan terkirim ke %s (ID: %s)", jid.String(), resp.ID)
	return nil
}

// SendSticker sends a webp sticker from local file path with media upload caching
func (s *Sender) SendSticker(ctx context.Context, jid types.JID, filePath string) error {
	s.mediaMu.RLock()
	cached, found := s.mediaCache[filePath]
	s.mediaMu.RUnlock()

	if !found {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read sticker file %s: %w", filePath, err)
		}

		uploaded, err := s.client.Upload(ctx, data, whatsmeow.MediaImage)
		if err != nil {
			return fmt.Errorf("upload sticker: %w", err)
		}

		cached = &CachedMedia{
			URL:           uploaded.URL,
			DirectPath:    uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileSHA256:    uploaded.FileSHA256,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileLength:    uploaded.FileLength,
		}

		s.mediaMu.Lock()
		s.mediaCache[filePath] = cached
		s.mediaMu.Unlock()
	}

	msg := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           proto.String(cached.URL),
			DirectPath:    proto.String(cached.DirectPath),
			MediaKey:      cached.MediaKey,
			Mimetype:      proto.String("image/webp"),
			FileSHA256:    cached.FileSHA256,
			FileEncSHA256: cached.FileEncSHA256,
			FileLength:    proto.Uint64(cached.FileLength),
		},
	}

	resp, err := s.client.SendMessage(ctx, jid, msg)
	if err != nil {
		log.Printf("❌ [SENDER ERROR] Gagal mengirim stiker ke %s: %v", jid.String(), err)
		return err
	}
	log.Printf("✅ [SENDER OK] Stiker terkirim ke %s (ID: %s)", jid.String(), resp.ID)
	return nil
}

// SendImageFromURL downloads image from URL, uploads to WA, and sends with optional caption
func (s *Sender) SendImageFromURL(ctx context.Context, jid types.JID, imageURL, caption string) error {
	if imageURL == "" {
		return nil
	}

	// Check cache
	s.mediaMu.RLock()
	cached, found := s.mediaCache[imageURL]
	s.mediaMu.RUnlock()

	if !found {
		resp, err := s.httpClient.Get(imageURL)
		if err != nil {
			return fmt.Errorf("download image %s: %w", imageURL, err)
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read image body: %w", err)
		}

		mimeType := resp.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "image/jpeg"
		}

		uploaded, err := s.client.Upload(ctx, data, whatsmeow.MediaImage)
		if err != nil {
			return fmt.Errorf("upload image: %w", err)
		}

		cached = &CachedMedia{
			URL:           uploaded.URL,
			DirectPath:    uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileSHA256:    uploaded.FileSHA256,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileLength:    uploaded.FileLength,
		}

		s.mediaMu.Lock()
		s.mediaCache[imageURL] = cached
		s.mediaMu.Unlock()
	}

	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(cached.URL),
			DirectPath:    proto.String(cached.DirectPath),
			MediaKey:      cached.MediaKey,
			Mimetype:      proto.String("image/jpeg"),
			Caption:       proto.String(caption),
			FileSHA256:    cached.FileSHA256,
			FileEncSHA256: cached.FileEncSHA256,
			FileLength:    proto.Uint64(cached.FileLength),
		},
	}

	_, err := s.client.SendMessage(ctx, jid, msg)
	return err
}
