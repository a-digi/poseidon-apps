package storage

import (
	"context"
	"database/sql"

	"mbud-plugin/model"
)

type RecurringRepo struct{ db *sql.DB }

func NewRecurringRepo(db *sql.DB) *RecurringRepo { return &RecurringRepo{db: db} }

func (r *RecurringRepo) List(ctx context.Context) ([]model.RecurringInvoice, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(business_id, ''), amount, currency, description, frequency, start_at, end_at, active, issue_day_of_week, issue_day_of_month, issue_month_of_year, created_at, updated_at
		FROM recurring_invoices
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.RecurringInvoice{}
	for rows.Next() {
		var ri model.RecurringInvoice
		var freq string
		var active int
		if err := rows.Scan(&ri.ID, &ri.BusinessID, &ri.Amount, &ri.Currency, &ri.Description, &freq, &ri.StartAt, &ri.EndAt, &active, &ri.IssueDayOfWeek, &ri.IssueDayOfMonth, &ri.IssueMonthOfYear, &ri.CreatedAt, &ri.UpdatedAt); err != nil {
			return nil, err
		}
		ri.Frequency = model.Frequency(freq)
		ri.Active = active != 0
		out = append(out, ri)
	}
	return out, rows.Err()
}

func (r *RecurringRepo) Get(ctx context.Context, id string) (model.RecurringInvoice, bool, error) {
	var ri model.RecurringInvoice
	var freq string
	var active int
	err := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(business_id, ''), amount, currency, description, frequency, start_at, end_at, active, issue_day_of_week, issue_day_of_month, issue_month_of_year, created_at, updated_at
		FROM recurring_invoices
		WHERE id = ?`, id).
		Scan(&ri.ID, &ri.BusinessID, &ri.Amount, &ri.Currency, &ri.Description, &freq, &ri.StartAt, &ri.EndAt, &active, &ri.IssueDayOfWeek, &ri.IssueDayOfMonth, &ri.IssueMonthOfYear, &ri.CreatedAt, &ri.UpdatedAt)
	if err == sql.ErrNoRows {
		return model.RecurringInvoice{}, false, nil
	}
	if err != nil {
		return model.RecurringInvoice{}, false, err
	}
	ri.Frequency = model.Frequency(freq)
	ri.Active = active != 0
	return ri, true, nil
}

func (r *RecurringRepo) Insert(ctx context.Context, ri model.RecurringInvoice) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO recurring_invoices (id, business_id, amount, currency, description, frequency, start_at, end_at, active, issue_day_of_week, issue_day_of_month, issue_month_of_year, created_at, updated_at)
		VALUES (?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ri.ID, ri.BusinessID, ri.Amount, ri.Currency, ri.Description, string(ri.Frequency), ri.StartAt, ri.EndAt, boolToInt(ri.Active), ri.IssueDayOfWeek, ri.IssueDayOfMonth, ri.IssueMonthOfYear, ri.CreatedAt, ri.UpdatedAt)
	return err
}

func (r *RecurringRepo) Update(ctx context.Context, ri model.RecurringInvoice) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE recurring_invoices
		SET business_id = NULLIF(?, ''), amount = ?, currency = ?, description = ?, frequency = ?, start_at = ?, end_at = ?, active = ?, issue_day_of_week = ?, issue_day_of_month = ?, issue_month_of_year = ?, updated_at = ?
		WHERE id = ?`,
		ri.BusinessID, ri.Amount, ri.Currency, ri.Description, string(ri.Frequency), ri.StartAt, ri.EndAt, boolToInt(ri.Active), ri.IssueDayOfWeek, ri.IssueDayOfMonth, ri.IssueMonthOfYear, ri.UpdatedAt, ri.ID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RecurringRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM recurring_invoices WHERE id = ?`, id)
	return err
}
