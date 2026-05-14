package recurring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"mbud-plugin/model"
)

func TestProjectPending_RecurringFutureOnly(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	start := fx.now - 35*86400
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    99,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   start,
		Active:    true,
	})
	if err := fx.reconciler.Reconcile(ctx, "rule1"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 365*86400})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if len(pending) == 0 {
		t.Fatal("pending = 0, want > 0")
	}
	for i, p := range pending {
		if p.DueAt <= fx.now {
			t.Fatalf("pending[%d].DueAt=%d not > now=%d", i, p.DueAt, fx.now)
		}
		if i > 0 && pending[i-1].DueAt > p.DueAt {
			t.Fatalf("not sorted ASC at %d", i)
		}
		var n int
		if err := fx.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM recurring_invoice_links
			 JOIN invoices ON invoices.id = recurring_invoice_links.invoice_id
			 WHERE recurring_invoice_links.recurring_id = ? AND invoices.due_at = ?`,
			p.SourceID, p.DueAt).Scan(&n); err != nil {
			t.Fatalf("link count: %v", err)
		}
		if n != 0 {
			t.Fatalf("pending[%d] (due=%d) already materialised", i, p.DueAt)
		}
	}
}

func TestProjectPending_UpcomingFutureOnly(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	due := fx.now + 3*86400
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   77,
		Currency: "EUR",
		DueAt:    due,
	})

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 30*86400})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("len = %d, want 1", len(pending))
	}
	if pending[0].Source != "upcoming" || pending[0].SourceID != "up1" || pending[0].DueAt != due {
		t.Fatalf("got %+v, want source=upcoming sourceID=up1 dueAt=%d", pending[0], due)
	}
}

func TestProjectPending_TimeframeBounds(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   fx.now - 35*86400,
		Active:    true,
	})
	if err := fx.reconciler.Reconcile(ctx, "rule1"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 90*86400})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if len(pending) > 3 {
		t.Fatalf("len = %d, want <= 3", len(pending))
	}
	for _, p := range pending {
		if p.DueAt > fx.now+90*86400 {
			t.Fatalf("DueAt=%d > To=%d", p.DueAt, fx.now+90*86400)
		}
	}
}

func TestProjectPending_BusinessFilter(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	if err := fx.business.Insert(ctx, model.Business{ID: "biz2", Name: "Other", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz2: %v", err)
	}
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:         "rule1",
		BusinessID: businessID,
		Amount:     10,
		Currency:   "EUR",
		Frequency:  model.FrequencyMonthly,
		StartAt:    fx.now - 35*86400,
		Active:     true,
	})
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:         "rule2",
		BusinessID: "biz2",
		Amount:     10,
		Currency:   "EUR",
		Frequency:  model.FrequencyMonthly,
		StartAt:    fx.now - 35*86400,
		Active:     true,
	})

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 90*86400, BusinessIDs: []string{businessID}})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if len(pending) == 0 {
		t.Fatal("len = 0, want > 0")
	}
	for _, p := range pending {
		if p.BusinessID != businessID {
			t.Fatalf("BusinessID=%q, want %q", p.BusinessID, businessID)
		}
	}
}

func TestProjectPending_TagFilter(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	if _, err := fx.db.ExecContext(ctx, `INSERT INTO tags (id, name, created_at, updated_at) VALUES ('t1','A',1,1),('t2','B',1,1)`); err != nil {
		t.Fatalf("tags: %v", err)
	}
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   fx.now - 35*86400,
		Active:    true,
	})
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO recurring_tags (recurring_id, tag_id) VALUES ('rule1','t1'),('rule1','t2')`); err != nil {
		t.Fatalf("recurring_tags: %v", err)
	}

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 90*86400, TagIDs: []string{"t2"}})
	if err != nil {
		t.Fatalf("project t2: %v", err)
	}
	if len(pending) == 0 {
		t.Fatal("len = 0, want > 0")
	}

	pending, err = ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 90*86400, TagIDs: []string{"tNope"}})
	if err != nil {
		t.Fatalf("project tNope: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("len = %d, want 0", len(pending))
	}
}

func TestProjectPending_UserFilter(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	if _, err := fx.db.ExecContext(ctx, `INSERT INTO users (id, name, created_at, updated_at) VALUES ('u1','A',1,1),('u2','B',1,1)`); err != nil {
		t.Fatalf("users: %v", err)
	}
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   fx.now - 35*86400,
		Active:    true,
	})
	if _, err := fx.db.ExecContext(ctx,
		`INSERT INTO recurring_users (recurring_id, user_id) VALUES ('rule1','u1'),('rule1','u2')`); err != nil {
		t.Fatalf("recurring_users: %v", err)
	}

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 90*86400, UserIDs: []string{"u2"}})
	if err != nil {
		t.Fatalf("project u2: %v", err)
	}
	if len(pending) == 0 {
		t.Fatal("len = 0, want > 0")
	}

	pending, err = ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 90*86400, UserIDs: []string{"uNope"}})
	if err != nil {
		t.Fatalf("project uNope: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("len = %d, want 0", len(pending))
	}
}

func TestProjectPending_EndAtClips(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	endAt := fx.now + 60*86400
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   fx.now - 35*86400,
		EndAt:     endAt,
		Active:    true,
	})

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 365*86400})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	for _, p := range pending {
		if p.DueAt > endAt {
			t.Fatalf("DueAt=%d > endAt=%d", p.DueAt, endAt)
		}
	}
}

func TestProjectPending_ExcludesInactive(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   fx.now - 35*86400,
		Active:    false,
	})

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 365*86400})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("len = %d, want 0", len(pending))
	}
}

func TestProjectPending_PerRuleCap(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    1,
		Currency:  "EUR",
		Frequency: model.FrequencyDaily,
		StartAt:   fx.now - 86400,
		Active:    true,
	})

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 5*365*86400})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	ruleCount := 0
	for _, p := range pending {
		if p.SourceID == "rule1" {
			ruleCount++
		}
	}
	if ruleCount > maxPeriodsPerRule {
		t.Fatalf("rule1 projections = %d, want <= %d", ruleCount, maxPeriodsPerRule)
	}
}

func TestProjectPending_TotalCap(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	for i := 0; i < 10; i++ {
		fx.insertRule(t, ctx, model.RecurringInvoice{
			ID:        fmt.Sprintf("rule%d", i),
			Amount:    1,
			Currency:  "EUR",
			Frequency: model.FrequencyDaily,
			StartAt:   fx.now - 86400,
			Active:    true,
		})
	}

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 5*365*86400})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if len(pending) > maxTotalPending {
		t.Fatalf("len = %d, want <= %d", len(pending), maxTotalPending)
	}
	if len(pending) != maxTotalPending {
		t.Fatalf("len = %d, want exactly %d (truncation)", len(pending), maxTotalPending)
	}
}

func TestProjectPending_SortedByDueAtAsc(t *testing.T) {
	ctx := context.Background()
	fx := newFixture(t)
	fx.now = time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC).Unix()
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule1",
		Amount:    10,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Active:    true,
	})
	fx.insertRule(t, ctx, model.RecurringInvoice{
		ID:        "rule2",
		Amount:    20,
		Currency:  "EUR",
		Frequency: model.FrequencyMonthly,
		StartAt:   time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC).Unix(),
		Active:    true,
	})
	fx.insertUpcoming(t, ctx, model.UpcomingInvoice{
		ID:       "up1",
		Amount:   5,
		Currency: "EUR",
		DueAt:    fx.now + 10*86400,
	})

	pending, err := ProjectPending(ctx, fx.db, fx.now, Filters{To: fx.now + 120*86400})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if len(pending) < 3 {
		t.Fatalf("len = %d, want >= 3", len(pending))
	}
	for i := 1; i < len(pending); i++ {
		if pending[i-1].DueAt > pending[i].DueAt {
			t.Fatalf("not sorted ASC at %d: %d > %d", i, pending[i-1].DueAt, pending[i].DueAt)
		}
	}
}
