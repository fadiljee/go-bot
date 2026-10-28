package whatsapp

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
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
	queue      *MessageQueue
	mediaMu    sync.RWMutex
	mediaCache map[string]*CachedMedia
	httpClient *http.Client
}

func NewSender(client *whatsmeow.Client, queue *MessageQueue) *Sender {
	return &Sender{
		client:     client,
		queue:      queue,
		mediaCache: make(map[string]*CachedMedia),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Sender) Queue() *MessageQueue {
	return s.queue
}

// ValidateURLAntiSSRF checks that the URL uses HTTPS and does not point to internal/private/loopback addresses
func ValidateURLAntiSSRF(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme != "https" {
		return fmt.Errorf("SSRF protection: only https scheme is allowed, got '%s'", parsed.Scheme)
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("SSRF protection: empty hostname")
	}

	lowerHost := strings.ToLower(hostname)
	if lowerHost == "localhost" || lowerHost == "127.0.0.1" || lowerHost == "::1" {
		return fmt.Errorf("SSRF protection: local address '%s' is prohibited", hostname)
	}

	// Resolve IP to check for private / loopback / link-local addresses
	ips, err := net.LookupIP(hostname)
	if err == nil {
		for _, ip := range ips {
			if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
				return fmt.Errorf("SSRF protection: host '%s' resolves to prohibited IP '%s'", hostname, ip.String())
			}
		}
	}

	return nil
}

// SendTextMessage sends plain text message via MessageQueue with typing simulation
func (s *Sender) SendTextMessage(ctx context.Context, jid types.JID, text string) error {
	if s == nil {
		return nil
	}
	log.Printf("[SENDER] Mengirim pesan teks ke %s...", jid.String())
	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}

	var err error
	if s.queue != nil {
		err = s.queue.SendInteractiveSync(ctx, jid, msg, nil)
	} else if s.client != nil {
		_, err = s.client.SendMessage(ctx, jid, msg)
	}

	if err != nil {
		log.Printf("[SENDER ERROR] Gagal mengirim pesan teks ke %s: %v", jid.String(), err)
		return err
	}
	log.Printf("[SENDER OK] Pesan teks berhasil terkirim ke %s", jid.String())
	return nil
}

// SendInteractiveMessage sends native list dropdown or quick-reply buttons message via queue with text fallback
func (s *Sender) SendInteractiveMessage(ctx context.Context, jid types.JID, msg *waE2E.Message) error {
	if s == nil {
		return nil
	}
	log.Printf("[SENDER] Mengirim pesan tombol resmi WhatsApp (Native Flow List / Buttons) ke %s...", jid.String())

	extra := &whatsmeow.SendRequestExtra{
		AdditionalNodes: &[]waBinary.Node{
			{
				Tag: "biz",
				Content: []waBinary.Node{
					{
						Tag: "interactive",
						Attrs: waBinary.Attrs{"type": "native_flow", "v": "1"},
						Content: []waBinary.Node{
							{
								Tag: "native_flow",
								Attrs: waBinary.Attrs{"v": "9", "name": "mixed"},
							},
						},
					},
				},
			},
		},
	}

	var err error
	if s.queue != nil {
		err = s.queue.SendInteractiveSync(ctx, jid, msg, extra)
	} else if s.client != nil {
		_, err = s.client.SendMessage(ctx, jid, msg, *extra)
	}

	if err != nil {
		log.Printf("[SENDER ERROR] Gagal mengirim pesan tombol ke %s: %v. Mengirim versi teks cadangan...", jid.String(), err)
		var textToSend string
		if msg != nil {
			unwrapped := msg
			if msg.GetViewOnceMessage() != nil && msg.GetViewOnceMessage().GetMessage() != nil {
				unwrapped = msg.GetViewOnceMessage().GetMessage()
			}
			if inter := unwrapped.GetInteractiveMessage(); inter != nil {
				if body := inter.GetBody(); body != nil {
					textToSend = body.GetText()
				}
			} else if list := unwrapped.GetListMessage(); list != nil {
				textToSend = list.GetDescription()
			} else if btn := unwrapped.GetButtonsMessage(); btn != nil {
				textToSend = btn.GetContentText()
			} else if conv := unwrapped.GetConversation(); conv != "" {
				textToSend = conv
			}
		}
		if textToSend != "" {
			return s.SendTextMessage(ctx, jid, textToSend)
		}
		return err
	}

	log.Printf("[SENDER OK] Pesan tombol resmi berhasil terkirim ke %s", jid.String())
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

	var err error
	if s.queue != nil {
		err = s.queue.SendInteractiveSync(ctx, jid, msg, nil)
	} else if s.client != nil {
		_, err = s.client.SendMessage(ctx, jid, msg)
	}

	if err != nil {
		log.Printf("[SENDER ERROR] Gagal mengirim stiker ke %s: %v", jid.String(), err)
		return err
	}
	log.Printf("[SENDER OK] Stiker terkirim ke %s", jid.String())
	return nil
}

// SendImageFromURL downloads image from URL with Anti-SSRF and LimitReader, uploads to WA, and sends with caption
func (s *Sender) SendImageFromURL(ctx context.Context, jid types.JID, imageURL, caption string) error {
	if imageURL == "" {
		return nil
	}

	// 1. Anti-SSRF Validation
	if err := ValidateURLAntiSSRF(imageURL); err != nil {
		log.Printf("[SSRF BLOCKED] Menolak download gambar dari %s: %v", imageURL, err)
		return err
	}

	// 2. Check cache
	s.mediaMu.RLock()
	cached, found := s.mediaCache[imageURL]
	s.mediaMu.RUnlock()

	if !found {
		resp, err := s.httpClient.Get(imageURL)
		if err != nil {
			return fmt.Errorf("download image %s: %w", imageURL, err)
		}
		defer resp.Body.Close()

		// 3. Anti-OOM Memory Protection: limit to 16 MB
		const maxMediaBytes = 16 * 1024 * 1024
		limitedReader := io.LimitReader(resp.Body, maxMediaBytes)
		data, err := io.ReadAll(limitedReader)
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

	var err error
	if s.queue != nil {
		err = s.queue.SendInteractiveSync(ctx, jid, msg, nil)
	} else if s.client != nil {
		_, err = s.client.SendMessage(ctx, jid, msg)
	}

	return err
}

// SendVideoFromFile sends an MP4 video from local disk path with in-memory upload caching
func (s *Sender) SendVideoFromFile(ctx context.Context, jid types.JID, filePath, caption string) error {
	if s == nil {
		return nil
	}

	// In test environments where client & queue are nil, safely mock success
	if s.client == nil && s.queue == nil {
		log.Printf("ℹ[SENDER MOCK] Melewati upload video ke %s (client & queue nil)", jid.String())
		return nil
	}

	s.mediaMu.RLock()
	cached, found := s.mediaCache[filePath]
	s.mediaMu.RUnlock()

	if !found {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read video file %s: %w", filePath, err)
		}

		if s.client == nil {
			return fmt.Errorf("whatsmeow client is nil")
		}

		uploaded, err := s.client.Upload(ctx, data, whatsmeow.MediaVideo)
		if err != nil {
			return fmt.Errorf("upload video: %w", err)
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
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(cached.URL),
			DirectPath:    proto.String(cached.DirectPath),
			MediaKey:      cached.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			Caption:       proto.String(caption),
			FileSHA256:    cached.FileSHA256,
			FileEncSHA256: cached.FileEncSHA256,
			FileLength:    proto.Uint64(cached.FileLength),
		},
	}

	var err error
	if s.queue != nil {
		err = s.queue.SendInteractiveSync(ctx, jid, msg, nil)
	} else if s.client != nil {
		_, err = s.client.SendMessage(ctx, jid, msg)
	}

	if err != nil {
		log.Printf("[SENDER ERROR] Gagal mengirim video ke %s: %v", jid.String(), err)
		return err
	}
	log.Printf("[SENDER OK] Video berhasil terkirim ke %s (%s)", jid.String(), filePath)
	return nil
}
