package storage

import (
	"context"
	"database/sql"

	"mbud-plugin/model"
)

type InvoiceAttachmentRepo struct{ db *sql.DB }

func NewInvoiceAttachmentRepo(db *sql.DB) *InvoiceAttachmentRepo {
	return &InvoiceAttachmentRepo{db: db}
}

func (r *InvoiceAttachmentRepo) Insert(ctx context.Context, a model.Attachment) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO invoice_attachments (id, invoice_id, session_id, mime, original_filename, size_bytes, created_at)
		VALUES (?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?)`,
		a.ID, a.InvoiceID, a.SessionID, a.Mime, a.OriginalFilename, a.SizeBytes, a.CreatedAt)
	return err
}

func (r *InvoiceAttachmentRepo) ListBySession(ctx context.Context, sessionID string) ([]model.Attachment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(invoice_id, ''), COALESCE(session_id, ''), mime, original_filename, size_bytes, created_at
		FROM invoice_attachments
		WHERE session_id = ?
		ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Attachment{}
	for rows.Next() {
		var a model.Attachment
		if err := rows.Scan(&a.ID, &a.InvoiceID, &a.SessionID, &a.Mime, &a.OriginalFilename, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *InvoiceAttachmentRepo) ListByInvoice(ctx context.Context, invoiceID string) ([]model.Attachment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(invoice_id, ''), COALESCE(session_id, ''), mime, original_filename, size_bytes, created_at
		FROM invoice_attachments
		WHERE invoice_id = ?
		ORDER BY created_at ASC`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Attachment{}
	for rows.Next() {
		var a model.Attachment
		if err := rows.Scan(&a.ID, &a.InvoiceID, &a.SessionID, &a.Mime, &a.OriginalFilename, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *InvoiceAttachmentRepo) AttachToInvoice(ctx context.Context, sessionID, invoiceID string) (int, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE invoice_attachments SET invoice_id = ?
		WHERE session_id = ? AND (invoice_id IS NULL OR invoice_id = '')`,
		invoiceID, sessionID)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *InvoiceAttachmentRepo) Get(ctx context.Context, id string) (model.Attachment, bool, error) {
	var a model.Attachment
	err := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(invoice_id, ''), COALESCE(session_id, ''), mime, original_filename, size_bytes, created_at
		FROM invoice_attachments
		WHERE id = ?`, id).
		Scan(&a.ID, &a.InvoiceID, &a.SessionID, &a.Mime, &a.OriginalFilename, &a.SizeBytes, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return model.Attachment{}, false, nil
	}
	if err != nil {
		return model.Attachment{}, false, err
	}
	return a, true, nil
}

func (r *InvoiceAttachmentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM invoice_attachments WHERE id = ?`, id)
	return err
}
