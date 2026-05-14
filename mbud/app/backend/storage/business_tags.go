package storage

import (
	"context"
	"database/sql"
	"strings"
)

type BusinessTagsRepo struct{ db *sql.DB }

func NewBusinessTagsRepo(db *sql.DB) *BusinessTagsRepo { return &BusinessTagsRepo{db: db} }

func (r *BusinessTagsRepo) ReplaceForBusiness(ctx context.Context, businessID string, tagIDs []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM business_tags WHERE business_id = ?`, businessID); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		if _, dup := seen[tagID]; dup {
			continue
		}
		seen[tagID] = struct{}{}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO business_tags (business_id, tag_id)
			VALUES (?, ?)`,
			businessID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func (r *BusinessTagsRepo) TagIDsByBusiness(ctx context.Context, businessID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tag_id
		FROM business_tags
		WHERE business_id = ?
		ORDER BY tag_id`, businessID)
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

func (r *BusinessTagsRepo) TagIDsBatch(ctx context.Context, businessIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(businessIDs) == 0 {
		return out, nil
	}
	placeholders := strings.Repeat("?,", len(businessIDs)-1) + "?"
	args := make([]interface{}, len(businessIDs))
	for i, id := range businessIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT business_id, tag_id
		FROM business_tags
		WHERE business_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bizID, tagID string
		if err := rows.Scan(&bizID, &tagID); err != nil {
			return nil, err
		}
		out[bizID] = append(out[bizID], tagID)
	}
	return out, rows.Err()
}
