package storage

import (
	"context"
	"database/sql"
	"strings"
)

type UpcomingUserLinksRepo struct{ db *sql.DB }

func NewUpcomingUserLinksRepo(db *sql.DB) *UpcomingUserLinksRepo {
	return &UpcomingUserLinksRepo{db: db}
}

func (r *UpcomingUserLinksRepo) ReplaceForUpcoming(ctx context.Context, upcomingID string, userIDs []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM upcoming_users WHERE upcoming_id = ?`, upcomingID); err != nil {
		return err
	}
	for _, userID := range userIDs {
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO upcoming_users (upcoming_id, user_id)
			VALUES (?, ?)`,
			upcomingID, userID); err != nil {
			return err
		}
	}
	return nil
}

func (r *UpcomingUserLinksRepo) UserIDsByUpcoming(ctx context.Context, upcomingID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id
		FROM upcoming_users
		WHERE upcoming_id = ?
		ORDER BY user_id`, upcomingID)
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

func (r *UpcomingUserLinksRepo) UserIDsBatch(ctx context.Context, upcomingIDs []string) (map[string][]string, error) {
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
		SELECT upcoming_id, user_id
		FROM upcoming_users
		WHERE upcoming_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var upID, userID string
		if err := rows.Scan(&upID, &userID); err != nil {
			return nil, err
		}
		out[upID] = append(out[upID], userID)
	}
	return out, rows.Err()
}
