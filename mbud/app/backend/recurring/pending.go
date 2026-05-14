package recurring

import (
	"context"
	"database/sql"
	"sort"
	"strings"

	"mbud-plugin/model"
)

type Filters struct {
	From, To    int64
	BusinessIDs []string
	UserIDs     []string
	TagIDs      []string
}

const (
	maxPeriodsPerRule = 366
	maxTotalPending   = 500
)

func ProjectPending(ctx context.Context, db *sql.DB, now int64, filters Filters) ([]model.PendingInvoice, error) {
	out := []model.PendingInvoice{}

	rules, err := loadActiveRulesForProjection(ctx, db)
	if err != nil {
		return nil, err
	}
	ruleIDs := make([]string, len(rules))
	for i, r := range rules {
		ruleIDs[i] = r.ID
	}
	ruleTags, err := loadParentTagsByIDs(ctx, db, "recurring_tags", "recurring_id", ruleIDs)
	if err != nil {
		return nil, err
	}
	ruleUsers, err := loadParentUsersByIDs(ctx, db, "recurring_users", "recurring_id", ruleIDs)
	if err != nil {
		return nil, err
	}
	maxByRule, err := loadMaxPeriodByRule(ctx, db, ruleIDs)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if len(out) >= maxTotalPending {
			break
		}
		maxN, hasMax := maxByRule[rule.ID]
		startIdx := 0
		if hasMax {
			startIdx = maxN + 1
		}
		for n := startIdx; n-startIdx < maxPeriodsPerRule; n++ {
			due := NextDueAt(rule, n)
			if filters.To > 0 && due > filters.To {
				break
			}
			if rule.EndAt > 0 && due > rule.EndAt {
				break
			}
			if due <= now {
				continue
			}
			if filters.From > 0 && due < filters.From {
				continue
			}
			tags := ruleTags[rule.ID]
			users := ruleUsers[rule.ID]
			if !passesFilters(rule.BusinessID, users, tags, filters) {
				continue
			}
			out = append(out, model.PendingInvoice{
				Source:      "recurring",
				SourceID:    rule.ID,
				BusinessID:  rule.BusinessID,
				Amount:      rule.Amount,
				Currency:    rule.Currency,
				Description: rule.Description,
				DueAt:       due,
				IssuedAt:    due,
				TagIDs:      tags,
				UserIDs:     users,
			})
			if len(out) >= maxTotalPending {
				break
			}
		}
	}

	if len(out) < maxTotalPending {
		ups, err := loadFutureUnmaterialisedUpcomings(ctx, db, now)
		if err != nil {
			return nil, err
		}
		upIDs := make([]string, len(ups))
		for i, u := range ups {
			upIDs[i] = u.ID
		}
		upTags, err := loadParentTagsByIDs(ctx, db, "upcoming_tags", "upcoming_id", upIDs)
		if err != nil {
			return nil, err
		}
		upUsers, err := loadParentUsersByIDs(ctx, db, "upcoming_users", "upcoming_id", upIDs)
		if err != nil {
			return nil, err
		}
		for _, u := range ups {
			if len(out) >= maxTotalPending {
				break
			}
			if filters.To > 0 && u.DueAt > filters.To {
				continue
			}
			if filters.From > 0 && u.DueAt < filters.From {
				continue
			}
			tags := upTags[u.ID]
			users := upUsers[u.ID]
			if !passesFilters(u.BusinessID, users, tags, filters) {
				continue
			}
			out = append(out, model.PendingInvoice{
				Source:      "upcoming",
				SourceID:    u.ID,
				BusinessID:  u.BusinessID,
				Amount:      u.Amount,
				Currency:    u.Currency,
				Description: u.Description,
				DueAt:       u.DueAt,
				IssuedAt:    u.DueAt,
				TagIDs:      tags,
				UserIDs:     users,
			})
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].DueAt != out[j].DueAt {
			return out[i].DueAt < out[j].DueAt
		}
		return out[i].SourceID < out[j].SourceID
	})
	return out, nil
}

func loadActiveRulesForProjection(ctx context.Context, db *sql.DB) ([]model.RecurringInvoice, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, COALESCE(business_id, ''), amount, currency, description, frequency, start_at, end_at, issue_day_of_week, issue_day_of_month, issue_month_of_year
		 FROM recurring_invoices WHERE active = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.RecurringInvoice{}
	for rows.Next() {
		var ri model.RecurringInvoice
		var freq string
		if err := rows.Scan(&ri.ID, &ri.BusinessID, &ri.Amount, &ri.Currency, &ri.Description, &freq, &ri.StartAt, &ri.EndAt, &ri.IssueDayOfWeek, &ri.IssueDayOfMonth, &ri.IssueMonthOfYear); err != nil {
			return nil, err
		}
		ri.Frequency = model.Frequency(freq)
		ri.Active = true
		out = append(out, ri)
	}
	return out, rows.Err()
}

func loadFutureUnmaterialisedUpcomings(ctx context.Context, db *sql.DB, now int64) ([]model.UpcomingInvoice, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, COALESCE(business_id, ''), amount, currency, description, due_at
		 FROM upcoming_invoices
		 WHERE due_at > ? AND id NOT IN (SELECT upcoming_id FROM upcoming_invoice_links)`,
		now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.UpcomingInvoice{}
	for rows.Next() {
		var u model.UpcomingInvoice
		if err := rows.Scan(&u.ID, &u.BusinessID, &u.Amount, &u.Currency, &u.Description, &u.DueAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func loadParentTagsByIDs(ctx context.Context, db *sql.DB, table, parentCol string, parentIDs []string) (map[string][]string, error) {
	return loadParentMembershipByIDs(ctx, db, table, parentCol, "tag_id", parentIDs)
}

func loadParentUsersByIDs(ctx context.Context, db *sql.DB, table, parentCol string, parentIDs []string) (map[string][]string, error) {
	return loadParentMembershipByIDs(ctx, db, table, parentCol, "user_id", parentIDs)
}

func loadParentMembershipByIDs(ctx context.Context, db *sql.DB, table, parentCol, childCol string, parentIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(parentIDs) == 0 {
		return out, nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(parentIDs)), ",")
	args := make([]any, len(parentIDs))
	for i, id := range parentIDs {
		args[i] = id
	}
	rows, err := db.QueryContext(ctx,
		`SELECT `+parentCol+`, `+childCol+` FROM `+table+` WHERE `+parentCol+` IN (`+placeholders+`)`,
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var parent, child string
		if err := rows.Scan(&parent, &child); err != nil {
			return nil, err
		}
		out[parent] = append(out[parent], child)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for k := range out {
		sort.Strings(out[k])
	}
	return out, nil
}

func loadMaxPeriodByRule(ctx context.Context, db *sql.DB, ruleIDs []string) (map[string]int, error) {
	out := map[string]int{}
	if len(ruleIDs) == 0 {
		return out, nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(ruleIDs)), ",")
	args := make([]any, len(ruleIDs))
	for i, id := range ruleIDs {
		args[i] = id
	}
	rows, err := db.QueryContext(ctx,
		`SELECT recurring_id, MAX(period_index) FROM recurring_invoice_links WHERE recurring_id IN (`+placeholders+`) GROUP BY recurring_id`,
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ruleID string
		var maxN int
		if err := rows.Scan(&ruleID, &maxN); err != nil {
			return nil, err
		}
		out[ruleID] = maxN
	}
	return out, rows.Err()
}

func passesFilters(businessID string, userIDs, tagIDs []string, f Filters) bool {
	if len(f.BusinessIDs) > 0 {
		found := false
		for _, b := range f.BusinessIDs {
			if b == businessID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(f.UserIDs) > 0 && !hasOverlap(userIDs, f.UserIDs) {
		return false
	}
	if len(f.TagIDs) > 0 && !hasOverlap(tagIDs, f.TagIDs) {
		return false
	}
	return true
}

func hasOverlap(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}
