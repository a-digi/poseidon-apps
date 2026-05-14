package recurring

import (
	"testing"
	"time"

	"mbud-plugin/model"
)

func mustUTC(y int, mo time.Month, d int) int64 {
	return time.Date(y, mo, d, 0, 0, 0, 0, time.UTC).Unix()
}

func TestNextDueAt_AnchorMonthly(t *testing.T) {
	rule := model.RecurringInvoice{
		StartAt:         mustUTC(2026, 1, 1),
		Frequency:       model.FrequencyMonthly,
		IssueDayOfMonth: 15,
	}
	cases := []struct {
		idx  int
		want int64
	}{
		{0, mustUTC(2026, 1, 15)},
		{1, mustUTC(2026, 2, 15)},
		{12, mustUTC(2027, 1, 15)},
	}
	for _, c := range cases {
		if got := NextDueAt(rule, c.idx); got != c.want {
			t.Fatalf("period %d = %d, want %d", c.idx, got, c.want)
		}
	}
}

func TestNextDueAt_MonthlyClamping(t *testing.T) {
	rule := model.RecurringInvoice{
		StartAt:         mustUTC(2026, 1, 1),
		Frequency:       model.FrequencyMonthly,
		IssueDayOfMonth: 31,
	}
	cases := []struct {
		idx  int
		want int64
	}{
		{0, mustUTC(2026, 1, 31)},
		{1, mustUTC(2026, 2, 28)},
		{2, mustUTC(2026, 3, 31)},
		{3, mustUTC(2026, 4, 30)},
	}
	for _, c := range cases {
		if got := NextDueAt(rule, c.idx); got != c.want {
			t.Fatalf("period %d = %d, want %d", c.idx, got, c.want)
		}
	}
}

func TestNextDueAt_MonthlySkipFirst(t *testing.T) {
	rule := model.RecurringInvoice{
		StartAt:         mustUTC(2026, 1, 20),
		Frequency:       model.FrequencyMonthly,
		IssueDayOfMonth: 15,
	}
	if got := NextDueAt(rule, 0); got != mustUTC(2026, 2, 15) {
		t.Fatalf("period 0 = %d, want %d", got, mustUTC(2026, 2, 15))
	}
}

func TestNextDueAt_AnchorWeekly(t *testing.T) {
	startAt := mustUTC(2026, 1, 5)
	if isoWeekday(startAt) != 1 {
		t.Fatalf("2026-01-05 ISO weekday = %d, want 1 (Monday)", isoWeekday(startAt))
	}
	rule := model.RecurringInvoice{
		StartAt:        startAt,
		Frequency:      model.FrequencyWeekly,
		IssueDayOfWeek: 4,
	}
	if got := NextDueAt(rule, 0); got != mustUTC(2026, 1, 8) {
		t.Fatalf("period 0 = %d, want %d", got, mustUTC(2026, 1, 8))
	}
	if got := NextDueAt(rule, 1); got != mustUTC(2026, 1, 15) {
		t.Fatalf("period 1 = %d, want %d", got, mustUTC(2026, 1, 15))
	}
}

func TestNextDueAt_WeeklySameDay(t *testing.T) {
	rule := model.RecurringInvoice{
		StartAt:        mustUTC(2026, 1, 5),
		Frequency:      model.FrequencyWeekly,
		IssueDayOfWeek: 1,
	}
	if got := NextDueAt(rule, 0); got != mustUTC(2026, 1, 5) {
		t.Fatalf("period 0 = %d, want %d", got, mustUTC(2026, 1, 5))
	}
}

func TestNextDueAt_AnchorYearly(t *testing.T) {
	rule := model.RecurringInvoice{
		StartAt:          mustUTC(2026, 1, 1),
		Frequency:        model.FrequencyYearly,
		IssueMonthOfYear: 4,
		IssueDayOfMonth:  15,
	}
	if got := NextDueAt(rule, 0); got != mustUTC(2026, 4, 15) {
		t.Fatalf("period 0 = %d, want %d", got, mustUTC(2026, 4, 15))
	}
	if got := NextDueAt(rule, 1); got != mustUTC(2027, 4, 15) {
		t.Fatalf("period 1 = %d, want %d", got, mustUTC(2027, 4, 15))
	}
}

func TestNextDueAt_YearlyFeb29(t *testing.T) {
	rule := model.RecurringInvoice{
		StartAt:          mustUTC(2024, 1, 1),
		Frequency:        model.FrequencyYearly,
		IssueMonthOfYear: 2,
		IssueDayOfMonth:  29,
	}
	cases := []struct {
		idx  int
		want int64
	}{
		{0, mustUTC(2024, 2, 29)},
		{1, mustUTC(2025, 2, 28)},
		{4, mustUTC(2028, 2, 29)},
	}
	for _, c := range cases {
		if got := NextDueAt(rule, c.idx); got != c.want {
			t.Fatalf("period %d = %d, want %d", c.idx, got, c.want)
		}
	}
}

func TestNextDueAt_LegacyMonthlyFallback(t *testing.T) {
	rule := model.RecurringInvoice{
		StartAt:   mustUTC(2026, 3, 17),
		Frequency: model.FrequencyMonthly,
	}
	if got := NextDueAt(rule, 0); got != mustUTC(2026, 3, 17) {
		t.Fatalf("period 0 = %d, want %d", got, mustUTC(2026, 3, 17))
	}
	if got := NextDueAt(rule, 1); got != mustUTC(2026, 4, 17) {
		t.Fatalf("period 1 = %d, want %d", got, mustUTC(2026, 4, 17))
	}
}

func TestNextDueAt_LegacyWeeklyFallback(t *testing.T) {
	startAt := mustUTC(2026, 1, 8)
	if isoWeekday(startAt) != 4 {
		t.Fatalf("2026-01-08 ISO weekday = %d, want 4 (Thursday)", isoWeekday(startAt))
	}
	rule := model.RecurringInvoice{
		StartAt:   startAt,
		Frequency: model.FrequencyWeekly,
	}
	if got := NextDueAt(rule, 0); got != mustUTC(2026, 1, 8) {
		t.Fatalf("period 0 = %d, want %d", got, mustUTC(2026, 1, 8))
	}
	if got := NextDueAt(rule, 1); got != mustUTC(2026, 1, 15) {
		t.Fatalf("period 1 = %d, want %d", got, mustUTC(2026, 1, 15))
	}
}

func TestNextDueAt_LegacyYearlyFallback(t *testing.T) {
	rule := model.RecurringInvoice{
		StartAt:   mustUTC(2026, 4, 15),
		Frequency: model.FrequencyYearly,
	}
	if got := NextDueAt(rule, 0); got != mustUTC(2026, 4, 15) {
		t.Fatalf("period 0 = %d, want %d", got, mustUTC(2026, 4, 15))
	}
	if got := NextDueAt(rule, 1); got != mustUTC(2027, 4, 15) {
		t.Fatalf("period 1 = %d, want %d", got, mustUTC(2027, 4, 15))
	}
}
