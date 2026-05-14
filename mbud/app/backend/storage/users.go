package storage

import (
	"context"
	"database/sql"

	"mbud-plugin/model"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, email, notes, created_at, updated_at
		FROM users
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Notes, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserRepo) Get(ctx context.Context, id string) (model.User, bool, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, notes, created_at, updated_at
		FROM users
		WHERE id = ?`, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.Notes, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return model.User{}, false, nil
	}
	if err != nil {
		return model.User{}, false, err
	}
	return u, true, nil
}

func (r *UserRepo) Insert(ctx context.Context, u model.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		u.ID, u.Name, u.Email, u.Notes, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *UserRepo) Update(ctx context.Context, u model.User) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET name = ?, email = ?, notes = ?, updated_at = ?
		WHERE id = ?`,
		u.Name, u.Email, u.Notes, u.UpdatedAt, u.ID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}
