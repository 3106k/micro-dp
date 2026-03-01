package notification

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type sendGridSender struct {
	client      *sendgrid.Client
	fromAddress string
	fromName    string
}

func newSendGridSender(cfg Config) *sendGridSender {
	return &sendGridSender{
		client:      sendgrid.NewSendClient(cfg.SendGridKey),
		fromAddress: cfg.FromAddress,
		fromName:    cfg.FromName,
	}
}

func (s *sendGridSender) Send(ctx context.Context, msg *EmailMessage) error {
	from := mail.NewEmail(s.fromName, s.fromAddress)
	to := mail.NewEmail("", msg.To)
	m := mail.NewSingleEmail(from, msg.Subject, to, msg.Text, msg.HTML)

	var lastErr error
	for attempt := range 3 {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		resp, err := s.client.SendWithContext(ctx, m)
		if err != nil {
			lastErr = err
		} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		} else {
			lastErr = fmt.Errorf("sendgrid: status=%d body=%s", resp.StatusCode, resp.Body)
		}

		log.Printf("sendgrid attempt %d failed: %v", attempt+1, lastErr)
		if attempt < 2 {
			delay := time.Duration(attempt+1) * 500 * time.Millisecond
			time.Sleep(delay)
		}
	}
	return lastErr
}
