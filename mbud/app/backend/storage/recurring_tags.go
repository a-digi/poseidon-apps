package storage

import (
	"context"
	"database/sql"
	"strings"
)

type RecurringTagsRepo struct{ db *sql.DB }

func NewRecurringTagsRepo(db *sql.DB) *RecurringTagsRepo { return &RecurringTagsRepo{db: db} }

func (r *RecurringTagsRepo) ReplaceForRecurring(ctx context.Context, recurringID string, tagIDs []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM recurring_tags WHERE recurring_id = ?`, recurringID); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		if _, dup := seen[tagID]; dup {
			continue
		}
		seen[tagID] = struct{}{}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO recurring_tags (recurring_id, tag_id)
			VALUES (?, ?)`,
			recurringID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func (r *RecurringTagsRepo) TagIDsByRecurring(ctx context.Context, recurringID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tag_id
		FROM recurring_tags
		WHERE recurring_id = ?
		ORDER BY tag_id`, recurringID)
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

func (r *RecurringTagsRepo) TagIDsBatch(ctx context.Context, recurringIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(recurringIDs) == 0 {
		return out, nil
	}
	placeholders := strings.Repeat("?,", len(recurringIDs)-1) + "?"
	args := make([]interface{}, len(recurringIDs))
	for i, id := range recurringIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT recurring_id, tag_id
		FROM recurring_tags
		WHERE recurring_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var recID, tagID string
		if err := rows.Scan(&recID, &tagID); err != nil {
			return nil, err
		}
		out[recID] = append(out[recID], tagID)
	}
	return out, rows.Err()
}
