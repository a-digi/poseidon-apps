package storage

import (
	"context"
	"database/sql"
	"reflect"
	"sort"
	"testing"
	"time"

	"mbud-plugin/model"
)

func openTestDB(t *testing.T) *BusinessRepo {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewBusinessRepo(db)
}

func openAllRepos(t *testing.T) (*BusinessRepo, *InvoiceRepo, *RecurringRepo, *UpcomingRepo) {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewBusinessRepo(db), NewInvoiceRepo(db), NewRecurringRepo(db), NewUpcomingRepo(db)
}

func TestRoundTrip_Business(t *testing.T) {
	ctx := context.Background()
	repo := openTestDB(t)

	seed := []model.Business{
		{ID: "b1", Name: "Acme", TaxID: "T-1", Email: "a@x", Address: "addr1", Notes: "n1", CreatedAt: 100, UpdatedAt: 100},
		{ID: "b2", Name: "Beta", TaxID: "T-2", Email: "b@x", Address: "addr2", Notes: "n2", CreatedAt: 200, UpdatedAt: 200},
		{ID: "b3", Name: "Gamma", CreatedAt: 300, UpdatedAt: 300},
	}
	for _, b := range seed {
		if err := repo.Insert(ctx, b); err != nil {
			t.Fatalf("insert %s: %v", b.ID, err)
		}
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("len = %d, want 3", len(list))
	}
	// ORDER BY created_at DESC → b3, b2, b1
	if list[0].ID != "b3" || list[1].ID != "b2" || list[2].ID != "b1" {
		t.Fatalf("order: %v %v %v", list[0].ID, list[1].ID, list[2].ID)
	}

	got, ok, err := repo.Get(ctx, "b1")
	if err != nil || !ok {
		t.Fatalf("get b1: ok=%v err=%v", ok, err)
	}
	if !reflect.DeepEqual(got, seed[0]) {
		t.Fatalf("get mismatch:\n got=%#v\nwant=%#v", got, seed[0])
	}

	updated := seed[0]
	updated.Name = "Acme Renamed"
	updated.UpdatedAt = 999
	ok, err = repo.Update(ctx, updated)
	if err != nil || !ok {
		t.Fatalf("update: ok=%v err=%v", ok, err)
	}
	got, _, _ = repo.Get(ctx, "b1")
	if got.Name != "Acme Renamed" || got.UpdatedAt != 999 || got.CreatedAt != 100 {
		t.Fatalf("after update: %+v", got)
	}

	if _, ok, _ := repo.Get(ctx, "missing"); ok {
		t.Fatal("get missing returned ok=true")
	}
	if ok, _ := repo.Update(ctx, model.Business{ID: "missing", Name: "x"}); ok {
		t.Fatal("update missing returned true")
	}

	if err := repo.Delete(ctx, "b2"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := repo.Delete(ctx, "b2"); err != nil {
		t.Fatalf("delete idempotent: %v", err)
	}
	list, _ = repo.List(ctx)
	if len(list) != 2 {
		t.Fatalf("after delete len = %d, want 2", len(list))
	}
}

func TestRoundTrip_BusinessLogoType(t *testing.T) {
	ctx := context.Background()
	repo := openTestDB(t)

	if err := repo.Insert(ctx, model.Business{ID: "b1", Name: "Acme", LogoType: "image/png", CreatedAt: 100, UpdatedAt: 100}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, ok, err := repo.Get(ctx, "b1")
	if err != nil || !ok {
		t.Fatalf("get: ok=%v err=%v", ok, err)
	}
	if got.LogoType != "image/png" {
		t.Fatalf("LogoType after insert = %q, want image/png", got.LogoType)
	}

	updated, err := repo.UpdateLogoType(ctx, "b1", "image/jpeg", 200)
	if err != nil || !updated {
		t.Fatalf("UpdateLogoType jpeg: ok=%v err=%v", updated, err)
	}
	got, _, _ = repo.Get(ctx, "b1")
	if got.LogoType != "image/jpeg" || got.UpdatedAt != 200 {
		t.Fatalf("after UpdateLogoType jpeg: %+v", got)
	}

	updated, err = repo.UpdateLogoType(ctx, "b1", "", 300)
	if err != nil || !updated {
		t.Fatalf("UpdateLogoType empty: ok=%v err=%v", updated, err)
	}
	got, _, _ = repo.Get(ctx, "b1")
	if got.LogoType != "" || got.UpdatedAt != 300 {
		t.Fatalf("after UpdateLogoType empty: %+v", got)
	}

	list, _ := repo.List(ctx)
	if len(list) != 1 || list[0].LogoType != "" {
		t.Fatalf("list after clear: %+v", list)
	}

	missing, err := repo.UpdateLogoType(ctx, "nope", "image/png", 400)
	if err != nil || missing {
		t.Fatalf("UpdateLogoType missing: ok=%v err=%v", missing, err)
	}
}

func TestRoundTrip_Invoice(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}

	seed := []model.Invoice{
		{ID: "i1", BusinessID: "biz1", Amount: 100.50, Currency: "EUR", Description: "d1", IssuedAt: 10, DueAt: 20, Paid: false, PaidAt: 0, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i2", BusinessID: "biz1", Amount: 250.00, Currency: "USD", IssuedAt: 11, DueAt: 21, Paid: true, PaidAt: 25, CreatedAt: 200, UpdatedAt: 200},
	}
	for _, i := range seed {
		if err := ir.Insert(ctx, i); err != nil {
			t.Fatalf("insert %s: %v", i.ID, err)
		}
	}

	list, err := ir.List(ctx, 0, 0, nil, nil, nil, false, 0, 0, "issuedAt", "desc")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}

	got, ok, _ := ir.Get(ctx, "i2")
	if !ok || !reflect.DeepEqual(got, seed[1]) {
		t.Fatalf("get i2 mismatch:\n got=%#v\nwant=%#v", got, seed[1])
	}

	updated := seed[0]
	updated.Paid = true
	updated.PaidAt = 50
	updated.Amount = 123.45
	updated.UpdatedAt = 999
	ok, err = ir.Update(ctx, updated)
	if err != nil || !ok {
		t.Fatalf("update: ok=%v err=%v", ok, err)
	}
	got, _, _ = ir.Get(ctx, "i1")
	if !got.Paid || got.PaidAt != 50 || got.Amount != 123.45 || got.CreatedAt != 100 {
		t.Fatalf("after update: %+v", got)
	}

	if err := ir.Delete(ctx, "i2"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ = ir.List(ctx, 0, 0, nil, nil, nil, false, 0, 0, "issuedAt", "desc")
	if len(list) != 1 {
		t.Fatalf("after delete len = %d, want 1", len(list))
	}
}

func TestInvoiceList_Filter(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}

	seed := []model.Invoice{
		{ID: "i100", BusinessID: "biz1", Amount: 10, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i200", BusinessID: "biz1", Amount: 20, Currency: "EUR", IssuedAt: 200, DueAt: 250, CreatedAt: 200, UpdatedAt: 200},
		{ID: "i300", BusinessID: "biz1", Amount: 30, Currency: "EUR", IssuedAt: 300, DueAt: 350, CreatedAt: 300, UpdatedAt: 300},
	}
	for _, i := range seed {
		if err := ir.Insert(ctx, i); err != nil {
			t.Fatalf("insert %s: %v", i.ID, err)
		}
	}

	cases := []struct {
		name     string
		from, to int64
		wantIDs  []string
	}{
		{"unbounded", 0, 0, []string{"i300", "i200", "i100"}},
		{"from only", 150, 0, []string{"i300", "i200"}},
		{"to only", 0, 250, []string{"i200", "i100"}},
		{"both bounds", 150, 250, []string{"i200"}},
		{"out of range", 500, 1000, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := ir.List(ctx, tc.from, tc.to, nil, nil, nil, false, 0, 0, "issuedAt", "desc")
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			got := make([]string, len(list))
			for i, inv := range list {
				got[i] = inv.ID
			}
			if !reflect.DeepEqual(got, tc.wantIDs) {
				t.Fatalf("got %v, want %v", got, tc.wantIDs)
			}
		})
	}
}

func TestInvoiceList_LimitOffset(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}

	seed := []model.Invoice{
		{ID: "i100", BusinessID: "biz1", Amount: 10, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i200", BusinessID: "biz1", Amount: 20, Currency: "EUR", IssuedAt: 200, DueAt: 250, CreatedAt: 200, UpdatedAt: 200},
		{ID: "i300", BusinessID: "biz1", Amount: 30, Currency: "EUR", IssuedAt: 300, DueAt: 350, CreatedAt: 300, UpdatedAt: 300},
		{ID: "i400", BusinessID: "biz1", Amount: 40, Currency: "EUR", IssuedAt: 400, DueAt: 450, CreatedAt: 400, UpdatedAt: 400},
		{ID: "i500", BusinessID: "biz1", Amount: 50, Currency: "EUR", IssuedAt: 500, DueAt: 550, CreatedAt: 500, UpdatedAt: 500},
	}
	for _, i := range seed {
		if err := ir.Insert(ctx, i); err != nil {
			t.Fatalf("insert %s: %v", i.ID, err)
		}
	}

	cases := []struct {
		name          string
		limit, offset int64
		wantIDs       []string
	}{
		{"page 1", 2, 0, []string{"i500", "i400"}},
		{"page 2", 2, 2, []string{"i300", "i200"}},
		{"page 3 partial", 2, 4, []string{"i100"}},
		{"limit larger than total", 10, 0, []string{"i500", "i400", "i300", "i200", "i100"}},
		{"no limit ignores offset", 0, 0, []string{"i500", "i400", "i300", "i200", "i100"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := ir.List(ctx, 0, 0, nil, nil, nil, false, tc.limit, tc.offset, "issuedAt", "desc")
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			got := make([]string, len(list))
			for i, inv := range list {
				got[i] = inv.ID
			}
			if !reflect.DeepEqual(got, tc.wantIDs) {
				t.Fatalf("got %v, want %v", got, tc.wantIDs)
			}
		})
	}
}

func TestInvoiceList_Sort(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}

	seed := []model.Invoice{
		{ID: "A", BusinessID: "biz1", Amount: 50, Currency: "EUR", IssuedAt: 100, DueAt: 1000, CreatedAt: 1, UpdatedAt: 1},
		{ID: "B", BusinessID: "biz1", Amount: 20, Currency: "EUR", IssuedAt: 300, DueAt: 800, CreatedAt: 2, UpdatedAt: 2},
		{ID: "C", BusinessID: "biz1", Amount: 70, Currency: "EUR", IssuedAt: 200, DueAt: 1200, CreatedAt: 3, UpdatedAt: 3},
		{ID: "D", BusinessID: "biz1", Amount: 30, Currency: "EUR", IssuedAt: 500, DueAt: 900, CreatedAt: 4, UpdatedAt: 4},
		{ID: "E", BusinessID: "biz1", Amount: 10, Currency: "EUR", IssuedAt: 400, DueAt: 1100, CreatedAt: 5, UpdatedAt: 5},
	}
	for _, inv := range seed {
		if err := ir.Insert(ctx, inv); err != nil {
			t.Fatalf("insert %s: %v", inv.ID, err)
		}
	}

	type pick func(model.Invoice) int64
	pickIssued := func(i model.Invoice) int64 { return i.IssuedAt }
	pickDue := func(i model.Invoice) int64 { return i.DueAt }
	pickAmount := func(i model.Invoice) int64 { return int64(i.Amount) }

	cases := []struct {
		name    string
		sortBy  string
		sortDir string
		pick    pick
		want    []int64
	}{
		{"issuedAt desc", "issuedAt", "desc", pickIssued, []int64{500, 400, 300, 200, 100}},
		{"issuedAt asc", "issuedAt", "asc", pickIssued, []int64{100, 200, 300, 400, 500}},
		{"dueAt asc", "dueAt", "asc", pickDue, []int64{800, 900, 1000, 1100, 1200}},
		{"dueAt desc", "dueAt", "desc", pickDue, []int64{1200, 1100, 1000, 900, 800}},
		{"amount desc", "amount", "desc", pickAmount, []int64{70, 50, 30, 20, 10}},
		{"amount asc", "amount", "asc", pickAmount, []int64{10, 20, 30, 50, 70}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := ir.List(ctx, 0, 0, nil, nil, nil, false, 0, 0, tc.sortBy, tc.sortDir)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			got := make([]int64, len(list))
			for i, inv := range list {
				got[i] = tc.pick(inv)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}

	t.Run("filter and sort and limit", func(t *testing.T) {
		list, err := ir.List(ctx, 0, 350, nil, nil, nil, false, 2, 0, "issuedAt", "asc")
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		got := make([]int64, len(list))
		for i, inv := range list {
			got[i] = inv.IssuedAt
		}
		want := []int64{100, 200}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}

func TestInvoiceCount(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}

	seed := []model.Invoice{
		{ID: "i100", BusinessID: "biz1", Amount: 10, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i200", BusinessID: "biz1", Amount: 20, Currency: "EUR", IssuedAt: 200, DueAt: 250, CreatedAt: 200, UpdatedAt: 200},
		{ID: "i300", BusinessID: "biz1", Amount: 30, Currency: "EUR", IssuedAt: 300, DueAt: 350, CreatedAt: 300, UpdatedAt: 300},
		{ID: "i400", BusinessID: "biz1", Amount: 40, Currency: "EUR", IssuedAt: 400, DueAt: 450, CreatedAt: 400, UpdatedAt: 400},
		{ID: "i500", BusinessID: "biz1", Amount: 50, Currency: "EUR", IssuedAt: 500, DueAt: 550, CreatedAt: 500, UpdatedAt: 500},
	}
	for _, i := range seed {
		if err := ir.Insert(ctx, i); err != nil {
			t.Fatalf("insert %s: %v", i.ID, err)
		}
	}

	cases := []struct {
		name     string
		from, to int64
		want     int
	}{
		{"unbounded", 0, 0, 5},
		{"from only", 250, 0, 3},
		{"to only", 0, 350, 3},
		{"both bounds", 150, 450, 3},
		{"out of range", 1000, 2000, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n, err := ir.Count(ctx, tc.from, tc.to, nil, nil, nil, false)
			if err != nil {
				t.Fatalf("count: %v", err)
			}
			if n != tc.want {
				t.Fatalf("got %d, want %d", n, tc.want)
			}
		})
	}
}

func TestRoundTrip_Recurring(t *testing.T) {
	ctx := context.Background()
	br, _, rr, _ := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}

	seed := []model.RecurringInvoice{
		{ID: "r1", BusinessID: "biz1", Amount: 50, Currency: "EUR", Frequency: model.FrequencyMonthly, StartAt: 100, Active: true, IssueDayOfMonth: 15, IssueDayOfWeek: 3, IssueMonthOfYear: 8, CreatedAt: 100, UpdatedAt: 100},
		{ID: "r2", BusinessID: "biz1", Amount: 75, Currency: "USD", Frequency: model.FrequencyYearly, StartAt: 200, EndAt: 1000, Active: false, CreatedAt: 200, UpdatedAt: 200},
	}
	for _, r := range seed {
		if err := rr.Insert(ctx, r); err != nil {
			t.Fatalf("insert %s: %v", r.ID, err)
		}
	}

	list, _ := rr.List(ctx)
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}

	got, ok, _ := rr.Get(ctx, "r1")
	if !ok || !reflect.DeepEqual(got, seed[0]) {
		t.Fatalf("get r1 mismatch:\n got=%#v\nwant=%#v", got, seed[0])
	}
	if got.IssueDayOfMonth != 15 || got.IssueDayOfWeek != 3 || got.IssueMonthOfYear != 8 {
		t.Fatalf("anchor fields after insert: %+v", got)
	}

	updated := seed[1]
	updated.Frequency = model.FrequencyWeekly
	updated.Active = true
	updated.IssueDayOfWeek = 5
	updated.IssueDayOfMonth = 22
	updated.IssueMonthOfYear = 11
	updated.UpdatedAt = 999
	ok, _ = rr.Update(ctx, updated)
	if !ok {
		t.Fatal("update returned false")
	}
	got, _, _ = rr.Get(ctx, "r2")
	if got.Frequency != model.FrequencyWeekly || !got.Active {
		t.Fatalf("after update: %+v", got)
	}
	if got.IssueDayOfWeek != 5 || got.IssueDayOfMonth != 22 || got.IssueMonthOfYear != 11 {
		t.Fatalf("anchor fields after update: %+v", got)
	}
	for _, ri := range mustListRecurring(t, rr, ctx) {
		if ri.ID == "r2" && (ri.IssueDayOfWeek != 5 || ri.IssueDayOfMonth != 22 || ri.IssueMonthOfYear != 11) {
			t.Fatalf("list r2 anchors: %+v", ri)
		}
	}

	updated.IssueDayOfWeek = 0
	updated.IssueDayOfMonth = 0
	updated.IssueMonthOfYear = 0
	updated.UpdatedAt = 1000
	ok, _ = rr.Update(ctx, updated)
	if !ok {
		t.Fatal("update clear returned false")
	}
	got, _, _ = rr.Get(ctx, "r2")
	if got.IssueDayOfWeek != 0 || got.IssueDayOfMonth != 0 || got.IssueMonthOfYear != 0 {
		t.Fatalf("anchors after clear: %+v", got)
	}
	for _, ri := range mustListRecurring(t, rr, ctx) {
		if ri.ID == "r2" && (ri.IssueDayOfWeek != 0 || ri.IssueDayOfMonth != 0 || ri.IssueMonthOfYear != 0) {
			t.Fatalf("list r2 after clear: %+v", ri)
		}
	}

	if err := rr.Delete(ctx, "r1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ = rr.List(ctx)
	if len(list) != 1 {
		t.Fatalf("after delete len = %d, want 1", len(list))
	}
}

func mustListRecurring(t *testing.T, rr *RecurringRepo, ctx context.Context) []model.RecurringInvoice {
	t.Helper()
	list, err := rr.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	return list
}

func TestRoundTrip_Upcoming(t *testing.T) {
	ctx := context.Background()
	br, _, _, ur := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}

	seed := []model.UpcomingInvoice{
		{ID: "u1", BusinessID: "biz1", Amount: 10, Currency: "EUR", Description: "d", DueAt: 500, CreatedAt: 100, UpdatedAt: 100},
		{ID: "u2", BusinessID: "biz1", Amount: 20, Currency: "USD", DueAt: 600, CreatedAt: 200, UpdatedAt: 200},
	}
	for _, u := range seed {
		if err := ur.Insert(ctx, u); err != nil {
			t.Fatalf("insert %s: %v", u.ID, err)
		}
	}

	list, _ := ur.List(ctx)
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}

	got, ok, _ := ur.Get(ctx, "u1")
	if !ok || !reflect.DeepEqual(got, seed[0]) {
		t.Fatalf("get u1 mismatch:\n got=%#v\nwant=%#v", got, seed[0])
	}

	updated := seed[0]
	updated.Amount = 999.99
	updated.UpdatedAt = 999
	ok, _ = ur.Update(ctx, updated)
	if !ok {
		t.Fatal("update returned false")
	}
	got, _, _ = ur.Get(ctx, "u1")
	if got.Amount != 999.99 || got.CreatedAt != 100 {
		t.Fatalf("after update: %+v", got)
	}

	if err := ur.Delete(ctx, "u2"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ = ur.List(ctx)
	if len(list) != 1 {
		t.Fatalf("after delete len = %d, want 1", len(list))
	}
}

func openLinksDB(t *testing.T) (*sql.DB, *BusinessRepo, *InvoiceRepo, *RecurringRepo, *LinksRepo) {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, NewBusinessRepo(db), NewInvoiceRepo(db), NewRecurringRepo(db), NewLinksRepo(db)
}

func TestRoundTrip_Links(t *testing.T) {
	ctx := context.Background()
	_, br, _, rr, lr := openLinksDB(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	ruleID := "rule1"
	if err := rr.Insert(ctx, model.RecurringInvoice{ID: ruleID, BusinessID: "biz", Amount: 1, Currency: "EUR", Frequency: model.FrequencyMonthly, StartAt: 1, Active: true, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("rec: %v", err)
	}

	for n, invID := range []string{"i0", "i1", "i2"} {
		if err := lr.Insert(ctx, invID, ruleID, n); err != nil {
			t.Fatalf("link insert %s: %v", invID, err)
		}
	}

	maxIdx, err := lr.MaxPeriodIndex(ctx, ruleID)
	if err != nil {
		t.Fatalf("max: %v", err)
	}
	if maxIdx != 2 {
		t.Fatalf("MaxPeriodIndex = %d, want 2", maxIdx)
	}

	ids, err := lr.InvoiceIDsByRecurring(ctx, ruleID)
	if err != nil {
		t.Fatalf("ids: %v", err)
	}
	sort.Strings(ids)
	if !reflect.DeepEqual(ids, []string{"i0", "i1", "i2"}) {
		t.Fatalf("InvoiceIDsByRecurring = %v, want [i0 i1 i2]", ids)
	}

	recIDs, err := lr.RecurringIDsByInvoice(ctx, "i1")
	if err != nil {
		t.Fatalf("recIDs: %v", err)
	}
	if !reflect.DeepEqual(recIDs, []string{ruleID}) {
		t.Fatalf("RecurringIDsByInvoice(i1) = %v, want [%s]", recIDs, ruleID)
	}

	empty, err := lr.RecurringIDsBatch(ctx, nil)
	if err != nil {
		t.Fatalf("batch nil: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("RecurringIDsBatch(nil) len = %d, want 0", len(empty))
	}

	batch, err := lr.RecurringIDsBatch(ctx, []string{"i0", "i1"})
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if len(batch) != 2 {
		t.Fatalf("RecurringIDsBatch len = %d, want 2", len(batch))
	}
	for _, k := range []string{"i0", "i1"} {
		if !reflect.DeepEqual(batch[k], []string{ruleID}) {
			t.Fatalf("batch[%s] = %v, want [%s]", k, batch[k], ruleID)
		}
	}

	unknown, err := lr.RecurringIDsBatch(ctx, []string{"unknown"})
	if err != nil {
		t.Fatalf("batch unknown: %v", err)
	}
	if len(unknown) != 0 {
		t.Fatalf("RecurringIDsBatch([unknown]) len = %d, want 0", len(unknown))
	}
}

func TestNoCascade_LinksOnInvoiceDelete(t *testing.T) {
	ctx := context.Background()
	db, br, ir, rr, lr := openLinksDB(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	ruleID := "rule1"
	if err := rr.Insert(ctx, model.RecurringInvoice{ID: ruleID, BusinessID: "biz", Amount: 1, Currency: "EUR", Frequency: model.FrequencyMonthly, StartAt: 1, Active: true, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("rec: %v", err)
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "real-i", BusinessID: "biz", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("inv: %v", err)
	}
	if err := lr.Insert(ctx, "real-i", ruleID, 0); err != nil {
		t.Fatalf("link: %v", err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM invoices WHERE id = ?`, "real-i"); err != nil {
		t.Fatalf("delete invoice: %v", err)
	}

	maxIdx, err := lr.MaxPeriodIndex(ctx, ruleID)
	if err != nil {
		t.Fatalf("max: %v", err)
	}
	if maxIdx != 0 {
		t.Fatalf("MaxPeriodIndex after invoice delete = %d, want 0 (link survives)", maxIdx)
	}
	ids, err := lr.InvoiceIDsByRecurring(ctx, ruleID)
	if err != nil {
		t.Fatalf("ids: %v", err)
	}
	if !reflect.DeepEqual(ids, []string{"real-i"}) {
		t.Fatalf("InvoiceIDsByRecurring after invoice delete = %v, want [real-i]", ids)
	}
}

func TestCascade_LinksOnRecurringDelete(t *testing.T) {
	ctx := context.Background()
	db, br, _, rr, lr := openLinksDB(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	ruleID := "rule1"
	if err := rr.Insert(ctx, model.RecurringInvoice{ID: ruleID, BusinessID: "biz", Amount: 1, Currency: "EUR", Frequency: model.FrequencyMonthly, StartAt: 1, Active: true, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("rec: %v", err)
	}
	if err := lr.Insert(ctx, "i0", ruleID, 0); err != nil {
		t.Fatalf("link 0: %v", err)
	}
	if err := lr.Insert(ctx, "i1", ruleID, 1); err != nil {
		t.Fatalf("link 1: %v", err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM recurring_invoices WHERE id = ?`, ruleID); err != nil {
		t.Fatalf("delete recurring: %v", err)
	}

	ids, err := lr.InvoiceIDsByRecurring(ctx, ruleID)
	if err != nil {
		t.Fatalf("ids: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("links after cascade = %v, want []", ids)
	}
	maxIdx, err := lr.MaxPeriodIndex(ctx, ruleID)
	if err != nil {
		t.Fatalf("max: %v", err)
	}
	if maxIdx != -1 {
		t.Fatalf("MaxPeriodIndex after cascade = %d, want -1", maxIdx)
	}
}

func seedStatsInvoices(t *testing.T, ctx context.Context, br *BusinessRepo, ir *InvoiceRepo) (string, int64, int64, int64) {
	t.Helper()
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC).Unix()
	jan15Plus6h := time.Date(2026, 1, 15, 6, 0, 0, 0, time.UTC).Unix()
	feb10 := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC).Unix()
	jan20 := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC).Unix()
	jan20Plus12h := time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC).Unix()
	mar1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC).Unix()

	seed := []model.Invoice{
		{ID: "e1", BusinessID: "biz1", Amount: 100, Currency: "EUR", IssuedAt: jan15, DueAt: jan15 + 86400, Paid: true, PaidAt: jan15, CreatedAt: 100, UpdatedAt: 100},
		{ID: "e2", BusinessID: "biz1", Amount: 50, Currency: "EUR", IssuedAt: jan15Plus6h, DueAt: jan15Plus6h + 86400, Paid: false, CreatedAt: 110, UpdatedAt: 110},
		{ID: "e3", BusinessID: "biz1", Amount: 200, Currency: "EUR", IssuedAt: feb10, DueAt: feb10 + 86400, Paid: false, CreatedAt: 120, UpdatedAt: 120},
		{ID: "u1", BusinessID: "biz1", Amount: 30, Currency: "USD", IssuedAt: jan20, DueAt: jan20 + 86400, Paid: true, PaidAt: jan20, CreatedAt: 130, UpdatedAt: 130},
		{ID: "u2", BusinessID: "biz1", Amount: 70, Currency: "USD", IssuedAt: jan20Plus12h, DueAt: jan20Plus12h + 86400, Paid: false, CreatedAt: 140, UpdatedAt: 140},
		{ID: "e4", BusinessID: "biz1", Amount: 999, Currency: "EUR", IssuedAt: mar1, DueAt: mar1 + 86400, Paid: true, PaidAt: mar1, CreatedAt: 150, UpdatedAt: 150},
	}
	for _, i := range seed {
		if err := ir.Insert(ctx, i); err != nil {
			t.Fatalf("insert %s: %v", i.ID, err)
		}
	}
	jan20Midnight := jan20
	mar1Midnight := mar1
	feb1 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC).Unix()
	return "biz1", jan20Midnight, mar1Midnight, feb1
}

func TestInvoiceStats(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	bizID, jan20Midnight, mar1Midnight, _ := seedStatsInvoices(t, ctx, br, ir)

	out, err := ir.Stats(ctx, 0, 0, nil, nil, nil)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}

	eur := out[0]
	if eur.Currency != "EUR" {
		t.Fatalf("stats[0].Currency = %q, want EUR", eur.Currency)
	}
	if eur.Total != 1349 || eur.Count != 4 {
		t.Fatalf("EUR total/count = %v/%d, want 1349/4", eur.Total, eur.Count)
	}
	if eur.PaidAmount != 1099 || eur.PaidCount != 2 {
		t.Fatalf("EUR paid = %v/%d, want 1099/2", eur.PaidAmount, eur.PaidCount)
	}
	if eur.UnpaidAmount != 250 || eur.UnpaidCount != 2 {
		t.Fatalf("EUR unpaid = %v/%d, want 250/2", eur.UnpaidAmount, eur.UnpaidCount)
	}
	if eur.Average != 337.25 {
		t.Fatalf("EUR average = %v, want 337.25", eur.Average)
	}
	if eur.MaxAmount != 999 {
		t.Fatalf("EUR max = %v, want 999", eur.MaxAmount)
	}
	if eur.BusinessCount != 1 {
		t.Fatalf("EUR businessCount = %d, want 1", eur.BusinessCount)
	}
	if len(eur.TopBusinesses) != 1 {
		t.Fatalf("EUR topBusinesses len = %d, want 1", len(eur.TopBusinesses))
	}
	if eur.TopBusinesses[0].BusinessID != bizID || eur.TopBusinesses[0].Amount != 1349 {
		t.Fatalf("EUR top biz = %q/%v, want %q/1349", eur.TopBusinesses[0].BusinessID, eur.TopBusinesses[0].Amount, bizID)
	}
	if eur.TopDayEpoch != mar1Midnight || eur.TopDayAmount != 999 {
		t.Fatalf("EUR top day = %d/%v, want %d/999", eur.TopDayEpoch, eur.TopDayAmount, mar1Midnight)
	}

	usd := out[1]
	if usd.Currency != "USD" {
		t.Fatalf("stats[1].Currency = %q, want USD", usd.Currency)
	}
	if usd.Total != 100 || usd.Count != 2 {
		t.Fatalf("USD total/count = %v/%d, want 100/2", usd.Total, usd.Count)
	}
	if usd.PaidAmount != 30 || usd.PaidCount != 1 {
		t.Fatalf("USD paid = %v/%d, want 30/1", usd.PaidAmount, usd.PaidCount)
	}
	if usd.UnpaidAmount != 70 || usd.UnpaidCount != 1 {
		t.Fatalf("USD unpaid = %v/%d, want 70/1", usd.UnpaidAmount, usd.UnpaidCount)
	}
	if usd.Average != 50 {
		t.Fatalf("USD average = %v, want 50", usd.Average)
	}
	if usd.MaxAmount != 70 {
		t.Fatalf("USD max = %v, want 70", usd.MaxAmount)
	}
	if usd.BusinessCount != 1 {
		t.Fatalf("USD businessCount = %d, want 1", usd.BusinessCount)
	}
	if usd.TopDayEpoch != jan20Midnight || usd.TopDayAmount != 100 {
		t.Fatalf("USD top day = %d/%v, want %d/100", usd.TopDayEpoch, usd.TopDayAmount, jan20Midnight)
	}
}

func TestInvoiceStats_FilterScope(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	_, _, _, feb1 := seedStatsInvoices(t, ctx, br, ir)

	out, err := ir.Stats(ctx, feb1, 0, nil, nil, nil)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1 (USD must be excluded)", len(out))
	}
	if out[0].Currency != "EUR" {
		t.Fatalf("Currency = %q, want EUR", out[0].Currency)
	}
	if out[0].Total != 1199 || out[0].Count != 2 {
		t.Fatalf("EUR total/count = %v/%d, want 1199/2", out[0].Total, out[0].Count)
	}
}

func TestInvoiceStats_Empty(t *testing.T) {
	ctx := context.Background()
	_, ir, _, _ := openAllRepos(t)

	out, err := ir.Stats(ctx, 0, 0, nil, nil, nil)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if out == nil {
		t.Fatal("stats on empty DB returned nil, want empty non-nil slice")
	}
	if len(out) != 0 {
		t.Fatalf("len = %d, want 0", len(out))
	}

	br, ir, _, _ := openAllRepos(t)
	seedStatsInvoices(t, ctx, br, ir)
	out, err = ir.Stats(ctx, time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix(), 0, nil, nil, nil)
	if err != nil {
		t.Fatalf("stats filter-miss: %v", err)
	}
	if out == nil {
		t.Fatal("stats with no-match filter returned nil, want empty non-nil slice")
	}
	if len(out) != 0 {
		t.Fatalf("filter-miss len = %d, want 0", len(out))
	}
}

func TestInvoiceStats_TopBusinessesTop5(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)

	issued := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC).Unix()
	for i := 1; i <= 7; i++ {
		bizID := "biz" + string(rune('0'+i))
		if err := br.Insert(ctx, model.Business{ID: bizID, Name: bizID, CreatedAt: int64(i), UpdatedAt: int64(i)}); err != nil {
			t.Fatalf("biz insert %s: %v", bizID, err)
		}
		if err := ir.Insert(ctx, model.Invoice{
			ID:         "inv" + string(rune('0'+i)),
			BusinessID: bizID,
			Amount:     float64(i * 100),
			Currency:   "EUR",
			IssuedAt:   issued,
			DueAt:      issued + 86400,
			CreatedAt:  int64(1000 + i),
			UpdatedAt:  int64(1000 + i),
		}); err != nil {
			t.Fatalf("inv insert %d: %v", i, err)
		}
	}

	stats, err := ir.Stats(ctx, 0, 0, nil, nil, nil)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("currencies = %d, want 1", len(stats))
	}
	if stats[0].BusinessCount != 7 {
		t.Fatalf("BusinessCount = %d, want 7", stats[0].BusinessCount)
	}
	if len(stats[0].TopBusinesses) != 5 {
		t.Fatalf("TopBusinesses len = %d, want 5", len(stats[0].TopBusinesses))
	}
	wantOrder := []struct {
		id     string
		amount float64
	}{
		{"biz7", 700},
		{"biz6", 600},
		{"biz5", 500},
		{"biz4", 400},
		{"biz3", 300},
	}
	for i, w := range wantOrder {
		got := stats[0].TopBusinesses[i]
		if got.BusinessID != w.id || got.Amount != w.amount {
			t.Fatalf("TopBusinesses[%d] = %q/%v, want %q/%v", i, got.BusinessID, got.Amount, w.id, w.amount)
		}
	}
	for _, tb := range stats[0].TopBusinesses {
		if tb.BusinessID == "biz1" || tb.BusinessID == "biz2" {
			t.Fatalf("biz1/biz2 must not be in top 5, got %q", tb.BusinessID)
		}
	}
}

func TestInvoiceStats_TopBusinessesTieBreak(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)

	issued := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC).Unix()
	created := []int64{1000, 2000, 3000}
	for i, c := range created {
		bizID := "biz" + string(rune('0'+i+1))
		if err := br.Insert(ctx, model.Business{ID: bizID, Name: bizID, CreatedAt: 1, UpdatedAt: 1}); err != nil {
			t.Fatalf("biz insert %s: %v", bizID, err)
		}
		if err := ir.Insert(ctx, model.Invoice{
			ID:         "inv" + string(rune('0'+i+1)),
			BusinessID: bizID,
			Amount:     100,
			Currency:   "EUR",
			IssuedAt:   issued,
			DueAt:      issued + 86400,
			CreatedAt:  c,
			UpdatedAt:  c,
		}); err != nil {
			t.Fatalf("inv insert %d: %v", i, err)
		}
	}

	stats, err := ir.Stats(ctx, 0, 0, nil, nil, nil)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("currencies = %d, want 1", len(stats))
	}
	if len(stats[0].TopBusinesses) != 3 {
		t.Fatalf("TopBusinesses len = %d, want 3", len(stats[0].TopBusinesses))
	}
	wantIDs := []string{"biz3", "biz2", "biz1"}
	for i, w := range wantIDs {
		got := stats[0].TopBusinesses[i]
		if got.BusinessID != w || got.Amount != 100 {
			t.Fatalf("TopBusinesses[%d] = %q/%v, want %q/100", i, got.BusinessID, got.Amount, w)
		}
	}
}

func openUpcomingLinksDB(t *testing.T) (*sql.DB, *BusinessRepo, *InvoiceRepo, *UpcomingRepo, *UpcomingLinksRepo) {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, NewBusinessRepo(db), NewInvoiceRepo(db), NewUpcomingRepo(db), NewUpcomingLinksRepo(db)
}

func TestRoundTrip_UpcomingLinks(t *testing.T) {
	ctx := context.Background()
	_, br, _, ur, ulr := openUpcomingLinksDB(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	upcomingID := "u1"
	if err := ur.Insert(ctx, model.UpcomingInvoice{ID: upcomingID, BusinessID: "biz", Amount: 1, Currency: "EUR", DueAt: 100, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("upcoming: %v", err)
	}

	for _, invID := range []string{"i0", "i1"} {
		if err := ulr.Insert(ctx, invID, upcomingID); err != nil {
			t.Fatalf("link insert %s: %v", invID, err)
		}
	}

	ids, err := ulr.InvoiceIDsByUpcoming(ctx, upcomingID)
	if err != nil {
		t.Fatalf("ids: %v", err)
	}
	sort.Strings(ids)
	if !reflect.DeepEqual(ids, []string{"i0", "i1"}) {
		t.Fatalf("InvoiceIDsByUpcoming = %v, want [i0 i1]", ids)
	}

	upIDs, err := ulr.UpcomingIDsByInvoice(ctx, "i0")
	if err != nil {
		t.Fatalf("upIDs: %v", err)
	}
	if !reflect.DeepEqual(upIDs, []string{upcomingID}) {
		t.Fatalf("UpcomingIDsByInvoice(i0) = %v, want [%s]", upIDs, upcomingID)
	}

	empty, err := ulr.UpcomingIDsBatch(ctx, []string{})
	if err != nil {
		t.Fatalf("batch empty: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("UpcomingIDsBatch([]) len = %d, want 0", len(empty))
	}

	emptyNil, err := ulr.UpcomingIDsBatch(ctx, nil)
	if err != nil {
		t.Fatalf("batch nil: %v", err)
	}
	if len(emptyNil) != 0 {
		t.Fatalf("UpcomingIDsBatch(nil) len = %d, want 0", len(emptyNil))
	}

	batch, err := ulr.UpcomingIDsBatch(ctx, []string{"i0", "i1"})
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if len(batch) != 2 {
		t.Fatalf("UpcomingIDsBatch len = %d, want 2", len(batch))
	}
	for _, k := range []string{"i0", "i1"} {
		if !reflect.DeepEqual(batch[k], []string{upcomingID}) {
			t.Fatalf("batch[%s] = %v, want [%s]", k, batch[k], upcomingID)
		}
	}

	unknown, err := ulr.UpcomingIDsBatch(ctx, []string{"unknown"})
	if err != nil {
		t.Fatalf("batch unknown: %v", err)
	}
	if len(unknown) != 0 {
		t.Fatalf("UpcomingIDsBatch([unknown]) len = %d, want 0", len(unknown))
	}

	if err := ulr.DeleteByUpcoming(ctx, upcomingID); err != nil {
		t.Fatalf("delete by upcoming: %v", err)
	}
	ids, err = ulr.InvoiceIDsByUpcoming(ctx, upcomingID)
	if err != nil {
		t.Fatalf("ids after delete: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("InvoiceIDsByUpcoming after DeleteByUpcoming = %v, want []", ids)
	}
}

func TestNoCascade_LinksOnInvoiceDelete_Upcoming(t *testing.T) {
	ctx := context.Background()
	db, br, ir, ur, ulr := openUpcomingLinksDB(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	upcomingID := "u1"
	if err := ur.Insert(ctx, model.UpcomingInvoice{ID: upcomingID, BusinessID: "biz", Amount: 1, Currency: "EUR", DueAt: 100, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("upcoming: %v", err)
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "real-i", BusinessID: "biz", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("inv: %v", err)
	}
	if err := ulr.Insert(ctx, "real-i", upcomingID); err != nil {
		t.Fatalf("link: %v", err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM invoices WHERE id = ?`, "real-i"); err != nil {
		t.Fatalf("delete invoice: %v", err)
	}

	ids, err := ulr.InvoiceIDsByUpcoming(ctx, upcomingID)
	if err != nil {
		t.Fatalf("ids: %v", err)
	}
	if !reflect.DeepEqual(ids, []string{"real-i"}) {
		t.Fatalf("InvoiceIDsByUpcoming after invoice delete = %v, want [real-i]", ids)
	}
}

func TestCascade_LinksOnUpcomingDelete(t *testing.T) {
	ctx := context.Background()
	db, br, _, ur, ulr := openUpcomingLinksDB(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	upcomingID := "u1"
	if err := ur.Insert(ctx, model.UpcomingInvoice{ID: upcomingID, BusinessID: "biz", Amount: 1, Currency: "EUR", DueAt: 100, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("upcoming: %v", err)
	}
	if err := ulr.Insert(ctx, "i0", upcomingID); err != nil {
		t.Fatalf("link 0: %v", err)
	}
	if err := ulr.Insert(ctx, "i1", upcomingID); err != nil {
		t.Fatalf("link 1: %v", err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM upcoming_invoices WHERE id = ?`, upcomingID); err != nil {
		t.Fatalf("delete upcoming: %v", err)
	}

	ids, err := ulr.InvoiceIDsByUpcoming(ctx, upcomingID)
	if err != nil {
		t.Fatalf("ids: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("links after cascade = %v, want []", ids)
	}
}

func countRows(t *testing.T, ctx context.Context, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func TestCascade_PlainInvoiceDelete(t *testing.T) {
	ctx := context.Background()
	db, br, ir, _, _ := openLinksDB(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "inv", BusinessID: "biz", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("inv: %v", err)
	}

	if err := ir.Delete(ctx, "inv"); err != nil {
		t.Fatalf("delete invoice: %v", err)
	}

	if got := countRows(t, ctx, db, "invoices"); got != 0 {
		t.Fatalf("invoices = %d, want 0", got)
	}
	if got := countRows(t, ctx, db, "recurring_invoice_links"); got != 0 {
		t.Fatalf("recurring_invoice_links = %d, want 0", got)
	}
	if got := countRows(t, ctx, db, "upcoming_invoice_links"); got != 0 {
		t.Fatalf("upcoming_invoice_links = %d, want 0", got)
	}
	if got := countRows(t, ctx, db, "businesses"); got != 1 {
		t.Fatalf("businesses = %d, want 1", got)
	}
}

func TestCascade_BusinessDeleteWipesEverything(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	br := NewBusinessRepo(db)
	ir := NewInvoiceRepo(db)
	rr := NewRecurringRepo(db)
	ur := NewUpcomingRepo(db)

	if err := br.Insert(ctx, model.Business{ID: "B1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	if err := rr.Insert(ctx, model.RecurringInvoice{ID: "R", BusinessID: "B1", Amount: 1, Currency: "EUR", Frequency: model.FrequencyMonthly, StartAt: 1, Active: true, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("rec: %v", err)
	}
	if err := ur.Insert(ctx, model.UpcomingInvoice{ID: "U", BusinessID: "B1", Amount: 1, Currency: "EUR", DueAt: 1, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("up: %v", err)
	}

	invoices := []model.Invoice{
		{ID: "I_plain", BusinessID: "B1", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 1, UpdatedAt: 1},
		{ID: "I_R", BusinessID: "B1", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 2, UpdatedAt: 2},
		{ID: "I_U", BusinessID: "B1", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 3, UpdatedAt: 3},
	}
	for _, inv := range invoices {
		if err := ir.Insert(ctx, inv); err != nil {
			t.Fatalf("inv %s: %v", inv.ID, err)
		}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO recurring_invoice_links (invoice_id, recurring_id, period_index) VALUES (?, ?, ?)`, "I_R", "R", 0); err != nil {
		t.Fatalf("rec link: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO upcoming_invoice_links (invoice_id, upcoming_id) VALUES (?, ?)`, "I_U", "U"); err != nil {
		t.Fatalf("up link: %v", err)
	}

	if err := br.Delete(ctx, "B1"); err != nil {
		t.Fatalf("delete biz: %v", err)
	}

	for _, table := range []string{"businesses", "invoices", "recurring_invoices", "recurring_invoice_links", "upcoming_invoices", "upcoming_invoice_links"} {
		if got := countRows(t, ctx, db, table); got != 0 {
			t.Fatalf("%s = %d, want 0", table, got)
		}
	}
}

func TestMaterialisedInvoiceUserDeletePreservesLink_Recurring(t *testing.T) {
	ctx := context.Background()
	db, br, ir, rr, lr := openLinksDB(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	const ruleID = "rule1"
	if err := rr.Insert(ctx, model.RecurringInvoice{ID: ruleID, BusinessID: "biz", Amount: 1, Currency: "EUR", Frequency: model.FrequencyMonthly, StartAt: 1, Active: true, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("rec: %v", err)
	}
	const invoiceID = "inv1"
	if err := ir.Insert(ctx, model.Invoice{ID: invoiceID, BusinessID: "biz", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("inv: %v", err)
	}
	if err := lr.Insert(ctx, invoiceID, ruleID, 0); err != nil {
		t.Fatalf("link: %v", err)
	}

	if err := ir.Delete(ctx, invoiceID); err != nil {
		t.Fatalf("delete invoice: %v", err)
	}

	if got := countRows(t, ctx, db, "invoices"); got != 0 {
		t.Fatalf("invoices = %d, want 0", got)
	}
	var linkN int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM recurring_invoice_links WHERE recurring_id = ?`, ruleID).Scan(&linkN); err != nil {
		t.Fatalf("count links: %v", err)
	}
	if linkN != 1 {
		t.Fatalf("recurring_invoice_links = %d, want 1 (orphan preserved)", linkN)
	}
	maxIdx, err := lr.MaxPeriodIndex(ctx, ruleID)
	if err != nil {
		t.Fatalf("max: %v", err)
	}
	if maxIdx != 0 {
		t.Fatalf("MaxPeriodIndex = %d, want 0", maxIdx)
	}
}

func TestMaterialisedInvoiceUserDeletePreservesLink_Upcoming(t *testing.T) {
	ctx := context.Background()
	db, br, ir, ur, ulr := openUpcomingLinksDB(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	const upcomingID = "up1"
	if err := ur.Insert(ctx, model.UpcomingInvoice{ID: upcomingID, BusinessID: "biz", Amount: 1, Currency: "EUR", DueAt: 100, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("upcoming: %v", err)
	}
	const invoiceID = "inv1"
	if err := ir.Insert(ctx, model.Invoice{ID: invoiceID, BusinessID: "biz", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("inv: %v", err)
	}
	if err := ulr.Insert(ctx, invoiceID, upcomingID); err != nil {
		t.Fatalf("link: %v", err)
	}

	if err := ir.Delete(ctx, invoiceID); err != nil {
		t.Fatalf("delete invoice: %v", err)
	}

	if got := countRows(t, ctx, db, "invoices"); got != 0 {
		t.Fatalf("invoices = %d, want 0", got)
	}
	var linkN int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM upcoming_invoice_links WHERE upcoming_id = ?`, upcomingID).Scan(&linkN); err != nil {
		t.Fatalf("count links: %v", err)
	}
	if linkN != 1 {
		t.Fatalf("upcoming_invoice_links = %d, want 1 (orphan preserved)", linkN)
	}
}

func TestDeleteInvoice_CascadesUpcoming(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	br := NewBusinessRepo(db)
	ir := NewInvoiceRepo(db)
	rr := NewRecurringRepo(db)
	ur := NewUpcomingRepo(db)
	lr := NewLinksRepo(db)
	ulr := NewUpcomingLinksRepo(db)

	if err := br.Insert(ctx, model.Business{ID: "B1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	if err := ur.Insert(ctx, model.UpcomingInvoice{ID: "U", BusinessID: "B1", Amount: 1, Currency: "EUR", DueAt: 0, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("up: %v", err)
	}
	if err := rr.Insert(ctx, model.RecurringInvoice{ID: "R", BusinessID: "B1", Amount: 1, Currency: "EUR", Frequency: model.FrequencyMonthly, StartAt: 1, Active: true, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("rec: %v", err)
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "I_U", BusinessID: "B1", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("inv I_U: %v", err)
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "I_R", BusinessID: "B1", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 2, UpdatedAt: 2}); err != nil {
		t.Fatalf("inv I_R: %v", err)
	}
	if err := ulr.Insert(ctx, "I_U", "U"); err != nil {
		t.Fatalf("up link: %v", err)
	}
	if err := lr.Insert(ctx, "I_R", "R", 0); err != nil {
		t.Fatalf("rec link: %v", err)
	}

	upcomingIDs, err := ulr.UpcomingIDsByInvoice(ctx, "I_U")
	if err != nil {
		t.Fatalf("UpcomingIDsByInvoice: %v", err)
	}
	if len(upcomingIDs) != 1 {
		t.Fatalf("upcomingIDs len = %d, want 1", len(upcomingIDs))
	}
	for _, upID := range upcomingIDs {
		if err := ur.Delete(ctx, upID); err != nil {
			t.Fatalf("upcoming delete %s: %v", upID, err)
		}
	}
	if err := ir.Delete(ctx, "I_U"); err != nil {
		t.Fatalf("delete invoice I_U: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoices WHERE id = ?`, "I_U").Scan(&n); err != nil {
		t.Fatalf("count I_U: %v", err)
	}
	if n != 0 {
		t.Fatalf("invoices I_U = %d, want 0", n)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM upcoming_invoices WHERE id = ?`, "U").Scan(&n); err != nil {
		t.Fatalf("count U: %v", err)
	}
	if n != 0 {
		t.Fatalf("upcoming_invoices U = %d, want 0", n)
	}
	if got := countRows(t, ctx, db, "upcoming_invoice_links"); got != 0 {
		t.Fatalf("upcoming_invoice_links = %d, want 0", got)
	}

	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoices WHERE id = ?`, "I_R").Scan(&n); err != nil {
		t.Fatalf("count I_R: %v", err)
	}
	if n != 1 {
		t.Fatalf("invoices I_R = %d, want 1", n)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM recurring_invoices WHERE id = ?`, "R").Scan(&n); err != nil {
		t.Fatalf("count R: %v", err)
	}
	if n != 1 {
		t.Fatalf("recurring_invoices R = %d, want 1", n)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM recurring_invoice_links WHERE recurring_id = ?`, "R").Scan(&n); err != nil {
		t.Fatalf("count rec links: %v", err)
	}
	if n != 1 {
		t.Fatalf("recurring_invoice_links for R = %d, want 1", n)
	}
}

func TestCascadeDelete(t *testing.T) {
	ctx := context.Background()
	br, ir, rr, ur := openAllRepos(t)

	if err := br.Insert(ctx, model.Business{ID: "biz", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "i", BusinessID: "biz", Amount: 1, Currency: "EUR", IssuedAt: 1, DueAt: 2, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("inv: %v", err)
	}
	if err := rr.Insert(ctx, model.RecurringInvoice{ID: "r", BusinessID: "biz", Amount: 1, Currency: "EUR", Frequency: model.FrequencyMonthly, StartAt: 1, Active: true, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("rec: %v", err)
	}
	if err := ur.Insert(ctx, model.UpcomingInvoice{ID: "u", BusinessID: "biz", Amount: 1, Currency: "EUR", DueAt: 2, CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("up: %v", err)
	}

	if err := br.Delete(ctx, "biz"); err != nil {
		t.Fatalf("delete biz: %v", err)
	}

	if inv, _ := ir.List(ctx, 0, 0, nil, nil, nil, false, 0, 0, "issuedAt", "desc"); len(inv) != 0 {
		t.Fatalf("invoices after cascade: %d, want 0", len(inv))
	}
	if rec, _ := rr.List(ctx); len(rec) != 0 {
		t.Fatalf("recurring after cascade: %d, want 0", len(rec))
	}
	if up, _ := ur.List(ctx); len(up) != 0 {
		t.Fatalf("upcoming after cascade: %d, want 0", len(up))
	}
}

func TestInvoiceList_BusinessFilter(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	for _, b := range []model.Business{
		{ID: "biz1", Name: "B1", CreatedAt: 1, UpdatedAt: 1},
		{ID: "biz2", Name: "B2", CreatedAt: 2, UpdatedAt: 2},
	} {
		if err := br.Insert(ctx, b); err != nil {
			t.Fatalf("biz insert %s: %v", b.ID, err)
		}
	}
	seed := []model.Invoice{
		{ID: "i1", BusinessID: "biz1", Amount: 10, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i2", BusinessID: "biz1", Amount: 20, Currency: "EUR", IssuedAt: 200, DueAt: 250, CreatedAt: 200, UpdatedAt: 200},
		{ID: "i3", BusinessID: "biz2", Amount: 30, Currency: "EUR", IssuedAt: 300, DueAt: 350, CreatedAt: 300, UpdatedAt: 300},
		{ID: "i4", BusinessID: "biz2", Amount: 40, Currency: "EUR", IssuedAt: 400, DueAt: 450, CreatedAt: 400, UpdatedAt: 400},
	}
	for _, inv := range seed {
		if err := ir.Insert(ctx, inv); err != nil {
			t.Fatalf("inv insert %s: %v", inv.ID, err)
		}
	}

	cases := []struct {
		name        string
		businessIDs []string
		wantIDs     []string
	}{
		{"nil", nil, []string{"i4", "i3", "i2", "i1"}},
		{"empty slice", []string{}, []string{"i4", "i3", "i2", "i1"}},
		{"biz1 only", []string{"biz1"}, []string{"i2", "i1"}},
		{"biz2 only", []string{"biz2"}, []string{"i4", "i3"}},
		{"both", []string{"biz1", "biz2"}, []string{"i4", "i3", "i2", "i1"}},
		{"nonexistent", []string{"nonexistent"}, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := ir.List(ctx, 0, 0, tc.businessIDs, nil, nil, false, 0, 0, "issuedAt", "desc")
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			got := make([]string, len(list))
			for i, inv := range list {
				got[i] = inv.ID
			}
			if !reflect.DeepEqual(got, tc.wantIDs) {
				t.Fatalf("got %v, want %v", got, tc.wantIDs)
			}
			n, err := ir.Count(ctx, 0, 0, tc.businessIDs, nil, nil, false)
			if err != nil {
				t.Fatalf("count: %v", err)
			}
			if n != len(list) {
				t.Fatalf("count = %d, want %d", n, len(list))
			}
		})
	}
}

func TestInvoiceStats_BusinessFilter(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	for _, b := range []model.Business{
		{ID: "biz1", Name: "B1", CreatedAt: 1, UpdatedAt: 1},
		{ID: "biz2", Name: "B2", CreatedAt: 2, UpdatedAt: 2},
	} {
		if err := br.Insert(ctx, b); err != nil {
			t.Fatalf("biz insert %s: %v", b.ID, err)
		}
	}
	seed := []model.Invoice{
		{ID: "i1", BusinessID: "biz1", Amount: 100, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i2", BusinessID: "biz1", Amount: 200, Currency: "EUR", IssuedAt: 200, DueAt: 250, CreatedAt: 200, UpdatedAt: 200},
		{ID: "i3", BusinessID: "biz2", Amount: 50, Currency: "EUR", IssuedAt: 300, DueAt: 350, CreatedAt: 300, UpdatedAt: 300},
		{ID: "i4", BusinessID: "biz2", Amount: 25, Currency: "EUR", IssuedAt: 400, DueAt: 450, CreatedAt: 400, UpdatedAt: 400},
	}
	for _, inv := range seed {
		if err := ir.Insert(ctx, inv); err != nil {
			t.Fatalf("inv insert %s: %v", inv.ID, err)
		}
	}

	t.Run("nil", func(t *testing.T) {
		out, err := ir.Stats(ctx, 0, 0, nil, nil, nil)
		if err != nil {
			t.Fatalf("stats: %v", err)
		}
		if len(out) != 1 {
			t.Fatalf("len = %d, want 1", len(out))
		}
		s := out[0]
		if s.Total != 375 || s.Count != 4 {
			t.Fatalf("total/count = %v/%d, want 375/4", s.Total, s.Count)
		}
		if s.BusinessCount != 2 {
			t.Fatalf("BusinessCount = %d, want 2", s.BusinessCount)
		}
	})

	t.Run("biz1 only", func(t *testing.T) {
		out, err := ir.Stats(ctx, 0, 0, []string{"biz1"}, nil, nil)
		if err != nil {
			t.Fatalf("stats: %v", err)
		}
		if len(out) != 1 {
			t.Fatalf("len = %d, want 1", len(out))
		}
		s := out[0]
		if s.Total != 300 || s.Count != 2 {
			t.Fatalf("total/count = %v/%d, want 300/2", s.Total, s.Count)
		}
		if s.BusinessCount != 1 {
			t.Fatalf("BusinessCount = %d, want 1", s.BusinessCount)
		}
		if len(s.TopBusinesses) != 1 {
			t.Fatalf("TopBusinesses len = %d, want 1", len(s.TopBusinesses))
		}
		if s.TopBusinesses[0].BusinessID != "biz1" || s.TopBusinesses[0].Amount != 300 {
			t.Fatalf("TopBusinesses[0] = %q/%v, want biz1/300", s.TopBusinesses[0].BusinessID, s.TopBusinesses[0].Amount)
		}
	})

	t.Run("nonexistent", func(t *testing.T) {
		out, err := ir.Stats(ctx, 0, 0, []string{"nonexistent"}, nil, nil)
		if err != nil {
			t.Fatalf("stats: %v", err)
		}
		if out == nil {
			t.Fatal("stats nonexistent returned nil, want empty non-nil slice")
		}
		if len(out) != 0 {
			t.Fatalf("len = %d, want 0", len(out))
		}
	})
}

func seedUsersAndInvoices(t *testing.T, ctx context.Context, br *BusinessRepo, ir *InvoiceRepo, ur *UserRepo, ulr *UserLinksRepo) {
	t.Helper()
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}
	for _, u := range []model.User{
		{ID: "u1", Name: "Alice", CreatedAt: 10, UpdatedAt: 10},
		{ID: "u2", Name: "Bob", CreatedAt: 20, UpdatedAt: 20},
	} {
		if err := ur.Insert(ctx, u); err != nil {
			t.Fatalf("user insert %s: %v", u.ID, err)
		}
	}
	invs := []model.Invoice{
		{ID: "i1", BusinessID: "biz1", Amount: 100, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i2", BusinessID: "biz1", Amount: 200, Currency: "EUR", IssuedAt: 200, DueAt: 250, CreatedAt: 200, UpdatedAt: 200},
		{ID: "i3", BusinessID: "biz1", Amount: 300, Currency: "EUR", IssuedAt: 300, DueAt: 350, CreatedAt: 300, UpdatedAt: 300},
	}
	for _, inv := range invs {
		if err := ir.Insert(ctx, inv); err != nil {
			t.Fatalf("inv insert %s: %v", inv.ID, err)
		}
	}
	if err := ulr.ReplaceForInvoice(ctx, "i1", []string{"u1"}); err != nil {
		t.Fatalf("replace i1: %v", err)
	}
	if err := ulr.ReplaceForInvoice(ctx, "i2", []string{"u1", "u2"}); err != nil {
		t.Fatalf("replace i2: %v", err)
	}
}

func TestInvoiceList_UserFilter(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	br := NewBusinessRepo(db)
	ir := NewInvoiceRepo(db)
	ur := NewUserRepo(db)
	ulr := NewUserLinksRepo(db)
	seedUsersAndInvoices(t, ctx, br, ir, ur, ulr)

	cases := []struct {
		name    string
		userIDs []string
		wantIDs []string
	}{
		{"nil", nil, []string{"i1", "i2", "i3"}},
		{"empty slice", []string{}, []string{"i1", "i2", "i3"}},
		{"u1 only", []string{"u1"}, []string{"i1", "i2"}},
		{"u2 only", []string{"u2"}, []string{"i2"}},
		{"u1 and u2 distinct", []string{"u1", "u2"}, []string{"i1", "i2"}},
		{"nonexistent", []string{"nonexistent"}, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := ir.List(ctx, 0, 0, nil, tc.userIDs, nil, false, 0, 0, "issuedAt", "asc")
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			got := make([]string, len(list))
			for i, inv := range list {
				got[i] = inv.ID
			}
			sort.Strings(got)
			want := append([]string{}, tc.wantIDs...)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
			if len(list) != len(tc.wantIDs) {
				t.Fatalf("len = %d, want %d (must not duplicate)", len(list), len(tc.wantIDs))
			}
			n, err := ir.Count(ctx, 0, 0, nil, tc.userIDs, nil, false)
			if err != nil {
				t.Fatalf("count: %v", err)
			}
			if n != len(list) {
				t.Fatalf("count = %d, want %d", n, len(list))
			}
		})
	}
}

func TestInvoiceStats_UserFilter(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	br := NewBusinessRepo(db)
	ir := NewInvoiceRepo(db)
	ur := NewUserRepo(db)
	ulr := NewUserLinksRepo(db)
	seedUsersAndInvoices(t, ctx, br, ir, ur, ulr)

	out, err := ir.Stats(ctx, 0, 0, nil, []string{"u1", "u2"}, nil)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1", len(out))
	}
	s := out[0]
	if s.Currency != "EUR" {
		t.Fatalf("Currency = %q, want EUR", s.Currency)
	}
	if s.Count != 2 {
		t.Fatalf("Count = %d, want 2 (i2 must not duplicate)", s.Count)
	}
	if s.Total != 300 {
		t.Fatalf("Total = %v, want 300 (i1.Amount + i2.Amount)", s.Total)
	}
}

func TestUserLinks_DistinctUserIDs(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	br := NewBusinessRepo(db)
	ir := NewInvoiceRepo(db)
	ur := NewUserRepo(db)
	ulr := NewUserLinksRepo(db)

	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	for _, u := range []model.User{
		{ID: "u1", Name: "Alice", CreatedAt: 10, UpdatedAt: 10},
		{ID: "u2", Name: "Bob", CreatedAt: 20, UpdatedAt: 20},
	} {
		if err := ur.Insert(ctx, u); err != nil {
			t.Fatalf("user %s: %v", u.ID, err)
		}
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "i1", BusinessID: "biz1", Amount: 1, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100}); err != nil {
		t.Fatalf("i1: %v", err)
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "i2", BusinessID: "biz1", Amount: 1, Currency: "EUR", IssuedAt: 500, DueAt: 550, CreatedAt: 500, UpdatedAt: 500}); err != nil {
		t.Fatalf("i2: %v", err)
	}
	if err := ulr.ReplaceForInvoice(ctx, "i1", []string{"u1"}); err != nil {
		t.Fatalf("link i1: %v", err)
	}
	if err := ulr.ReplaceForInvoice(ctx, "i2", []string{"u2"}); err != nil {
		t.Fatalf("link i2: %v", err)
	}

	t.Run("unbounded", func(t *testing.T) {
		got, err := ulr.DistinctUserIDs(ctx, 0, 0)
		if err != nil {
			t.Fatalf("distinct: %v", err)
		}
		if !reflect.DeepEqual(got, []string{"u1", "u2"}) {
			t.Fatalf("got %v, want [u1 u2]", got)
		}
	})

	t.Run("range covers only i1", func(t *testing.T) {
		got, err := ulr.DistinctUserIDs(ctx, 0, 200)
		if err != nil {
			t.Fatalf("distinct: %v", err)
		}
		if !reflect.DeepEqual(got, []string{"u1"}) {
			t.Fatalf("got %v, want [u1]", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		got, err := ulr.DistinctUserIDs(ctx, 10000, 20000)
		if err != nil {
			t.Fatalf("distinct: %v", err)
		}
		if got == nil {
			t.Fatal("got nil, want non-nil empty slice")
		}
		if len(got) != 0 {
			t.Fatalf("len = %d, want 0", len(got))
		}
	})
}

func seedTagsAndInvoices(t *testing.T, ctx context.Context, br *BusinessRepo, ir *InvoiceRepo, tr *TagRepo, itr *InvoiceTagsRepo) {
	t.Helper()
	if err := br.Insert(ctx, model.Business{ID: "b1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}
	for _, tg := range []model.Tag{
		{ID: "t1", Name: "alpha", CreatedAt: 10, UpdatedAt: 10},
		{ID: "t2", Name: "beta", CreatedAt: 20, UpdatedAt: 20},
	} {
		if err := tr.Insert(ctx, tg); err != nil {
			t.Fatalf("tag insert %s: %v", tg.ID, err)
		}
	}
	invs := []model.Invoice{
		{ID: "i1", BusinessID: "b1", Amount: 100, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i2", BusinessID: "b1", Amount: 200, Currency: "EUR", IssuedAt: 200, DueAt: 250, CreatedAt: 200, UpdatedAt: 200},
		{ID: "i3", BusinessID: "b1", Amount: 50, Currency: "EUR", IssuedAt: 300, DueAt: 350, CreatedAt: 300, UpdatedAt: 300},
	}
	for _, inv := range invs {
		if err := ir.Insert(ctx, inv); err != nil {
			t.Fatalf("inv insert %s: %v", inv.ID, err)
		}
	}
	if err := itr.ReplaceForInvoice(ctx, "i1", []string{"t1"}); err != nil {
		t.Fatalf("link i1: %v", err)
	}
	if err := itr.ReplaceForInvoice(ctx, "i2", []string{"t1", "t2"}); err != nil {
		t.Fatalf("link i2: %v", err)
	}
}

func TestInvoiceList_TagFilter(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	br := NewBusinessRepo(db)
	ir := NewInvoiceRepo(db)
	tr := NewTagRepo(db)
	itr := NewInvoiceTagsRepo(db)
	seedTagsAndInvoices(t, ctx, br, ir, tr, itr)

	cases := []struct {
		name    string
		tagIDs  []string
		wantIDs []string
	}{
		{"nil", nil, []string{"i1", "i2", "i3"}},
		{"empty slice", []string{}, []string{"i1", "i2", "i3"}},
		{"t1 only", []string{"t1"}, []string{"i1", "i2"}},
		{"t2 only", []string{"t2"}, []string{"i2"}},
		{"t1 and t2 distinct", []string{"t1", "t2"}, []string{"i1", "i2"}},
		{"nonexistent", []string{"nonexistent"}, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := ir.List(ctx, 0, 0, nil, nil, tc.tagIDs, false, 0, 0, "issuedAt", "asc")
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			got := make([]string, len(list))
			for i, inv := range list {
				got[i] = inv.ID
			}
			sort.Strings(got)
			want := append([]string{}, tc.wantIDs...)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
			if len(list) != len(tc.wantIDs) {
				t.Fatalf("len = %d, want %d (must not duplicate)", len(list), len(tc.wantIDs))
			}
			n, err := ir.Count(ctx, 0, 0, nil, nil, tc.tagIDs, false)
			if err != nil {
				t.Fatalf("count: %v", err)
			}
			if n != len(list) {
				t.Fatalf("count = %d, want %d", n, len(list))
			}
		})
	}
}

func TestInvoiceStats_TagFilter(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	br := NewBusinessRepo(db)
	ir := NewInvoiceRepo(db)
	tr := NewTagRepo(db)
	itr := NewInvoiceTagsRepo(db)
	seedTagsAndInvoices(t, ctx, br, ir, tr, itr)

	t.Run("t1 and t2 distinct", func(t *testing.T) {
		out, err := ir.Stats(ctx, 0, 0, nil, nil, []string{"t1", "t2"})
		if err != nil {
			t.Fatalf("stats: %v", err)
		}
		if len(out) != 1 {
			t.Fatalf("len = %d, want 1", len(out))
		}
		s := out[0]
		if s.Count != 2 {
			t.Fatalf("Count = %d, want 2 (i2 must not duplicate)", s.Count)
		}
		if s.Total != 300 {
			t.Fatalf("Total = %v, want 300", s.Total)
		}
	})

	t.Run("t1 only", func(t *testing.T) {
		out, err := ir.Stats(ctx, 0, 0, nil, nil, []string{"t1"})
		if err != nil {
			t.Fatalf("stats: %v", err)
		}
		if len(out) != 1 {
			t.Fatalf("len = %d, want 1", len(out))
		}
		s := out[0]
		if s.Count != 2 {
			t.Fatalf("Count = %d, want 2", s.Count)
		}
		if s.Total != 300 {
			t.Fatalf("Total = %v, want 300", s.Total)
		}
	})

	t.Run("nonexistent", func(t *testing.T) {
		out, err := ir.Stats(ctx, 0, 0, nil, nil, []string{"nonexistent"})
		if err != nil {
			t.Fatalf("stats: %v", err)
		}
		if out == nil {
			t.Fatal("got nil, want non-nil empty slice")
		}
		if len(out) != 0 {
			t.Fatalf("len = %d, want 0", len(out))
		}
	})
}

func TestInvoiceTags_DistinctTagIDs(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	br := NewBusinessRepo(db)
	ir := NewInvoiceRepo(db)
	tr := NewTagRepo(db)
	itr := NewInvoiceTagsRepo(db)

	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz: %v", err)
	}
	for _, tg := range []model.Tag{
		{ID: "t1", Name: "alpha", CreatedAt: 10, UpdatedAt: 10},
		{ID: "t2", Name: "beta", CreatedAt: 20, UpdatedAt: 20},
	} {
		if err := tr.Insert(ctx, tg); err != nil {
			t.Fatalf("tag %s: %v", tg.ID, err)
		}
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "i1", BusinessID: "biz1", Amount: 1, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100}); err != nil {
		t.Fatalf("i1: %v", err)
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "i2", BusinessID: "biz1", Amount: 1, Currency: "EUR", IssuedAt: 500, DueAt: 550, CreatedAt: 500, UpdatedAt: 500}); err != nil {
		t.Fatalf("i2: %v", err)
	}
	if err := itr.ReplaceForInvoice(ctx, "i1", []string{"t1"}); err != nil {
		t.Fatalf("link i1: %v", err)
	}
	if err := itr.ReplaceForInvoice(ctx, "i2", []string{"t2"}); err != nil {
		t.Fatalf("link i2: %v", err)
	}

	t.Run("unbounded", func(t *testing.T) {
		got, err := itr.DistinctTagIDs(ctx, 0, 0)
		if err != nil {
			t.Fatalf("distinct: %v", err)
		}
		if !reflect.DeepEqual(got, []string{"t1", "t2"}) {
			t.Fatalf("got %v, want [t1 t2]", got)
		}
	})

	t.Run("range covers only i1", func(t *testing.T) {
		got, err := itr.DistinctTagIDs(ctx, 0, 200)
		if err != nil {
			t.Fatalf("distinct: %v", err)
		}
		if !reflect.DeepEqual(got, []string{"t1"}) {
			t.Fatalf("got %v, want [t1]", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		got, err := itr.DistinctTagIDs(ctx, 10000, 20000)
		if err != nil {
			t.Fatalf("distinct: %v", err)
		}
		if got == nil {
			t.Fatal("got nil, want non-nil empty slice")
		}
		if len(got) != 0 {
			t.Fatalf("len = %d, want 0", len(got))
		}
	})
}

func TestTagMapping_BusinessAndInvoiceAreIndependent(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	br := NewBusinessRepo(db)
	ir := NewInvoiceRepo(db)
	tr := NewTagRepo(db)
	btr := NewBusinessTagsRepo(db)
	itr := NewInvoiceTagsRepo(db)

	if err := br.Insert(ctx, model.Business{ID: "b1", Name: "B1", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("b1: %v", err)
	}
	for _, tg := range []model.Tag{
		{ID: "t1", Name: "t1n", CreatedAt: 10, UpdatedAt: 10},
		{ID: "t2", Name: "t2n", CreatedAt: 20, UpdatedAt: 20},
		{ID: "t3", Name: "t3n", CreatedAt: 30, UpdatedAt: 30},
		{ID: "t4", Name: "t4n", CreatedAt: 40, UpdatedAt: 40},
	} {
		if err := tr.Insert(ctx, tg); err != nil {
			t.Fatalf("tag %s: %v", tg.ID, err)
		}
	}
	if err := btr.ReplaceForBusiness(ctx, "b1", []string{"t1"}); err != nil {
		t.Fatalf("biz tag: %v", err)
	}
	if err := ir.Insert(ctx, model.Invoice{ID: "i1", BusinessID: "b1", Amount: 1, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100}); err != nil {
		t.Fatalf("i1: %v", err)
	}
	if err := itr.ReplaceForInvoice(ctx, "i1", []string{"t1", "t2"}); err != nil {
		t.Fatalf("inv tags: %v", err)
	}

	if err := btr.ReplaceForBusiness(ctx, "b1", []string{"t3"}); err != nil {
		t.Fatalf("biz tag rewrite: %v", err)
	}
	bizTags, err := btr.TagIDsByBusiness(ctx, "b1")
	if err != nil {
		t.Fatalf("biz tags: %v", err)
	}
	if !reflect.DeepEqual(bizTags, []string{"t3"}) {
		t.Fatalf("biz tags after rewrite = %v, want [t3]", bizTags)
	}
	invTags, err := itr.TagIDsByInvoice(ctx, "i1")
	if err != nil {
		t.Fatalf("inv tags: %v", err)
	}
	if !reflect.DeepEqual(invTags, []string{"t1", "t2"}) {
		t.Fatalf("inv tags after biz rewrite = %v, want [t1 t2]", invTags)
	}

	if err := br.Insert(ctx, model.Business{ID: "b2", Name: "B2", CreatedAt: 2, UpdatedAt: 2}); err != nil {
		t.Fatalf("b2: %v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE invoices SET business_id = ? WHERE id = ?`, "b2", "i1"); err != nil {
		t.Fatalf("move invoice: %v", err)
	}
	invTags, err = itr.TagIDsByInvoice(ctx, "i1")
	if err != nil {
		t.Fatalf("inv tags: %v", err)
	}
	if !reflect.DeepEqual(invTags, []string{"t1", "t2"}) {
		t.Fatalf("inv tags after biz move = %v, want [t1 t2]", invTags)
	}

	if err := tr.Delete(ctx, "t1"); err != nil {
		t.Fatalf("delete t1: %v", err)
	}
	bizTags, err = btr.TagIDsByBusiness(ctx, "b1")
	if err != nil {
		t.Fatalf("biz tags: %v", err)
	}
	if !reflect.DeepEqual(bizTags, []string{"t3"}) {
		t.Fatalf("biz tags after t1 delete = %v, want [t3]", bizTags)
	}
	invTags, err = itr.TagIDsByInvoice(ctx, "i1")
	if err != nil {
		t.Fatalf("inv tags: %v", err)
	}
	if !reflect.DeepEqual(invTags, []string{"t2"}) {
		t.Fatalf("inv tags after t1 delete = %v, want [t2]", invTags)
	}
}

func TestInvoiceInsert_AllowsEmptyBusinessID(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	if err := InitSchema(db); err != nil {
		t.Fatalf("init: %v", err)
	}
	ir := NewInvoiceRepo(db)

	inv := model.Invoice{
		ID: "i1", BusinessID: "", Amount: 99.5, Currency: "EUR",
		Description: "no biz", IssuedAt: 10, DueAt: 20,
		CreatedAt: 100, UpdatedAt: 100,
	}
	if err := ir.Insert(ctx, inv); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, ok, err := ir.Get(ctx, "i1")
	if err != nil || !ok {
		t.Fatalf("get: ok=%v err=%v", ok, err)
	}
	if got.BusinessID != "" {
		t.Fatalf("BusinessID = %q, want empty", got.BusinessID)
	}

	var raw sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT business_id FROM invoices WHERE id = ?`, "i1").Scan(&raw); err != nil {
		t.Fatalf("raw select: %v", err)
	}
	if raw.Valid {
		t.Fatalf("business_id = %q, want NULL", raw.String)
	}
}

func TestInvoiceList_IncludesNoBusinessInvoices(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "b1", Name: "B1", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}
	seed := []model.Invoice{
		{ID: "i1", BusinessID: "b1", Amount: 10, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i2", BusinessID: "", Amount: 20, Currency: "EUR", IssuedAt: 200, DueAt: 250, CreatedAt: 200, UpdatedAt: 200},
		{ID: "i3", BusinessID: "b1", Amount: 30, Currency: "EUR", IssuedAt: 300, DueAt: 350, CreatedAt: 300, UpdatedAt: 300},
	}
	for _, inv := range seed {
		if err := ir.Insert(ctx, inv); err != nil {
			t.Fatalf("insert %s: %v", inv.ID, err)
		}
	}

	list, err := ir.List(ctx, 0, 0, nil, nil, nil, false, 0, 0, "issuedAt", "desc")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("len = %d, want 3", len(list))
	}
	byID := map[string]model.Invoice{}
	for _, inv := range list {
		byID[inv.ID] = inv
	}
	if byID["i2"].BusinessID != "" {
		t.Fatalf("i2 BusinessID = %q, want empty", byID["i2"].BusinessID)
	}

	ids, err := ir.DistinctBusinessIDs(ctx, 0, 0)
	if err != nil {
		t.Fatalf("distinct: %v", err)
	}
	if !reflect.DeepEqual(ids, []string{"b1"}) {
		t.Fatalf("distinct = %v, want [b1]", ids)
	}
}

func TestInvoiceList_BusinessFilterExcludesNoBusinessRows(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "b1", Name: "B1", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}
	seed := []model.Invoice{
		{ID: "i1", BusinessID: "b1", Amount: 10, Currency: "EUR", IssuedAt: 100, DueAt: 150, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i2", BusinessID: "", Amount: 20, Currency: "EUR", IssuedAt: 200, DueAt: 250, CreatedAt: 200, UpdatedAt: 200},
		{ID: "i3", BusinessID: "b1", Amount: 30, Currency: "EUR", IssuedAt: 300, DueAt: 350, CreatedAt: 300, UpdatedAt: 300},
	}
	for _, inv := range seed {
		if err := ir.Insert(ctx, inv); err != nil {
			t.Fatalf("insert %s: %v", inv.ID, err)
		}
	}

	list, err := ir.List(ctx, 0, 0, []string{"b1"}, nil, nil, false, 0, 0, "issuedAt", "desc")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	got := make([]string, len(list))
	for i, inv := range list {
		got[i] = inv.ID
	}
	want := []string{"i3", "i1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestMigrateNullableBusinessID_Idempotent(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("pragma: %v", err)
	}

	oldSchema := []string{
		`CREATE TABLE businesses (
			id          TEXT    PRIMARY KEY,
			name        TEXT    NOT NULL,
			tax_id      TEXT    NOT NULL DEFAULT '',
			email       TEXT    NOT NULL DEFAULT '',
			address     TEXT    NOT NULL DEFAULT '',
			notes       TEXT    NOT NULL DEFAULT '',
			logo_type   TEXT    NOT NULL DEFAULT '',
			created_at  INTEGER NOT NULL,
			updated_at  INTEGER NOT NULL
		)`,
		`CREATE TABLE invoices (
			id           TEXT    PRIMARY KEY,
			business_id  TEXT    NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
			amount       REAL    NOT NULL,
			currency     TEXT    NOT NULL,
			description  TEXT    NOT NULL DEFAULT '',
			issued_at    INTEGER NOT NULL,
			due_at       INTEGER NOT NULL,
			paid         INTEGER NOT NULL DEFAULT 0,
			paid_at      INTEGER NOT NULL DEFAULT 0,
			created_at   INTEGER NOT NULL,
			updated_at   INTEGER NOT NULL
		)`,
		`CREATE TABLE recurring_invoices (
			id                  TEXT    PRIMARY KEY,
			business_id         TEXT    NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
			amount              REAL    NOT NULL,
			currency            TEXT    NOT NULL,
			description         TEXT    NOT NULL DEFAULT '',
			frequency           TEXT    NOT NULL,
			start_at            INTEGER NOT NULL,
			end_at              INTEGER NOT NULL DEFAULT 0,
			active              INTEGER NOT NULL DEFAULT 1,
			issue_day_of_week   INTEGER NOT NULL DEFAULT 0,
			issue_day_of_month  INTEGER NOT NULL DEFAULT 0,
			issue_month_of_year INTEGER NOT NULL DEFAULT 0,
			created_at          INTEGER NOT NULL,
			updated_at          INTEGER NOT NULL
		)`,
		`CREATE TABLE upcoming_invoices (
			id           TEXT    PRIMARY KEY,
			business_id  TEXT    NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
			amount       REAL    NOT NULL,
			currency     TEXT    NOT NULL,
			description  TEXT    NOT NULL DEFAULT '',
			due_at       INTEGER NOT NULL,
			created_at   INTEGER NOT NULL,
			updated_at   INTEGER NOT NULL
		)`,
	}
	for _, s := range oldSchema {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO businesses (id, name, created_at, updated_at) VALUES ('b1', 'B1', 1, 1)`); err != nil {
		t.Fatalf("biz insert: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO invoices (id, business_id, amount, currency, issued_at, due_at, created_at, updated_at) VALUES ('i1', 'b1', 42.5, 'EUR', 10, 20, 100, 100)`); err != nil {
		t.Fatalf("inv insert: %v", err)
	}

	for _, table := range []string{"invoices", "recurring_invoices", "upcoming_invoices"} {
		var nn int
		if err := db.QueryRow(`SELECT "notnull" FROM pragma_table_info(?) WHERE name = 'business_id'`, table).Scan(&nn); err != nil {
			t.Fatalf("pragma %s: %v", table, err)
		}
		if nn != 1 {
			t.Fatalf("%s notnull pre-migrate = %d, want 1", table, nn)
		}
	}

	if err := migrateNullableBusinessID(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	for _, table := range []string{"invoices", "recurring_invoices", "upcoming_invoices"} {
		var nn int
		if err := db.QueryRow(`SELECT "notnull" FROM pragma_table_info(?) WHERE name = 'business_id'`, table).Scan(&nn); err != nil {
			t.Fatalf("pragma %s post: %v", table, err)
		}
		if nn != 0 {
			t.Fatalf("%s notnull post-migrate = %d, want 0", table, nn)
		}
	}

	var (
		gotID, gotBiz, gotCur string
		gotAmount             float64
	)
	if err := db.QueryRow(`SELECT id, business_id, amount, currency FROM invoices WHERE id = 'i1'`).Scan(&gotID, &gotBiz, &gotAmount, &gotCur); err != nil {
		t.Fatalf("post select: %v", err)
	}
	if gotID != "i1" || gotBiz != "b1" || gotAmount != 42.5 || gotCur != "EUR" {
		t.Fatalf("row preserved badly: id=%q biz=%q amount=%v cur=%q", gotID, gotBiz, gotAmount, gotCur)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO invoices (id, business_id, amount, currency, issued_at, due_at, created_at, updated_at) VALUES ('i2', NULL, 5, 'EUR', 11, 21, 101, 101)`); err != nil {
		t.Fatalf("null insert: %v", err)
	}

	var countBefore int
	if err := db.QueryRow(`SELECT COUNT(*) FROM invoices`).Scan(&countBefore); err != nil {
		t.Fatalf("count: %v", err)
	}
	if err := migrateNullableBusinessID(db); err != nil {
		t.Fatalf("migrate idempotent: %v", err)
	}
	var countAfter int
	if err := db.QueryRow(`SELECT COUNT(*) FROM invoices`).Scan(&countAfter); err != nil {
		t.Fatalf("count after: %v", err)
	}
	if countBefore != countAfter {
		t.Fatalf("row count changed: before=%d after=%d", countBefore, countAfter)
	}
}

func TestInvoiceList_UnpaidOnly(t *testing.T) {
	ctx := context.Background()
	br, ir, _, _ := openAllRepos(t)
	if err := br.Insert(ctx, model.Business{ID: "biz1", Name: "B", CreatedAt: 1, UpdatedAt: 1}); err != nil {
		t.Fatalf("biz insert: %v", err)
	}
	seed := []model.Invoice{
		{ID: "i1", BusinessID: "biz1", Amount: 100, Currency: "EUR", IssuedAt: 100, DueAt: 150, Paid: true, PaidAt: 120, CreatedAt: 100, UpdatedAt: 100},
		{ID: "i2", BusinessID: "biz1", Amount: 200, Currency: "EUR", IssuedAt: 200, DueAt: 250, Paid: false, CreatedAt: 200, UpdatedAt: 200},
	}
	for _, inv := range seed {
		if err := ir.Insert(ctx, inv); err != nil {
			t.Fatalf("insert %s: %v", inv.ID, err)
		}
	}

	list, err := ir.List(ctx, 0, 0, nil, nil, nil, true, 0, 0, "issuedAt", "desc")
	if err != nil {
		t.Fatalf("list unpaidOnly: %v", err)
	}
	if len(list) != 1 || list[0].ID != "i2" {
		t.Fatalf("unpaidOnly list = %+v, want only i2", list)
	}
	n, err := ir.Count(ctx, 0, 0, nil, nil, nil, true)
	if err != nil {
		t.Fatalf("count unpaidOnly: %v", err)
	}
	if n != 1 {
		t.Fatalf("count unpaidOnly = %d, want 1", n)
	}

	list, err = ir.List(ctx, 0, 0, nil, nil, nil, false, 0, 0, "issuedAt", "desc")
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list all len = %d, want 2", len(list))
	}
}
