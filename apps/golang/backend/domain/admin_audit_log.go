package domain

import (
	"context"
	"time"
)

type AdminAuditLog struct {
	ID           string    `json:"id"`
	ActorUserID  string    `json:"actor_user_id"`
	Action       string    `json:"action"`
	TargetType   string    `json:"target_type"`
	TargetID     string    `json:"target_id"`
	MetadataJSON string    `json:"metadata_json"`
	CreatedAt    time.Time `json:"created_at"`
}

type AdminAuditLogRepository interface {
	Create(ctx context.Context, log *AdminAuditLog) error
}
