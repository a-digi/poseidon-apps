package recurring

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"mbud-plugin/model"
	"mbud-plugin/storage"
)

const businessID = "biz1"

type fixture struct {
	db            *sql.DB
	business      *storage.BusinessRepo
	invoices      *storage.InvoiceRepo
	recurring     *storage.RecurringRepo
	upcoming      *storage.UpcomingRepo
	links         *storage.LinksRepo
	upcomingLinks *storage.UpcomingLinksRepo
	reconciler    *Reconciler
	now           int64
	counter       int
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db, err := storage.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	br := storage.NewBusinessRepo(db)
	ir := storage.NewInvoiceRepo(db)
	rr := storage.NewRecurringRepo(db)
	ur := storage.NewUpcomingRepo(db)
	lr := storage.NewLinksRepo(db)
	ulr := storage.NewUpcomingLinksRepo(db)

	ctx := context.Background()
	if err := br.Insert(ctx, model.Business{ID: businessID, Name: "Acme", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}

	fx := &fixture{
		db:            db,
		business:      br,
		invoices:      ir,
		recurring:     rr,
		upcoming:      ur,
		links:         lr,
		upcomingLinks: ulr,
		now:           time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC).Unix(),
	}
	fx.reconciler = New(db, rr, ur, ir, lr, ulr, fx.nextID, fx.nowFn)
	return fx
}

func (f *fixture) nextID() string {
	id := fmt.Sprintf("id%d", f.counter)
	f.counter++
	return id
}

func (f *fixture) nowFn() int64 { return f.now }

func (f *fixture) insertRule(t *testing.T, ctx context.Context, ri model.RecurringInvoice) {
	t.Helper()
	if ri.CreatedAt == 0 {
		ri.CreatedAt = f.now
	}
	if ri.UpdatedAt == 0 {
		ri.UpdatedAt = f.now
	}
	if ri.BusinessID == "" {
		ri.BusinessID = businessID
	}
	if err := f.recurring.Insert(ctx, ri); err != nil {
		t.Fatalf("recurring insert: %v", err)
	}
}

func (f *fixture) countInvoices(t *testing.T, ctx context.Context) int {
	t.Helper()
	var n int
	if err := f.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoices`).Scan(&n); err != nil {
		t.Fatalf("count invoices: %v", err)
	}
	return n
}

func (f *fixture) countLinks(t *testing.T, ctx context.Context, ruleID string) int {
	t.Helper()
	var n int
	if err := f.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM recurring_invoice_links WHERE recurring_id = ?`, ruleID).Scan(&n); err != nil {
		t.Fatalf("count links: %v", err)
	}
	return n
}

func TestNextDueAt(t *testing.T) {
	cases := []struct {
		name  string
		start int64
		freq  model.Frequency
		index int
		want  int64
	}{
		{"daily 0", 0, model.FrequencyDaily, 0, 0},
		{"daily 1", 0, model.FrequencyDaily, 1, 86400},
		{"daily 7", 0, model.FrequencyDaily, 7, 7 * 86400},
		{"weekly 1", 0, model.FrequencyWeekly, 1, 7 * 86400},
		{"weekly 3", 0, model.FrequencyWeekly, 3, 21 * 86400},
		{"monthly 12 == yearly 1", 0, model.FrequencyMonthly, 12,
			NextDueAt(model.RecurringInvoice{StartAt: 0, Frequency: model.FrequencyYearly}, 1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := NextDueAt(model.RecurringInvoice{StartAt: c.start, Frequency: c.freq}, c.index)
			if got != c.want {
				t.Fatalf("NextDueAt(%d, %s, %d) = %d, want %d", c.start, c.freq, c.index, got, c.want)
			}
		})
	}

	t.Run("yearly +4 calendar years", func(t *testing.T) {
		start := time.Date(2020, 3, 15, 12, 30, 45, 0, time.UTC)
		got := NextDueAt(model.RecurringInvoice{StartAt: start.Unix(), Frequency: model.FrequencyYearly}, 4)
		want := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC).Unix()
		if got != want {
			t.Fatalf("yearly +4 = %d, want %d", got, want)
		}
	})

	t.Run("monthly across year boundary", func(t *testing.T) {
		start := time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC)
		got := NextDueAt(model.RecurringInvoice{StartAt: start.Unix(), Frequency: model.FrequencyMonthly}, 3)
		want := time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC).Unix()
		if got != want {
			t.Fatalf("monthly cross-year = %d, want %d", got, want)
		}
	})
}

func TestCatchup_HappyPath(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	ruleStart := time.Date(2025, 11, 13, 0, 0, 0, 0, time.UTC).Unix()
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   ruleStart,
		Active:    true,
	})

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}

	if got := fx.countLinks(t, ctx, "rule1"); got != 7 {
		t.Fatalf("links = %d, want 7", got)
	}
	if got := fx.countInvoices(t, ctx); got != 7 {
		t.Fatalf("invoices = %d, want 7", got)
	}

	maxN, err := fx.links.MaxPeriodIndex(ctx, "rule1")
	if err != nil {
		t.Fatalf("max: %v", err)
	}
	if maxN != 6 {
		t.Fatalf("max period = %d, want 6", maxN)
	}
}

func TestCatchup_Idempotent(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   time.Date(2025, 11, 13, 0, 0, 0, 0, time.UTC).Unix(),
		Active:    true,
	})

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup 1: %v", err)
	}
	firstInv := fx.countInvoices(t, ctx)
	firstLinks := fx.countLinks(t, ctx, "rule1")

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup 2: %v", err)
	}
	if got := fx.countInvoices(t, ctx); got != firstInv {
		t.Fatalf("invoices after 2nd catchup = %d, want %d", got, firstInv)
	}
	if got := fx.countLinks(t, ctx, "rule1"); got != firstLinks {
		t.Fatalf("links after 2nd catchup = %d, want %d", got, firstLinks)
	}
}

func TestReconcile_WipesAndRegenerates(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   time.Date(2025, 11, 13, 0, 0, 0, 0, time.UTC).Unix(),
		Active:    true,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}
	oldIDs, err := fx.links.InvoiceIDsByRecurring(ctx, "rule1")
	if err != nil {
		t.Fatalf("old ids: %v", err)
	}

	if _, err := fx.db.ExecContext(ctx,
		`UPDATE recurring_invoices SET amount = 99 WHERE id = ?`, "rule1"); err != nil {
		t.Fatalf("amount update: %v", err)
	}

	if err := fx.reconciler.Reconcile(ctx, "rule1"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	for _, oid := range oldIDs {
		if _, ok, _ := fx.invoices.Get(ctx, oid); ok {
			t.Fatalf("old invoice %s still exists after reconcile", oid)
		}
	}

	rows, err := fx.db.QueryContext(ctx, `SELECT amount FROM invoices`)
	if err != nil {
		t.Fatalf("query amounts: %v", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var amt float64
		if err := rows.Scan(&amt); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if amt != 99 {
			t.Fatalf("amount = %v, want 99", amt)
		}
		count++
	}
	if count != 7 {
		t.Fatalf("new invoice count = %d, want 7", count)
	}
}

func TestReconcile_ActiveFalseWipes(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   time.Date(2025, 11, 13, 0, 0, 0, 0, time.UTC).Unix(),
		Active:    true,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}
	if fx.countLinks(t, ctx, "rule1") == 0 {
		t.Fatal("precondition: links should exist")
	}

	if _, err := fx.db.ExecContext(ctx,
		`UPDATE recurring_invoices SET active = 0 WHERE id = ?`, "rule1"); err != nil {
		t.Fatalf("update active: %v", err)
	}
	if err := fx.reconciler.Reconcile(ctx, "rule1"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if got := fx.countLinks(t, ctx, "rule1"); got != 0 {
		t.Fatalf("links after active=false reconcile = %d, want 0", got)
	}
	if got := fx.countInvoices(t, ctx); got != 0 {
		t.Fatalf("invoices after active=false reconcile = %d, want 0", got)
	}
}

func TestReconcile_ActiveTrueRegens(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   time.Date(2025, 11, 13, 0, 0, 0, 0, time.UTC).Unix(),
		Active:    true,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}
	if _, err := fx.db.ExecContext(ctx,
		`UPDATE recurring_invoices SET active = 0 WHERE id = ?`, "rule1"); err != nil {
		t.Fatalf("set inactive: %v", err)
	}
	if err := fx.reconciler.Reconcile(ctx, "rule1"); err != nil {
		t.Fatalf("reconcile inactive: %v", err)
	}
	if got := fx.countInvoices(t, ctx); got != 0 {
		t.Fatalf("precondition: invoices = %d, want 0", got)
	}

	if _, err := fx.db.ExecContext(ctx,
		`UPDATE recurring_invoices SET active = 1 WHERE id = ?`, "rule1"); err != nil {
		t.Fatalf("set active: %v", err)
	}
	if err := fx.reconciler.Reconcile(ctx, "rule1"); err != nil {
		t.Fatalf("reconcile active: %v", err)
	}

	if got := fx.countLinks(t, ctx, "rule1"); got != 7 {
		t.Fatalf("links after regen = %d, want 7", got)
	}
	if got := fx.countInvoices(t, ctx); got != 7 {
		t.Fatalf("invoices after regen = %d, want 7", got)
	}
}

func TestCatchup_UserDeleteSticks(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   time.Date(2025, 11, 13, 0, 0, 0, 0, time.UTC).Unix(),
		Active:    true,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}

	var victim string
	if err := fx.db.QueryRowContext(ctx,
		`SELECT invoice_id FROM recurring_invoice_links WHERE recurring_id = ? AND period_index = 3`,
		"rule1").Scan(&victim); err != nil {
		t.Fatalf("pick victim: %v", err)
	}
	if err := fx.invoices.Delete(ctx, victim); err != nil {
		t.Fatalf("delete invoice: %v", err)
	}

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup 2: %v", err)
	}

	if _, ok, _ := fx.invoices.Get(ctx, victim); ok {
		t.Fatalf("deleted invoice %s came back", victim)
	}

	var linkN int
	if err := fx.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM recurring_invoice_links WHERE recurring_id = ? AND period_index = 3`,
		"rule1").Scan(&linkN); err != nil {
		t.Fatalf("count link 3: %v", err)
	}
	if linkN != 1 {
		t.Fatalf("link for period 3 = %d, want 1 (orphan survives)", linkN)
	}
}

func TestCatchup_EndAtClamps(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	start := time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC)
	endAt := time.Date(2025, 8, 13, 0, 0, 0, 0, time.UTC).Unix()
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   start.Unix(),
		EndAt:     endAt,
		Active:    true,
	})

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}

	if got := fx.countLinks(t, ctx, "rule1"); got != 4 {
		t.Fatalf("links = %d, want 4 (period 0..3)", got)
	}
}

func TestCatchup_NoFuturePeriods(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	start := time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC).Unix()
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   start,
		Active:    true,
	})

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}

	if got := fx.countLinks(t, ctx, "rule1"); got != 1 {
		t.Fatalf("links = %d, want 1 (only period 0 due)", got)
	}
}

func TestWipeOnly(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   time.Date(2025, 11, 13, 0, 0, 0, 0, time.UTC).Unix(),
		Active:    true,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}

	if err := fx.reconciler.WipeOnly(ctx, "rule1"); err != nil {
		t.Fatalf("wipe: %v", err)
	}

	if got := fx.countInvoices(t, ctx); got != 0 {
		t.Fatalf("invoices after WipeOnly = %d, want 0", got)
	}
	if got := fx.countLinks(t, ctx, "rule1"); got != 0 {
		t.Fatalf("links after WipeOnly = %d, want 0", got)
	}

	if _, ok, err := fx.recurring.Get(ctx, "rule1"); err != nil || !ok {
		t.Fatalf("rule should still exist: ok=%v err=%v", ok, err)
	}
}

func (f *fixture) insertUpcoming(t *testing.T, ctx context.Context, u model.UpcomingInvoice) {
	t.Helper()
	if u.CreatedAt == 0 {
		u.CreatedAt = f.now
	}
	if u.UpdatedAt == 0 {
		u.UpdatedAt = f.now
	}
	if u.BusinessID == "" {
		u.BusinessID = businessID
	}
	if err := f.upcoming.Insert(ctx, u); err != nil {
		t.Fatalf("upcoming insert: %v", err)
	}
}

func (f *fixture) countUpcomingLinks(t *testing.T, ctx context.Context, upcomingID string) int {
	t.Helper()
	var n int
	if err := f.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM upcoming_invoice_links WHERE upcoming_id = ?`, upcomingID).Scan(&n); err != nil {
		t.Fatalf("count upcoming links: %v", err)
	}
	return n
}

func TestCatchup_UpcomingMaterialises(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:          "up1",
		Amount:      42,
		Currency:    "EUR",
		Description: "rent",
		DueAt:       fx.now - 3600,
	})

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}

	if got := fx.countInvoices(t, ctx); got != 1 {
		t.Fatalf("invoices = %d, want 1", got)
	}
	if got := fx.countUpcomingLinks(t, ctx, "up1"); got != 1 {
		t.Fatalf("upcoming links = %d, want 1", got)
	}

	var amount float64
	var currency, description string
	var issuedAt, dueAt int64
	if err := fx.db.QueryRowContext(ctx,
		`SELECT amount, currency, description, issued_at, due_at FROM invoices`).
		Scan(&amount, &currency, &description, &issuedAt, &dueAt); err != nil {
		t.Fatalf("scan invoice: %v", err)
	}
	if amount != 42 || currency != "EUR" || description != "rent" {
		t.Fatalf("invoice fields = %v/%v/%v, want 42/EUR/rent", amount, currency, description)
	}
	if issuedAt != fx.now-3600 || dueAt != fx.now-3600 {
		t.Fatalf("issued/due = %d/%d, want both %d", issuedAt, dueAt, fx.now-3600)
	}
}

func TestCatchup_UpcomingNotYetDue(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   10,
		Currency: "EUR",
		DueAt:    fx.now + 86400,
	})

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}

	if got := fx.countInvoices(t, ctx); got != 0 {
		t.Fatalf("invoices = %d, want 0", got)
	}
	if got := fx.countUpcomingLinks(t, ctx, "up1"); got != 0 {
		t.Fatalf("upcoming links = %d, want 0", got)
	}
}

func TestCatchup_UpcomingIdempotent(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   10,
		Currency: "EUR",
		DueAt:    fx.now - 3600,
	})

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup 1: %v", err)
	}
	firstInv := fx.countInvoices(t, ctx)
	firstLinks := fx.countUpcomingLinks(t, ctx, "up1")

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup 2: %v", err)
	}
	if got := fx.countInvoices(t, ctx); got != firstInv {
		t.Fatalf("invoices after 2nd catchup = %d, want %d", got, firstInv)
	}
	if got := fx.countUpcomingLinks(t, ctx, "up1"); got != firstLinks {
		t.Fatalf("links after 2nd catchup = %d, want %d", got, firstLinks)
	}
	if firstInv != 1 || firstLinks != 1 {
		t.Fatalf("first counts inv=%d links=%d, want 1/1", firstInv, firstLinks)
	}
}

func TestCatchup_UpcomingUserDeleteSticks(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   10,
		Currency: "EUR",
		DueAt:    fx.now - 3600,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}

	var victim string
	if err := fx.db.QueryRowContext(ctx,
		`SELECT invoice_id FROM upcoming_invoice_links WHERE upcoming_id = ?`, "up1").Scan(&victim); err != nil {
		t.Fatalf("pick victim: %v", err)
	}
	if _, err := fx.db.ExecContext(ctx, `DELETE FROM invoices WHERE id = ?`, victim); err != nil {
		t.Fatalf("delete invoice: %v", err)
	}

	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup 2: %v", err)
	}

	if got := fx.countInvoices(t, ctx); got != 0 {
		t.Fatalf("invoices after user-delete catchup = %d, want 0", got)
	}
	if got := fx.countUpcomingLinks(t, ctx, "up1"); got != 1 {
		t.Fatalf("links = %d, want 1 (orphan survives)", got)
	}
}

func TestReconcileUpcoming_WipesAndRegens(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   10,
		Currency: "EUR",
		DueAt:    fx.now - 3600,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}
	oldIDs, err := fx.upcomingLinks.InvoiceIDsByUpcoming(ctx, "up1")
	if err != nil {
		t.Fatalf("old ids: %v", err)
	}
	if len(oldIDs) != 1 {
		t.Fatalf("oldIDs = %d, want 1", len(oldIDs))
	}

	if _, err := fx.db.ExecContext(ctx,
		`UPDATE upcoming_invoices SET amount = 99 WHERE id = ?`, "up1"); err != nil {
		t.Fatalf("amount update: %v", err)
	}

	if err := fx.reconciler.ReconcileUpcoming(ctx, "up1"); err != nil {
		t.Fatalf("reconcile upcoming: %v", err)
	}

	if _, ok, _ := fx.invoices.Get(ctx, oldIDs[0]); ok {
		t.Fatalf("old invoice %s still exists after reconcile", oldIDs[0])
	}

	var amount float64
	if err := fx.db.QueryRowContext(ctx, `SELECT amount FROM invoices`).Scan(&amount); err != nil {
		t.Fatalf("scan amount: %v", err)
	}
	if amount != 99 {
		t.Fatalf("amount = %v, want 99", amount)
	}

	newIDs, err := fx.upcomingLinks.InvoiceIDsByUpcoming(ctx, "up1")
	if err != nil {
		t.Fatalf("new ids: %v", err)
	}
	if len(newIDs) != 1 {
		t.Fatalf("newIDs = %d, want 1", len(newIDs))
	}
	if newIDs[0] == oldIDs[0] {
		t.Fatalf("link did not update: still points to %s", oldIDs[0])
	}
}

func TestReconcileUpcoming_OnlyEmitsWhenDue(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   10,
		Currency: "EUR",
		DueAt:    fx.now + 86400,
	})

	if err := fx.reconciler.ReconcileUpcoming(ctx, "up1"); err != nil {
		t.Fatalf("reconcile upcoming: %v", err)
	}

	if got := fx.countInvoices(t, ctx); got != 0 {
		t.Fatalf("invoices = %d, want 0", got)
	}
	if got := fx.countUpcomingLinks(t, ctx, "up1"); got != 0 {
		t.Fatalf("links = %d, want 0", got)
	}
}

func TestRecurringDeleteWipesMaterialisedInvoices(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	const ruleID = "rule1"
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        ruleID,
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   time.Date(2025, 11, 13, 0, 0, 0, 0, time.UTC).Unix(),
		Active:    true,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}
	if fx.countInvoices(t, ctx) == 0 {
		t.Fatal("precondition: invoices should exist")
	}
	if fx.countLinks(t, ctx, ruleID) == 0 {
		t.Fatal("precondition: links should exist")
	}

	if err := fx.reconciler.WipeOnly(ctx, ruleID); err != nil {
		t.Fatalf("wipe: %v", err)
	}

	var invN int
	if err := fx.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM invoices WHERE business_id = ?`, businessID).Scan(&invN); err != nil {
		t.Fatalf("count invoices: %v", err)
	}
	if invN != 0 {
		t.Fatalf("invoices for business = %d, want 0", invN)
	}
	if got := fx.countLinks(t, ctx, ruleID); got != 0 {
		t.Fatalf("links after WipeOnly = %d, want 0", got)
	}
	var ruleN int
	if err := fx.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM recurring_invoices WHERE id = ?`, ruleID).Scan(&ruleN); err != nil {
		t.Fatalf("count rule: %v", err)
	}
	if ruleN != 1 {
		t.Fatalf("recurring rule rows = %d, want 1 (WipeOnly preserves the rule)", ruleN)
	}
}

func TestUpcomingDeleteWipesMaterialisedInvoice(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	const upcomingID = "up1"
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       upcomingID,
		Amount:   10,
		Currency: "EUR",
		DueAt:    fx.now - 86400,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}
	if got := fx.countInvoices(t, ctx); got != 1 {
		t.Fatalf("invoices after catchup = %d, want 1", got)
	}
	if got := fx.countUpcomingLinks(t, ctx, upcomingID); got != 1 {
		t.Fatalf("links after catchup = %d, want 1", got)
	}

	if err := fx.reconciler.WipeOnlyUpcoming(ctx, upcomingID); err != nil {
		t.Fatalf("wipe upcoming: %v", err)
	}

	var invN int
	if err := fx.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM invoices WHERE business_id = ?`, businessID).Scan(&invN); err != nil {
		t.Fatalf("count invoices: %v", err)
	}
	if invN != 0 {
		t.Fatalf("invoices for business = %d, want 0", invN)
	}
	if got := fx.countUpcomingLinks(t, ctx, upcomingID); got != 0 {
		t.Fatalf("links after WipeOnlyUpcoming = %d, want 0", got)
	}
	var upN int
	if err := fx.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM upcoming_invoices WHERE id = ?`, upcomingID).Scan(&upN); err != nil {
		t.Fatalf("count upcoming: %v", err)
	}
	if upN != 1 {
		t.Fatalf("upcoming rows = %d, want 1 (WipeOnlyUpcoming preserves the upcoming)", upN)
	}
}

func TestWipeOnlyUpcoming(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   10,
		Currency: "EUR",
		DueAt:    fx.now - 3600,
	})
	if err := fx.reconciler.Catchup(ctx); err != nil {
		t.Fatalf("catchup: %v", err)
	}

	if err := fx.reconciler.WipeOnlyUpcoming(ctx, "up1"); err != nil {
		t.Fatalf("wipe: %v", err)
	}

	if got := fx.countInvoices(t, ctx); got != 0 {
		t.Fatalf("invoices = %d, want 0", got)
	}
	if got := fx.countUpcomingLinks(t, ctx, "up1"); got != 0 {
		t.Fatalf("links = %d, want 0", got)
	}
	if _, ok, err := fx.upcoming.Get(ctx, "up1"); err != nil || !ok {
		t.Fatalf("upcoming should still exist: ok=%v err=%v", ok, err)
	}
}

func (f *fixture) insertUser(t *testing.T, ctx context.Context, id, name string) {
	t.Helper()
	if _, err := f.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, notes, created_at, updated_at) VALUES (?, ?, '', '', ?, ?)`,
		id, name, f.now, f.now); err != nil {
		t.Fatalf("insert user %s: %v", id, err)
	}
}

func (f *fixture) invoiceUserIDs(t *testing.T, ctx context.Context, invoiceID string) []string {
	t.Helper()
	rows, err := f.db.QueryContext(ctx,
		`SELECT user_id FROM invoice_users WHERE invoice_id = ? ORDER BY user_id`, invoiceID)
	if err != nil {
		t.Fatalf("query invoice_users: %v", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, id)
	}
	return out
}

func (f *fixture) materialisedRecurringInvoiceIDs(t *testing.T, ctx context.Context, ruleID string) []string {
	t.Helper()
	rows, err := f.db.QueryContext(ctx,
		`SELECT invoice_id FROM recurring_invoice_links WHERE recurring_id = ?`, ruleID)
	if err != nil {
		t.Fatalf("query links: %v", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, id)
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestEmit_PropagatesRecurringUserIDs(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUser(t, ctx, "u1", "Alice")
	fx.insertUser(t, ctx, "u2", "Bob")
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:              "r1",
		Amount:          100,
		Currency:        "EUR",
		Frequency:       model.FrequencyMonthly,
		StartAt:         fx.now - 35*86400,
		IssueDayOfMonth: 1,
		Active:          true,
	})
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO recurring_users (recurring_id, user_id) VALUES (?, ?), (?, ?)`,
		"r1", "u1", "r1", "u2"); err != nil {
		t.Fatalf("seed recurring_users: %v", err)
	}

	if err := fx.reconciler.Reconcile(ctx, "r1"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	invIDs := fx.materialisedRecurringInvoiceIDs(t, ctx, "r1")
	if len(invIDs) < 1 {
		t.Fatalf("materialised invoices = %d, want >= 1", len(invIDs))
	}
	for _, id := range invIDs {
		got := fx.invoiceUserIDs(t, ctx, id)
		if !equalStrings(got, []string{"u1", "u2"}) {
			t.Fatalf("invoice %s users = %v, want [u1 u2]", id, got)
		}
	}
}

func TestEmitUpcoming_PropagatesUpcomingUserIDs(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUser(t, ctx, "u1", "Alice")
	fx.insertUser(t, ctx, "u2", "Bob")
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   50,
		Currency: "EUR",
		DueAt:    fx.now - 86400,
	})
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO upcoming_users (upcoming_id, user_id) VALUES (?, ?), (?, ?)`,
		"up1", "u1", "up1", "u2"); err != nil {
		t.Fatalf("seed upcoming_users: %v", err)
	}

	if err := fx.reconciler.ReconcileUpcoming(ctx, "up1"); err != nil {
		t.Fatalf("reconcile upcoming: %v", err)
	}

	var invoiceID string
	if err := fx.db.QueryRowContext(ctx,
		`SELECT invoice_id FROM upcoming_invoice_links WHERE upcoming_id = ?`, "up1").Scan(&invoiceID); err != nil {
		t.Fatalf("pick materialised invoice: %v", err)
	}
	got := fx.invoiceUserIDs(t, ctx, invoiceID)
	if !equalStrings(got, []string{"u1", "u2"}) {
		t.Fatalf("invoice %s users = %v, want [u1 u2]", invoiceID, got)
	}
}

func TestReconcile_UpdatedUserListPropagatesAfterReconcile(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertUser(t, ctx, "u1", "Alice")
	fx.insertUser(t, ctx, "u2", "Bob")
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:              "r1",
		Amount:          100,
		Currency:        "EUR",
		Frequency:       model.FrequencyMonthly,
		StartAt:         fx.now - 35*86400,
		IssueDayOfMonth: 1,
		Active:          true,
	})
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO recurring_users (recurring_id, user_id) VALUES (?, ?)`,
		"r1", "u1"); err != nil {
		t.Fatalf("seed recurring_users: %v", err)
	}

	if err := fx.reconciler.Reconcile(ctx, "r1"); err != nil {
		t.Fatalf("reconcile 1: %v", err)
	}
	for _, id := range fx.materialisedRecurringInvoiceIDs(t, ctx, "r1") {
		got := fx.invoiceUserIDs(t, ctx, id)
		if !equalStrings(got, []string{"u1"}) {
			t.Fatalf("invoice %s users (1st) = %v, want [u1]", id, got)
		}
	}

	if _, err := fx.db.ExecContext(ctx, `DELETE FROM recurring_users WHERE recurring_id = ?`, "r1"); err != nil {
		t.Fatalf("wipe recurring_users: %v", err)
	}
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO recurring_users (recurring_id, user_id) VALUES (?, ?)`, "r1", "u2"); err != nil {
		t.Fatalf("insert u2: %v", err)
	}

	if err := fx.reconciler.Reconcile(ctx, "r1"); err != nil {
		t.Fatalf("reconcile 2: %v", err)
	}
	for _, id := range fx.materialisedRecurringInvoiceIDs(t, ctx, "r1") {
		got := fx.invoiceUserIDs(t, ctx, id)
		if !equalStrings(got, []string{"u2"}) {
			t.Fatalf("invoice %s users (2nd) = %v, want [u2]", id, got)
		}
	}
}

func (f *fixture) insertTag(t *testing.T, ctx context.Context, id, name string) {
	t.Helper()
	if _, err := f.db.ExecContext(ctx,
		`INSERT INTO tags (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		id, name, f.now, f.now); err != nil {
		t.Fatalf("insert tag %s: %v", id, err)
	}
}

func (f *fixture) invoiceTagIDs(t *testing.T, ctx context.Context, invoiceID string) []string {
	t.Helper()
	rows, err := f.db.QueryContext(ctx,
		`SELECT tag_id FROM invoice_tags WHERE invoice_id = ? ORDER BY tag_id`, invoiceID)
	if err != nil {
		t.Fatalf("query invoice_tags: %v", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, id)
	}
	return out
}

func TestEmit_PropagatesRecurringTagIDs(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertTag(t, ctx, "t1", "alpha")
	fx.insertTag(t, ctx, "t2", "beta")
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:              "r1",
		Amount:          100,
		Currency:        "EUR",
		Frequency:       model.FrequencyMonthly,
		StartAt:         fx.now - 35*86400,
		IssueDayOfMonth: 1,
		Active:          true,
	})
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO recurring_tags (recurring_id, tag_id) VALUES (?, ?), (?, ?)`,
		"r1", "t1", "r1", "t2"); err != nil {
		t.Fatalf("seed recurring_tags: %v", err)
	}

	if err := fx.reconciler.Reconcile(ctx, "r1"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	invIDs := fx.materialisedRecurringInvoiceIDs(t, ctx, "r1")
	if len(invIDs) < 1 {
		t.Fatalf("materialised invoices = %d, want >= 1", len(invIDs))
	}
	for _, id := range invIDs {
		got := fx.invoiceTagIDs(t, ctx, id)
		if !equalStrings(got, []string{"t1", "t2"}) {
			t.Fatalf("invoice %s tags = %v, want [t1 t2]", id, got)
		}
	}
}

func TestEmitUpcoming_PropagatesUpcomingTagIDs(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertTag(t, ctx, "t1", "alpha")
	fx.insertTag(t, ctx, "t2", "beta")
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   50,
		Currency: "EUR",
		DueAt:    fx.now - 86400,
	})
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO upcoming_tags (upcoming_id, tag_id) VALUES (?, ?), (?, ?)`,
		"up1", "t1", "up1", "t2"); err != nil {
		t.Fatalf("seed upcoming_tags: %v", err)
	}

	if err := fx.reconciler.ReconcileUpcoming(ctx, "up1"); err != nil {
		t.Fatalf("reconcile upcoming: %v", err)
	}

	var invoiceID string
	if err := fx.db.QueryRowContext(ctx,
		`SELECT invoice_id FROM upcoming_invoice_links WHERE upcoming_id = ?`, "up1").Scan(&invoiceID); err != nil {
		t.Fatalf("pick materialised invoice: %v", err)
	}
	got := fx.invoiceTagIDs(t, ctx, invoiceID)
	if !equalStrings(got, []string{"t1", "t2"}) {
		t.Fatalf("invoice %s tags = %v, want [t1 t2]", invoiceID, got)
	}
}

func TestEmit_DefaultsPaidTrue(t *testing.T) {
	t.Run("recurring", func(t *testing.T) {
		ctx := context.Background()
		fx := newFixture(t)
		if err := fx.business.Insert(ctx, model.Business{ID: "b1", Name: "B1", CreatedAt: fx.now, UpdatedAt: fx.now}); err != nil {
			t.Fatalf("biz insert: %v", err)
		}
		fx.insertRule(t, ctx, model.RecurringInvoice{
			ID:              "r1",
			BusinessID:      "b1",
			Amount:          100,
			Currency:        "EUR",
			Frequency:       model.FrequencyMonthly,
			StartAt:         fx.now - 35*86400,
			IssueDayOfMonth: 1,
			Active:          true,
		})

		if err := fx.reconciler.Reconcile(ctx, "r1"); err != nil {
			t.Fatalf("reconcile: %v", err)
		}

		var paidCount int
		if err := fx.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM invoices WHERE id IN (SELECT invoice_id FROM recurring_invoice_links WHERE recurring_id = ?) AND paid = 1`,
			"r1").Scan(&paidCount); err != nil {
			t.Fatalf("count paid: %v", err)
		}
		total := fx.countLinks(t, ctx, "r1")
		if paidCount == 0 {
			t.Fatalf("paid count = 0, want > 0")
		}
		if paidCount != total {
			t.Fatalf("paid count = %d, want %d (all materialised)", paidCount, total)
		}
	})

	t.Run("upcoming", func(t *testing.T) {
		ctx := context.Background()
		fx := newFixture(t)
		if err := fx.business.Insert(ctx, model.Business{ID: "b1", Name: "B1", CreatedAt: fx.now, UpdatedAt: fx.now}); err != nil {
			t.Fatalf("biz insert: %v", err)
		}
		fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
			ID:         "up1",
			BusinessID: "b1",
			Amount:     50,
			Currency:   "EUR",
			DueAt:      fx.now - 86400,
		})

		if err := fx.reconciler.ReconcileUpcoming(ctx, "up1"); err != nil {
			t.Fatalf("reconcile upcoming: %v", err)
		}

		rows, err := fx.db.QueryContext(ctx,
			`SELECT paid FROM invoices WHERE id IN (SELECT invoice_id FROM upcoming_invoice_links WHERE upcoming_id = ?)`,
			"up1")
		if err != nil {
			t.Fatalf("query paid: %v", err)
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			var paid int
			if err := rows.Scan(&paid); err != nil {
				t.Fatalf("scan: %v", err)
			}
			if paid != 1 {
				t.Fatalf("paid = %d, want 1", paid)
			}
			count++
		}
		if count != 1 {
			t.Fatalf("rows = %d, want 1", count)
		}
	})
}

func TestReconcile_UpdatedTagListPropagatesAfterReconcile(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertTag(t, ctx, "t1", "alpha")
	fx.insertTag(t, ctx, "t2", "beta")
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:              "r1",
		Amount:          100,
		Currency:        "EUR",
		Frequency:       model.FrequencyMonthly,
		StartAt:         fx.now - 35*86400,
		IssueDayOfMonth: 1,
		Active:          true,
	})
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO recurring_tags (recurring_id, tag_id) VALUES (?, ?)`,
		"r1", "t1"); err != nil {
		t.Fatalf("seed recurring_tags: %v", err)
	}

	if err := fx.reconciler.Reconcile(ctx, "r1"); err != nil {
		t.Fatalf("reconcile 1: %v", err)
	}
	for _, id := range fx.materialisedRecurringInvoiceIDs(t, ctx, "r1") {
		got := fx.invoiceTagIDs(t, ctx, id)
		if !equalStrings(got, []string{"t1"}) {
			t.Fatalf("invoice %s tags (1st) = %v, want [t1]", id, got)
		}
	}

	if _, err := fx.db.ExecContext(ctx, `DELETE FROM recurring_tags WHERE recurring_id = ?`, "r1"); err != nil {
		t.Fatalf("wipe recurring_tags: %v", err)
	}
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO recurring_tags (recurring_id, tag_id) VALUES (?, ?)`, "r1", "t2"); err != nil {
		t.Fatalf("insert t2: %v", err)
	}

	if err := fx.reconciler.Reconcile(ctx, "r1"); err != nil {
		t.Fatalf("reconcile 2: %v", err)
	}
	for _, id := range fx.materialisedRecurringInvoiceIDs(t, ctx, "r1") {
		got := fx.invoiceTagIDs(t, ctx, id)
		if !equalStrings(got, []string{"t2"}) {
			t.Fatalf("invoice %s tags (2nd) = %v, want [t2]", id, got)
		}
	}
}
