package storage

import (
	"context"
	"database/sql"
	"strings"
)

type UpcomingLinksRepo struct{ db *sql.DB }

func NewUpcomingLinksRepo(db *sql.DB) *UpcomingLinksRepo { return &UpcomingLinksRepo{db: db} }

func (r *UpcomingLinksRepo) Insert(ctx context.Context, invoiceID, upcomingID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO upcoming_invoice_links (invoice_id, upcoming_id)
		VALUES (?, ?)`,
		invoiceID, upcomingID)
	return err
}

func (r *UpcomingLinksRepo) InvoiceIDsByUpcoming(ctx context.Context, upcomingID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT invoice_id
		FROM upcoming_invoice_links
		WHERE upcoming_id = ?`, upcomingID)
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

func (r *UpcomingLinksRepo) DeleteByUpcoming(ctx context.Context, upcomingID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM upcoming_invoice_links WHERE upcoming_id = ?`, upcomingID)
	return err
}

func (r *UpcomingLinksRepo) UpcomingIDsByInvoice(ctx context.Context, invoiceID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT upcoming_id
		FROM upcoming_invoice_links
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

func (r *UpcomingLinksRepo) UpcomingIDsBatch(ctx context.Context, invoiceIDs []string) (map[string][]string, error) {
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
		SELECT invoice_id, upcoming_id
		FROM upcoming_invoice_links
		WHERE invoice_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var invID, upID string
		if err := rows.Scan(&invID, &upID); err != nil {
			return nil, err
		}
		out[invID] = append(out[invID], upID)
	}
	return out, rows.Err()
}
