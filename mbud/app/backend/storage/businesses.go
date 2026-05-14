package storage

import (
	"context"
	"database/sql"

	"mbud-plugin/model"
)

type BusinessRepo struct{ db *sql.DB }

func NewBusinessRepo(db *sql.DB) *BusinessRepo { return &BusinessRepo{db: db} }

func (r *BusinessRepo) List(ctx context.Context) ([]model.Business, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, tax_id, email, address, notes, logo_type, created_at, updated_at
		FROM businesses
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Business{}
	for rows.Next() {
		var b model.Business
		if err := rows.Scan(&b.ID, &b.Name, &b.TaxID, &b.Email, &b.Address, &b.Notes, &b.LogoType, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BusinessRepo) Get(ctx context.Context, id string) (model.Business, bool, error) {
	var b model.Business
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, tax_id, email, address, notes, logo_type, created_at, updated_at
		FROM businesses
		WHERE id = ?`, id).
		Scan(&b.ID, &b.Name, &b.TaxID, &b.Email, &b.Address, &b.Notes, &b.LogoType, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return model.Business{}, false, nil
	}
	if err != nil {
		return model.Business{}, false, err
	}
	return b, true, nil
}

func (r *BusinessRepo) Insert(ctx context.Context, b model.Business) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO businesses (id, name, tax_id, email, address, notes, logo_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.Name, b.TaxID, b.Email, b.Address, b.Notes, b.LogoType, b.CreatedAt, b.UpdatedAt)
	return err
}

func (r *BusinessRepo) Update(ctx context.Context, b model.Business) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE businesses
		SET name = ?, tax_id = ?, email = ?, address = ?, notes = ?, updated_at = ?
		WHERE id = ?`,
		b.Name, b.TaxID, b.Email, b.Address, b.Notes, b.UpdatedAt, b.ID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *BusinessRepo) UpdateLogoType(ctx context.Context, id, logoType string, updatedAt int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE businesses
		SET logo_type = ?, updated_at = ?
		WHERE id = ?`,
		logoType, updatedAt, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *BusinessRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM businesses WHERE id = ?`, id)
	return err
}
