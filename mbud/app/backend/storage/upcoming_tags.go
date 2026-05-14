package storage

import (
	"context"
	"database/sql"
	"strings"
)

type UpcomingTagsRepo struct{ db *sql.DB }

func NewUpcomingTagsRepo(db *sql.DB) *UpcomingTagsRepo { return &UpcomingTagsRepo{db: db} }

func (r *UpcomingTagsRepo) ReplaceForUpcoming(ctx context.Context, upcomingID string, tagIDs []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM upcoming_tags WHERE upcoming_id = ?`, upcomingID); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		if _, dup := seen[tagID]; dup {
			continue
		}
		seen[tagID] = struct{}{}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO upcoming_tags (upcoming_id, tag_id)
			VALUES (?, ?)`,
			upcomingID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func (r *UpcomingTagsRepo) TagIDsByUpcoming(ctx context.Context, upcomingID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tag_id
		FROM upcoming_tags
		WHERE upcoming_id = ?
		ORDER BY tag_id`, upcomingID)
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

func (r *UpcomingTagsRepo) TagIDsBatch(ctx context.Context, upcomingIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(upcomingIDs) == 0 {
		return out, nil
	}
	placeholders := strings.Repeat("?,", len(upcomingIDs)-1) + "?"
	args := make([]interface{}, len(upcomingIDs))
	for i, id := range upcomingIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT upcoming_id, tag_id
		FROM upcoming_tags
		WHERE upcoming_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var upID, tagID string
		if err := rows.Scan(&upID, &tagID); err != nil {
			return nil, err
		}
		out[upID] = append(out[upID], tagID)
	}
	return out, rows.Err()
}
