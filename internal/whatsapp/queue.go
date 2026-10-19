package whatsapp

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type OutboundJob struct {
	JID          types.JID
	Message      *waE2E.Message
	Extra        *whatsmeow.SendRequestExtra
	SimulateType bool
	ResultChan   chan error
}

type MessageQueue struct {
	client           *whatsmeow.Client
	interactiveQueue chan OutboundJob
	blastQueue       chan OutboundJob
	minInteractiveMs int
	maxInteractiveMs int
	minBlastMs       int
	maxBlastMs       int
}

func NewMessageQueue(cli *whatsmeow.Client, minInter, maxInter, minBlast, maxBlast int) *MessageQueue {
	if minInter <= 0 {
		minInter = 600
	}
	if maxInter <= minInter {
		maxInter = 1200
	}
	if minBlast <= 0 {
		minBlast = 5000
	}
	if maxBlast <= minBlast {
		maxBlast = 10000
	}

	mq := &MessageQueue{
		client:           cli,
		interactiveQueue: make(chan OutboundJob, 500),
		blastQueue:       make(chan OutboundJob, 2000),
		minInteractiveMs: minInter,
		maxInteractiveMs: maxInter,
		minBlastMs:       minBlast,
		maxBlastMs:       maxBlast,
	}
	go mq.worker()
	return mq
}

// EnqueueInteractive queues interactive response with high priority
func (mq *MessageQueue) EnqueueInteractive(jid types.JID, msg *waE2E.Message, extra *whatsmeow.SendRequestExtra) {
	mq.interactiveQueue <- OutboundJob{
		JID:          jid,
		Message:      msg,
		Extra:        extra,
		SimulateType: true,
	}
}

// EnqueueBlast queues broadcast messages with low priority
func (mq *MessageQueue) EnqueueBlast(jid types.JID, msg *waE2E.Message, extra *whatsmeow.SendRequestExtra) {
	mq.blastQueue <- OutboundJob{
		JID:          jid,
		Message:      msg,
		Extra:        extra,
		SimulateType: false,
	}
}

// SendInteractiveSync queues interactive response and waits for delivery confirmation
func (mq *MessageQueue) SendInteractiveSync(ctx context.Context, jid types.JID, msg *waE2E.Message, extra *whatsmeow.SendRequestExtra) error {
	ch := make(chan error, 1)
	job := OutboundJob{
		JID:          jid,
		Message:      msg,
		Extra:        extra,
		SimulateType: true,
		ResultChan:   ch,
	}

	select {
	case mq.interactiveQueue <- job:
		select {
		case err := <-ch:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (mq *MessageQueue) worker() {
	for {
		select {
		case job := <-mq.interactiveQueue:
			mq.sendWithJitter(job, mq.minInteractiveMs, mq.maxInteractiveMs)
		default:
			select {
			case job := <-mq.interactiveQueue:
				mq.sendWithJitter(job, mq.minInteractiveMs, mq.maxInteractiveMs)
			case job := <-mq.blastQueue:
				mq.sendWithJitter(job, mq.minBlastMs, mq.maxBlastMs)
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
}

func (mq *MessageQueue) sendWithJitter(job OutboundJob, minMs, maxMs int) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if job.SimulateType && mq.client != nil {
		_ = mq.client.SendChatPresence(ctx, job.JID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	}

	diff := maxMs - minMs
	if diff <= 0 {
		diff = 100
	}
	delay := time.Duration(minMs+rand.Intn(diff)) * time.Millisecond
	time.Sleep(delay)

	if job.SimulateType && mq.client != nil {
		_ = mq.client.SendChatPresence(ctx, job.JID, types.ChatPresencePaused, types.ChatPresenceMediaText)
	}

	var err error
	if mq.client != nil {
		if job.Extra != nil {
			_, err = mq.client.SendMessage(ctx, job.JID, job.Message, *job.Extra)
		} else {
			_, err = mq.client.SendMessage(ctx, job.JID, job.Message)
		}
	}

	if job.ResultChan != nil {
		job.ResultChan <- err
	}

	if err != nil {
		slog.Error("Gagal mengirim pesan WA dari antrean", "jid", job.JID.String(), "error", err)
	}
}
