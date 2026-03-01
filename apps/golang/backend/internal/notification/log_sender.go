package notification

import (
	"context"
	"log"
)

type logSender struct{}

func newLogSender() *logSender {
	return &logSender{}
}

func (s *logSender) Send(_ context.Context, msg *EmailMessage) error {
	log.Printf("notification [log] to=%s subject=%q body_len=%d", msg.To, msg.Subject, len(msg.HTML))
	return nil
}
