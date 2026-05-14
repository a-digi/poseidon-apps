package storage

import (
	"context"
	"database/sql"
	"sort"
	"strings"

	"mbud-plugin/model"
)

var invoiceSortColumns = map[string]string{
	"issuedAt": "issued_at",
	"dueAt":    "due_at",
	"amount":   "amount",
}

type InvoiceRepo struct{ db *sql.DB }

func NewInvoiceRepo(db *sql.DB) *InvoiceRepo { return &InvoiceRepo{db: db} }

// SQLite stores booleans as 0/1 integers; cast on read/write boundaries.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func invoiceFilterWhere(from, to int64, businessIDs, userIDs, tagIDs []string, unpaidOnly bool) (string, []any) {
	clauses := []string{}
	args := []any{}
	if from > 0 {
		clauses = append(clauses, "issued_at >= ?")
		args = append(args, from)
	}
	if to > 0 {
		clauses = append(clauses, "issued_at <= ?")
		args = append(args, to)
	}
	if len(businessIDs) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(businessIDs)), ",")
		clauses = append(clauses, "business_id IN ("+placeholders+")")
		for _, id := range businessIDs {
			args = append(args, id)
		}
	}
	if len(userIDs) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(userIDs)), ",")
		clauses = append(clauses,
			"EXISTS (SELECT 1 FROM invoice_users WHERE invoice_users.invoice_id = invoices.id AND invoice_users.user_id IN ("+placeholders+"))")
		for _, id := range userIDs {
			args = append(args, id)
		}
	}
	if len(tagIDs) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(tagIDs)), ",")
		clauses = append(clauses,
			"EXISTS (SELECT 1 FROM invoice_tags WHERE invoice_tags.invoice_id = invoices.id AND invoice_tags.tag_id IN ("+placeholders+"))")
		for _, id := range tagIDs {
			args = append(args, id)
		}
	}
	if unpaidOnly {
		clauses = append(clauses, "paid = 0")
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *InvoiceRepo) List(ctx context.Context, from, to int64, businessIDs, userIDs, tagIDs []string, unpaidOnly bool, limit, offset int64, sortBy, sortDir string) ([]model.Invoice, error) {
	col, ok := invoiceSortColumns[sortBy]
	if !ok {
		col = "issued_at"
	}
	dir := "DESC"
	if sortDir == "asc" {
		dir = "ASC"
	}
	where, args := invoiceFilterWhere(from, to, businessIDs, userIDs, tagIDs, unpaidOnly)
	query := `SELECT id, COALESCE(business_id, ''), amount, currency, description, issued_at, due_at, paid, paid_at, created_at, updated_at FROM invoices` + where + ` ORDER BY ` + col + ` ` + dir + `, created_at DESC`
	if limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Invoice{}
	for rows.Next() {
		var i model.Invoice
		var paid int
		if err := rows.Scan(&i.ID, &i.BusinessID, &i.Amount, &i.Currency, &i.Description, &i.IssuedAt, &i.DueAt, &paid, &i.PaidAt, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		i.Paid = paid != 0
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *InvoiceRepo) DistinctBusinessIDs(ctx context.Context, from, to int64) ([]string, error) {
	where, args := invoiceFilterWhere(from, to, nil, nil, nil, false)
	if where == "" {
		where = " WHERE business_id IS NOT NULL"
	} else {
		where += " AND business_id IS NOT NULL"
	}
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT business_id FROM invoices`+where+` ORDER BY business_id`, args...)
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

func (r *InvoiceRepo) Count(ctx context.Context, from, to int64, businessIDs, userIDs, tagIDs []string, unpaidOnly bool) (int, error) {
	where, args := invoiceFilterWhere(from, to, businessIDs, userIDs, tagIDs, unpaidOnly)
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoices`+where, args...).Scan(&n)
	return n, err
}

func (r *InvoiceRepo) Get(ctx context.Context, id string) (model.Invoice, bool, error) {
	var i model.Invoice
	var paid int
	err := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(business_id, ''), amount, currency, description, issued_at, due_at, paid, paid_at, created_at, updated_at
		FROM invoices
		WHERE id = ?`, id).
		Scan(&i.ID, &i.BusinessID, &i.Amount, &i.Currency, &i.Description, &i.IssuedAt, &i.DueAt, &paid, &i.PaidAt, &i.CreatedAt, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		return model.Invoice{}, false, nil
	}
	if err != nil {
		return model.Invoice{}, false, err
	}
	i.Paid = paid != 0
	return i, true, nil
}

func (r *InvoiceRepo) Insert(ctx context.Context, i model.Invoice) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO invoices (id, business_id, amount, currency, description, issued_at, due_at, paid, paid_at, created_at, updated_at)
		VALUES (?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		i.ID, i.BusinessID, i.Amount, i.Currency, i.Description, i.IssuedAt, i.DueAt, boolToInt(i.Paid), i.PaidAt, i.CreatedAt, i.UpdatedAt)
	return err
}

func (r *InvoiceRepo) Update(ctx context.Context, i model.Invoice) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE invoices
		SET business_id = NULLIF(?, ''), amount = ?, currency = ?, description = ?, issued_at = ?, due_at = ?, paid = ?, paid_at = ?, updated_at = ?
		WHERE id = ?`,
		i.BusinessID, i.Amount, i.Currency, i.Description, i.IssuedAt, i.DueAt, boolToInt(i.Paid), i.PaidAt, i.UpdatedAt, i.ID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *InvoiceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM invoices WHERE id = ?`, id)
	return err
}

type businessAgg struct {
	amount          float64
	latestCreatedAt int64
}

func (r *InvoiceRepo) Stats(ctx context.Context, from, to int64, businessIDs, userIDs, tagIDs []string) ([]model.CurrencyStats, error) {
	where, args := invoiceFilterWhere(from, to, businessIDs, userIDs, tagIDs, false)
	rows, err := r.db.QueryContext(ctx,
		`SELECT COALESCE(business_id, ''), amount, currency, issued_at, paid, created_at FROM invoices`+where,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type agg struct {
		total, paid, unpaid           float64
		count, paidCount, unpaidCount int
		max                           float64
		byBusiness                    map[string]*businessAgg
		byDay                         map[int64]float64
	}
	byCurrency := map[string]*agg{}
	for rows.Next() {
		var businessID, currency string
		var amount float64
		var issuedAt, createdAt int64
		var paid int
		if err := rows.Scan(&businessID, &amount, &currency, &issuedAt, &paid, &createdAt); err != nil {
			return nil, err
		}
		a, ok := byCurrency[currency]
		if !ok {
			a = &agg{byBusiness: map[string]*businessAgg{}, byDay: map[int64]float64{}}
			byCurrency[currency] = a
		}
		a.total += amount
		a.count++
		if amount > a.max {
			a.max = amount
		}
		if paid != 0 {
			a.paid += amount
			a.paidCount++
		} else {
			a.unpaid += amount
			a.unpaidCount++
		}
		b, ok := a.byBusiness[businessID]
		if !ok {
			b = &businessAgg{}
			a.byBusiness[businessID] = b
		}
		b.amount += amount
		if createdAt > b.latestCreatedAt {
			b.latestCreatedAt = createdAt
		}
		dayEpoch := (issuedAt / 86400) * 86400
		a.byDay[dayEpoch] += amount
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.CurrencyStats, 0, len(byCurrency))
	for currency, a := range byCurrency {
		topBusinesses := pickTopBusinesses(a.byBusiness, 5)
		topDay, topDayAmount := pickTopDay(a.byDay)
		average := 0.0
		if a.count > 0 {
			average = a.total / float64(a.count)
		}
		out = append(out, model.CurrencyStats{
			Currency:      currency,
			Total:         a.total,
			Count:         a.count,
			PaidAmount:    a.paid,
			PaidCount:     a.paidCount,
			UnpaidAmount:  a.unpaid,
			UnpaidCount:   a.unpaidCount,
			Average:       average,
			MaxAmount:     a.max,
			BusinessCount: len(a.byBusiness),
			TopBusinesses: topBusinesses,
			TopDayEpoch:   topDay,
			TopDayAmount:  topDayAmount,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}
		return out[i].Currency < out[j].Currency
	})
	return out, nil
}

func pickTopBusinesses(m map[string]*businessAgg, limit int) []model.TopBusiness {
	type entry struct {
		id     string
		amount float64
		latest int64
	}
	list := make([]entry, 0, len(m))
	for id, b := range m {
		list = append(list, entry{id: id, amount: b.amount, latest: b.latestCreatedAt})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].amount != list[j].amount {
			return list[i].amount > list[j].amount
		}
		return list[i].latest > list[j].latest
	})
	if len(list) > limit {
		list = list[:limit]
	}
	out := make([]model.TopBusiness, len(list))
	for i, e := range list {
		out[i] = model.TopBusiness{BusinessID: e.id, Amount: e.amount}
	}
	return out
}

func pickTopDay(m map[int64]float64) (int64, float64) {
	var topDay int64
	var topAmount float64
	for d, a := range m {
		if a > topAmount || (a == topAmount && d > topDay) {
			topDay = d
			topAmount = a
		}
	}
	return topDay, topAmount
}
