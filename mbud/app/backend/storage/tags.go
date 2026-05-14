package storage

import (
	"context"
	"database/sql"

	"mbud-plugin/model"
)

type TagRepo struct{ db *sql.DB }

func NewTagRepo(db *sql.DB) *TagRepo { return &TagRepo{db: db} }

func (r *TagRepo) List(ctx context.Context) ([]model.Tag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, created_at, updated_at
		FROM tags
		ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Tag{}
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TagRepo) Get(ctx context.Context, id string) (model.Tag, bool, error) {
	var t model.Tag
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, created_at, updated_at
		FROM tags
		WHERE id = ?`, id).
		Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return model.Tag{}, false, nil
	}
	if err != nil {
		return model.Tag{}, false, err
	}
	return t, true, nil
}

func (r *TagRepo) GetByName(ctx context.Context, name string) (model.Tag, bool, error) {
	var t model.Tag
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, created_at, updated_at
		FROM tags
		WHERE name = ?`, name).
		Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return model.Tag{}, false, nil
	}
	if err != nil {
		return model.Tag{}, false, err
	}
	return t, true, nil
}

func (r *TagRepo) Insert(ctx context.Context, t model.Tag) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tags (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)`,
		t.ID, t.Name, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *TagRepo) Update(ctx context.Context, t model.Tag) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tags
		SET name = ?, updated_at = ?
		WHERE id = ?`,
		t.Name, t.UpdatedAt, t.ID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *TagRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tags WHERE id = ?`, id)
	return err
}
