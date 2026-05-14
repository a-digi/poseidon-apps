package recurring

import (
	"context"
	"database/sql"
	"strings"

	"mbud-plugin/model"
	"mbud-plugin/storage"
)

type Reconciler struct {
	db            *sql.DB
	recurring     *storage.RecurringRepo
	upcoming      *storage.UpcomingRepo
	invoices      *storage.InvoiceRepo
	links         *storage.LinksRepo
	upcomingLinks *storage.UpcomingLinksRepo
	newID         func() string
	now           func() int64
}

func New(
	db *sql.DB,
	rec *storage.RecurringRepo,
	upcoming *storage.UpcomingRepo,
	inv *storage.InvoiceRepo,
	links *storage.LinksRepo,
	upcomingLinks *storage.UpcomingLinksRepo,
	newID func() string,
	now func() int64,
) *Reconciler {
	return &Reconciler{
		db:            db,
		recurring:     rec,
		upcoming:      upcoming,
		invoices:      inv,
		links:         links,
		upcomingLinks: upcomingLinks,
		newID:         newID,
		now:           now,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, recurringID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rule, ok, err := loadRule(ctx, tx, recurringID)
	if err != nil {
		return err
	}
	if !ok {
		return tx.Commit()
	}
	if err := wipe(ctx, tx, recurringID); err != nil {
		return err
	}
	userIDs, err := readRecurringUserIDs(ctx, tx, recurringID)
	if err != nil {
		return err
	}
	tagIDs, err := readRecurringTagIDs(ctx, tx, recurringID)
	if err != nil {
		return err
	}
	if rule.Active {
		if err := r.emit(ctx, tx, rule, 0, userIDs, tagIDs); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Reconciler) WipeOnly(ctx context.Context, recurringID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := wipe(ctx, tx, recurringID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Reconciler) Catchup(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rules, err := loadActiveRules(ctx, tx)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		var maxN int
		if err := tx.QueryRowContext(ctx,
			`SELECT COALESCE(MAX(period_index), -1) FROM recurring_invoice_links WHERE recurring_id = ?`,
			rule.ID).Scan(&maxN); err != nil {
			return err
		}
		userIDs, err := readRecurringUserIDs(ctx, tx, rule.ID)
		if err != nil {
			return err
		}
		tagIDs, err := readRecurringTagIDs(ctx, tx, rule.ID)
		if err != nil {
			return err
		}
		if err := r.emit(ctx, tx, rule, maxN+1, userIDs, tagIDs); err != nil {
			return err
		}
	}

	ups, err := loadDueUpcomings(ctx, tx, r.now())
	if err != nil {
		return err
	}
	for _, u := range ups {
		userIDs, err := readUpcomingUserIDs(ctx, tx, u.ID)
		if err != nil {
			return err
		}
		tagIDs, err := readUpcomingTagIDs(ctx, tx, u.ID)
		if err != nil {
			return err
		}
		if err := r.emitUpcoming(ctx, tx, u, userIDs, tagIDs); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Reconciler) ReconcileUpcoming(ctx context.Context, upcomingID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	u, ok, err := loadUpcoming(ctx, tx, upcomingID)
	if err != nil {
		return err
	}
	if !ok {
		return tx.Commit()
	}
	if err := wipeUpcoming(ctx, tx, upcomingID); err != nil {
		return err
	}
	userIDs, err := readUpcomingUserIDs(ctx, tx, upcomingID)
	if err != nil {
		return err
	}
	tagIDs, err := readUpcomingTagIDs(ctx, tx, upcomingID)
	if err != nil {
		return err
	}
	if u.DueAt > 0 && u.DueAt <= r.now() {
		if err := r.emitUpcoming(ctx, tx, u, userIDs, tagIDs); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Reconciler) WipeOnlyUpcoming(ctx context.Context, upcomingID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := wipeUpcoming(ctx, tx, upcomingID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Reconciler) emit(ctx context.Context, tx *sql.Tx, rule model.RecurringInvoice, startIndex int, userIDs, tagIDs []string) error {
	nowSec := r.now()
	for n := startIndex; ; n++ {
		due := NextDueAt(rule, n)
		if due > nowSec {
			break
		}
		if rule.EndAt > 0 && due > rule.EndAt {
			break
		}
		invoiceID := r.newID()
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO invoices (id, business_id, amount, currency, description, issued_at, due_at, paid, paid_at, created_at, updated_at)
			 VALUES (?, NULLIF(?, ''), ?, ?, ?, ?, ?, 1, 0, ?, ?)`,
			invoiceID, rule.BusinessID, rule.Amount, rule.Currency, rule.Description,
			due, due, nowSec, nowSec); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO recurring_invoice_links (invoice_id, recurring_id, period_index) VALUES (?, ?, ?)`,
			invoiceID, rule.ID, n); err != nil {
			return err
		}
		if err := insertInvoiceUsers(ctx, tx, invoiceID, userIDs); err != nil {
			return err
		}
		if err := insertInvoiceTags(ctx, tx, invoiceID, tagIDs); err != nil {
			return err
		}
	}
	return nil
}

func loadRule(ctx context.Context, tx *sql.Tx, recurringID string) (model.RecurringInvoice, bool, error) {
	var ri model.RecurringInvoice
	var freq string
	var active int
	err := tx.QueryRowContext(ctx,
		`SELECT id, COALESCE(business_id, ''), amount, currency, description, frequency, start_at, end_at, active, issue_day_of_week, issue_day_of_month, issue_month_of_year
		 FROM recurring_invoices WHERE id = ?`, recurringID).
		Scan(&ri.ID, &ri.BusinessID, &ri.Amount, &ri.Currency, &ri.Description, &freq, &ri.StartAt, &ri.EndAt, &active, &ri.IssueDayOfWeek, &ri.IssueDayOfMonth, &ri.IssueMonthOfYear)
	if err == sql.ErrNoRows {
		return model.RecurringInvoice{}, false, nil
	}
	if err != nil {
		return model.RecurringInvoice{}, false, err
	}
	ri.Frequency = model.Frequency(freq)
	ri.Active = active != 0
	return ri, true, nil
}

func loadActiveRules(ctx context.Context, tx *sql.Tx) ([]model.RecurringInvoice, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, COALESCE(business_id, ''), amount, currency, description, frequency, start_at, end_at, active, issue_day_of_week, issue_day_of_month, issue_month_of_year
		 FROM recurring_invoices WHERE active = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.RecurringInvoice{}
	for rows.Next() {
		var ri model.RecurringInvoice
		var freq string
		var active int
		if err := rows.Scan(&ri.ID, &ri.BusinessID, &ri.Amount, &ri.Currency, &ri.Description, &freq, &ri.StartAt, &ri.EndAt, &active, &ri.IssueDayOfWeek, &ri.IssueDayOfMonth, &ri.IssueMonthOfYear); err != nil {
			return nil, err
		}
		ri.Frequency = model.Frequency(freq)
		ri.Active = active != 0
		out = append(out, ri)
	}
	return out, rows.Err()
}

func (r *Reconciler) emitUpcoming(ctx context.Context, tx *sql.Tx, u model.UpcomingInvoice, userIDs, tagIDs []string) error {
	nowSec := r.now()
	invoiceID := r.newID()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO invoices (id, business_id, amount, currency, description, issued_at, due_at, paid, paid_at, created_at, updated_at)
		 VALUES (?, NULLIF(?, ''), ?, ?, ?, ?, ?, 1, 0, ?, ?)`,
		invoiceID, u.BusinessID, u.Amount, u.Currency, u.Description, u.DueAt, u.DueAt, nowSec, nowSec); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO upcoming_invoice_links (invoice_id, upcoming_id) VALUES (?, ?)`,
		invoiceID, u.ID); err != nil {
		return err
	}
	if err := insertInvoiceUsers(ctx, tx, invoiceID, userIDs); err != nil {
		return err
	}
	if err := insertInvoiceTags(ctx, tx, invoiceID, tagIDs); err != nil {
		return err
	}
	return nil
}

func loadUpcoming(ctx context.Context, tx *sql.Tx, upcomingID string) (model.UpcomingInvoice, bool, error) {
	var u model.UpcomingInvoice
	err := tx.QueryRowContext(ctx,
		`SELECT id, COALESCE(business_id, ''), amount, currency, description, due_at
		 FROM upcoming_invoices WHERE id = ?`, upcomingID).
		Scan(&u.ID, &u.BusinessID, &u.Amount, &u.Currency, &u.Description, &u.DueAt)
	if err == sql.ErrNoRows {
		return model.UpcomingInvoice{}, false, nil
	}
	if err != nil {
		return model.UpcomingInvoice{}, false, err
	}
	return u, true, nil
}

func loadDueUpcomings(ctx context.Context, tx *sql.Tx, nowSec int64) ([]model.UpcomingInvoice, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, COALESCE(business_id, ''), amount, currency, description, due_at
		 FROM upcoming_invoices
		 WHERE due_at > 0 AND due_at <= ?
		   AND id NOT IN (SELECT upcoming_id FROM upcoming_invoice_links)`,
		nowSec)
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

func wipeUpcoming(ctx context.Context, tx *sql.Tx, upcomingID string) error {
	rows, err := tx.QueryContext(ctx,
		`SELECT invoice_id FROM upcoming_invoice_links WHERE upcoming_id = ?`, upcomingID)
	if err != nil {
		return err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(ids) > 0 {
		placeholders := strings.Repeat("?,", len(ids)-1) + "?"
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			args[i] = id
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM invoices WHERE id IN (`+placeholders+`)`, args...); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM upcoming_invoice_links WHERE upcoming_id = ?`, upcomingID); err != nil {
		return err
	}
	return nil
}

func wipe(ctx context.Context, tx *sql.Tx, recurringID string) error {
	rows, err := tx.QueryContext(ctx,
		`SELECT invoice_id FROM recurring_invoice_links WHERE recurring_id = ?`, recurringID)
	if err != nil {
		return err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(ids) > 0 {
		placeholders := strings.Repeat("?,", len(ids)-1) + "?"
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			args[i] = id
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM invoices WHERE id IN (`+placeholders+`)`, args...); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM recurring_invoice_links WHERE recurring_id = ?`, recurringID); err != nil {
		return err
	}
	return nil
}

func readRecurringUserIDs(ctx context.Context, tx *sql.Tx, recurringID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT user_id FROM recurring_users WHERE recurring_id = ? ORDER BY user_id`, recurringID)
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

func readUpcomingUserIDs(ctx context.Context, tx *sql.Tx, upcomingID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT user_id FROM upcoming_users WHERE upcoming_id = ? ORDER BY user_id`, upcomingID)
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

func insertInvoiceUsers(ctx context.Context, tx *sql.Tx, invoiceID string, userIDs []string) error {
	for _, uid := range userIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO invoice_users (invoice_id, user_id) VALUES (?, ?)`,
			invoiceID, uid); err != nil {
			return err
		}
	}
	return nil
}

func readRecurringTagIDs(ctx context.Context, tx *sql.Tx, recurringID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT tag_id FROM recurring_tags WHERE recurring_id = ? ORDER BY tag_id`, recurringID)
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

func readUpcomingTagIDs(ctx context.Context, tx *sql.Tx, upcomingID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT tag_id FROM upcoming_tags WHERE upcoming_id = ? ORDER BY tag_id`, upcomingID)
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

func insertInvoiceTags(ctx context.Context, tx *sql.Tx, invoiceID string, tagIDs []string) error {
	for _, tid := range tagIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO invoice_tags (invoice_id, tag_id) VALUES (?, ?)`,
			invoiceID, tid); err != nil {
			return err
		}
	}
	return nil
}
