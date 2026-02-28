package db

import (
	"context"

	"github.com/user/micro-dp/domain"
)

type AdminAuditLogRepo struct {
	db DBTX
}

func NewAdminAuditLogRepo(db DBTX) *AdminAuditLogRepo {
	return &AdminAuditLogRepo{db: db}
}

func (r *AdminAuditLogRepo) Create(ctx context.Context, log *domain.AdminAuditLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO admin_audit_logs
		 (id, actor_user_id, action, target_type, target_id, metadata_json, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`,
		log.ID, log.ActorUserID, log.Action, log.TargetType, log.TargetID, log.MetadataJSON,
	)
	return err
}
