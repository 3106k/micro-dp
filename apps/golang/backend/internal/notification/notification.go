package notification

import (
	"context"
	"log"
	"os"
)

// EmailMessage represents an email to be sent.
type EmailMessage struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

// EmailSender sends email messages.
type EmailSender interface {
	Send(ctx context.Context, msg *EmailMessage) error
}

// Config holds notification settings.
type Config struct {
	Provider    string // "sendgrid" or "log"
	SendGridKey string
	FromAddress string
	FromName    string
}

// LoadConfig reads notification configuration from environment variables.
func LoadConfig() Config {
	provider := os.Getenv("NOTIFICATION_PROVIDER")
	if provider == "" {
		provider = "log"
	}
	fromAddress := os.Getenv("NOTIFICATION_FROM_ADDRESS")
	if fromAddress == "" {
		fromAddress = "noreply@example.com"
	}
	fromName := os.Getenv("NOTIFICATION_FROM_NAME")
	if fromName == "" {
		fromName = "micro-dp"
	}
	return Config{
		Provider:    provider,
		SendGridKey: os.Getenv("SENDGRID_API_KEY"),
		FromAddress: fromAddress,
		FromName:    fromName,
	}
}

// NewEmailSender creates an EmailSender based on Config.Provider.
func NewEmailSender(cfg Config) EmailSender {
	if cfg.Provider == "sendgrid" {
		return newSendGridSender(cfg)
	}
	return newLogSender()
}

// LogStartup logs the notification configuration at startup.
func LogStartup(cfg Config) {
	log.Printf("notification initialized provider=%s from=%s", cfg.Provider, cfg.FromAddress)
}
