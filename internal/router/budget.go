package router

import (
	"context"
	"database/sql"
	"time"
)

func (r *ModelRouter) needsApproval(ctx context.Context, estimate float64) bool {
	if estimate <= 0 || r.cfg.Budget.RequireApprovalAboveUSD <= 0 {
		return false
	}
	return estimate > r.cfg.Budget.RequireApprovalAboveUSD
}

func (r *ModelRouter) exceedsDailyBudget(ctx context.Context, estimate float64) bool {
	if estimate <= 0 || r.cfg.Budget.DailyUSDLimit <= 0 {
		return false
	}
	return TodayCost(ctx, r.db)+estimate > r.cfg.Budget.DailyUSDLimit
}

func TodayCost(ctx context.Context, db *sql.DB) float64 {
	if db == nil {
		return 0
	}
	start := time.Now().Local().Format("2006-01-02")
	var total sql.NullFloat64
	_ = db.QueryRowContext(ctx, `SELECT COALESCE(SUM(cost_usd), 0) FROM cost_events WHERE substr(created_at, 1, 10) = ?`, start).Scan(&total)
	if total.Valid {
		return total.Float64
	}
	return 0
}
