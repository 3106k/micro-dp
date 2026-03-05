package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/user/micro-dp/domain"
)

type DashboardWidgetRepo struct {
	db DBTX
}

func NewDashboardWidgetRepo(db DBTX) *DashboardWidgetRepo {
	return &DashboardWidgetRepo{db: db}
}

func (r *DashboardWidgetRepo) Create(ctx context.Context, w *domain.DashboardWidget) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO dashboard_widgets (id, dashboard_id, chart_id, position, created_at)
		 VALUES (?, ?, ?, ?, datetime('now'))`,
		w.ID, w.DashboardID, w.ChartID, w.Position,
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return domain.ErrWidgetDuplicate
	}
	return err
}

func (r *DashboardWidgetRepo) FindByID(ctx context.Context, dashboardID, id string) (*domain.DashboardWidget, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, dashboard_id, chart_id, position, created_at
		 FROM dashboard_widgets WHERE dashboard_id = ? AND id = ?`,
		dashboardID, id,
	)
	var w domain.DashboardWidget
	if err := row.Scan(&w.ID, &w.DashboardID, &w.ChartID, &w.Position, &w.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrWidgetNotFound
		}
		return nil, err
	}
	return &w, nil
}

func (r *DashboardWidgetRepo) ListByDashboard(ctx context.Context, dashboardID string) ([]domain.DashboardWidget, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, dashboard_id, chart_id, position, created_at
		 FROM dashboard_widgets WHERE dashboard_id = ? ORDER BY position`,
		dashboardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var widgets []domain.DashboardWidget
	for rows.Next() {
		var w domain.DashboardWidget
		if err := rows.Scan(&w.ID, &w.DashboardID, &w.ChartID, &w.Position, &w.CreatedAt); err != nil {
			return nil, err
		}
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

func (r *DashboardWidgetRepo) Delete(ctx context.Context, dashboardID, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM dashboard_widgets WHERE dashboard_id = ? AND id = ?`,
		dashboardID, id,
	)
	return err
}
