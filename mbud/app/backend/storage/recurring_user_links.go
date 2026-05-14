package storage

import (
	"context"
	"database/sql"
	"strings"
)

type RecurringUserLinksRepo struct{ db *sql.DB }

func NewRecurringUserLinksRepo(db *sql.DB) *RecurringUserLinksRepo {
	return &RecurringUserLinksRepo{db: db}
}

func (r *RecurringUserLinksRepo) ReplaceForRecurring(ctx context.Context, recurringID string, userIDs []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM recurring_users WHERE recurring_id = ?`, recurringID); err != nil {
		return err
	}
	for _, userID := range userIDs {
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO recurring_users (recurring_id, user_id)
			VALUES (?, ?)`,
			recurringID, userID); err != nil {
			return err
		}
	}
	return nil
}

func (r *RecurringUserLinksRepo) UserIDsByRecurring(ctx context.Context, recurringID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id
		FROM recurring_users
		WHERE recurring_id = ?
		ORDER BY user_id`, recurringID)
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

func (r *RecurringUserLinksRepo) UserIDsBatch(ctx context.Context, recurringIDs []string) (map[string][]string, error) {
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
		SELECT recurring_id, user_id
		FROM recurring_users
		WHERE recurring_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var recID, userID string
		if err := rows.Scan(&recID, &userID); err != nil {
			return nil, err
		}
		out[recID] = append(out[recID], userID)
	}
	return out, rows.Err()
}
