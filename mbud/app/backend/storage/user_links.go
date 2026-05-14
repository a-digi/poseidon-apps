package storage

import (
	"context"
	"database/sql"
	"strings"
)

type UserLinksRepo struct{ db *sql.DB }

func NewUserLinksRepo(db *sql.DB) *UserLinksRepo { return &UserLinksRepo{db: db} }

func (r *UserLinksRepo) ReplaceForInvoice(ctx context.Context, invoiceID string, userIDs []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM invoice_users WHERE invoice_id = ?`, invoiceID); err != nil {
		return err
	}
	for _, userID := range userIDs {
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO invoice_users (invoice_id, user_id)
			VALUES (?, ?)`,
			invoiceID, userID); err != nil {
			return err
		}
	}
	return nil
}

func (r *UserLinksRepo) UserIDsByInvoice(ctx context.Context, invoiceID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id
		FROM invoice_users
		WHERE invoice_id = ?`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *UserLinksRepo) UserIDsBatch(ctx context.Context, invoiceIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(invoiceIDs) == 0 {
		return out, nil
	}
	placeholders := strings.Repeat("?,", len(invoiceIDs)-1) + "?"
	args := make([]interface{}, len(invoiceIDs))
	for i, id := range invoiceIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT invoice_id, user_id
		FROM invoice_users
		WHERE invoice_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var invID, userID string
		if err := rows.Scan(&invID, &userID); err != nil {
			return nil, err
		}
		out[invID] = append(out[invID], userID)
	}
	return out, rows.Err()
}

func (r *UserLinksRepo) DistinctUserIDs(ctx context.Context, from, to int64) ([]string, error) {
	clauses := []string{}
	args := []any{}
	if from > 0 {
		clauses = append(clauses, "i.issued_at >= ?")
		args = append(args, from)
	}
	if to > 0 {
		clauses = append(clauses, "i.issued_at <= ?")
		args = append(args, to)
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT iu.user_id
		FROM invoice_users iu
		JOIN invoices i ON i.id = iu.invoice_id`+where+`
		ORDER BY iu.user_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
