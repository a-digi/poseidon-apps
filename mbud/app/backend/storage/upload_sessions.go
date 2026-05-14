package storage

import (
	"context"
	"database/sql"

	"mbud-plugin/model"
)

type UploadSessionRepo struct{ db *sql.DB }

func NewUploadSessionRepo(db *sql.DB) *UploadSessionRepo { return &UploadSessionRepo{db: db} }

func (r *UploadSessionRepo) Insert(ctx context.Context, s model.UploadSession) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO upload_sessions (id, invoice_id, status, created_at, expires_at)
		VALUES (?, NULLIF(?, ''), ?, ?, ?)`,
		s.ID, s.InvoiceID, s.Status, s.CreatedAt, s.ExpiresAt)
	return err
}

func (r *UploadSessionRepo) Get(ctx context.Context, id string) (model.UploadSession, bool, error) {
	var s model.UploadSession
	err := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(invoice_id, ''), status, created_at, expires_at
		FROM upload_sessions
		WHERE id = ?`, id).
		Scan(&s.ID, &s.InvoiceID, &s.Status, &s.CreatedAt, &s.ExpiresAt)
	if err == sql.ErrNoRows {
		return model.UploadSession{}, false, nil
	}
	if err != nil {
		return model.UploadSession{}, false, err
	}
	return s, true, nil
}

func (r *UploadSessionRepo) MarkConsumed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE upload_sessions SET status = 'consumed' WHERE id = ?`, id)
	return err
}

func (r *UploadSessionRepo) ExpireBefore(ctx context.Context, now int64) (int, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE upload_sessions SET status = 'expired'
		WHERE status = 'active' AND expires_at <= ?`, now)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}
