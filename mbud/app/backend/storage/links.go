package storage

import (
	"context"
	"database/sql"
	"strings"
)

type LinksRepo struct{ db *sql.DB }

func NewLinksRepo(db *sql.DB) *LinksRepo { return &LinksRepo{db: db} }

func (r *LinksRepo) Insert(ctx context.Context, invoiceID, recurringID string, periodIndex int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO recurring_invoice_links (invoice_id, recurring_id, period_index)
		VALUES (?, ?, ?)`,
		invoiceID, recurringID, periodIndex)
	return err
}

func (r *LinksRepo) MaxPeriodIndex(ctx context.Context, recurringID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(period_index), -1)
		FROM recurring_invoice_links
		WHERE recurring_id = ?`, recurringID).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r *LinksRepo) InvoiceIDsByRecurring(ctx context.Context, recurringID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT invoice_id
		FROM recurring_invoice_links
		WHERE recurring_id = ?`, recurringID)
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

func (r *LinksRepo) DeleteByRecurring(ctx context.Context, recurringID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM recurring_invoice_links WHERE recurring_id = ?`, recurringID)
	return err
}

func (r *LinksRepo) RecurringIDsByInvoice(ctx context.Context, invoiceID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT recurring_id
		FROM recurring_invoice_links
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

func (r *LinksRepo) RecurringIDsBatch(ctx context.Context, invoiceIDs []string) (map[string][]string, error) {
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
		SELECT invoice_id, recurring_id
		FROM recurring_invoice_links
		WHERE invoice_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var invID, recID string
		if err := rows.Scan(&invID, &recID); err != nil {
			return nil, err
		}
		out[invID] = append(out[invID], recID)
	}
	return out, rows.Err()
}
