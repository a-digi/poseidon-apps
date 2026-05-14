package storage

import (
	"context"
	"database/sql"

	"mbud-plugin/model"
)

type UpcomingRepo struct{ db *sql.DB }

func NewUpcomingRepo(db *sql.DB) *UpcomingRepo { return &UpcomingRepo{db: db} }

func (r *UpcomingRepo) List(ctx context.Context) ([]model.UpcomingInvoice, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(business_id, ''), amount, currency, description, due_at, created_at, updated_at
		FROM upcoming_invoices
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.UpcomingInvoice{}
	for rows.Next() {
		var u model.UpcomingInvoice
		if err := rows.Scan(&u.ID, &u.BusinessID, &u.Amount, &u.Currency, &u.Description, &u.DueAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UpcomingRepo) Get(ctx context.Context, id string) (model.UpcomingInvoice, bool, error) {
	var u model.UpcomingInvoice
	err := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(business_id, ''), amount, currency, description, due_at, created_at, updated_at
		FROM upcoming_invoices
		WHERE id = ?`, id).
		Scan(&u.ID, &u.BusinessID, &u.Amount, &u.Currency, &u.Description, &u.DueAt, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return model.UpcomingInvoice{}, false, nil
	}
	if err != nil {
		return model.UpcomingInvoice{}, false, err
	}
	return u, true, nil
}

func (r *UpcomingRepo) Insert(ctx context.Context, u model.UpcomingInvoice) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO upcoming_invoices (id, business_id, amount, currency, description, due_at, created_at, updated_at)
		VALUES (?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?)`,
		u.ID, u.BusinessID, u.Amount, u.Currency, u.Description, u.DueAt, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *UpcomingRepo) Update(ctx context.Context, u model.UpcomingInvoice) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE upcoming_invoices
		SET business_id = NULLIF(?, ''), amount = ?, currency = ?, description = ?, due_at = ?, updated_at = ?
		WHERE id = ?`,
		u.BusinessID, u.Amount, u.Currency, u.Description, u.DueAt, u.UpdatedAt, u.ID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *UpcomingRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM upcoming_invoices WHERE id = ?`, id)
	return err
}
