package storage

import (
	"context"
	"database/sql"
	"strings"
)

type InvoiceTagsRepo struct{ db *sql.DB }

func NewInvoiceTagsRepo(db *sql.DB) *InvoiceTagsRepo { return &InvoiceTagsRepo{db: db} }

func (r *InvoiceTagsRepo) ReplaceForInvoice(ctx context.Context, invoiceID string, tagIDs []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM invoice_tags WHERE invoice_id = ?`, invoiceID); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		if _, dup := seen[tagID]; dup {
			continue
		}
		seen[tagID] = struct{}{}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO invoice_tags (invoice_id, tag_id)
			VALUES (?, ?)`,
			invoiceID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func (r *InvoiceTagsRepo) TagIDsByInvoice(ctx context.Context, invoiceID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tag_id
		FROM invoice_tags
		WHERE invoice_id = ?
		ORDER BY tag_id`, invoiceID)
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

func (r *InvoiceTagsRepo) TagIDsBatch(ctx context.Context, invoiceIDs []string) (map[string][]string, error) {
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
		SELECT invoice_id, tag_id
		FROM invoice_tags
		WHERE invoice_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var invID, tagID string
		if err := rows.Scan(&invID, &tagID); err != nil {
			return nil, err
		}
		out[invID] = append(out[invID], tagID)
	}
	return out, rows.Err()
}

func (r *InvoiceTagsRepo) DistinctTagIDs(ctx context.Context, from, to int64) ([]string, error) {
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
		SELECT DISTINCT it.tag_id
		FROM invoice_tags it
		JOIN invoices i ON i.id = it.invoice_id`+where+`
		ORDER BY it.tag_id`, args...)
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
