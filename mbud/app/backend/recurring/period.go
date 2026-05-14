package recurring

import (
	"time"

	"mbud-plugin/model"
)

func NextDueAt(rule model.RecurringInvoice, periodIndex int) int64 {
	switch rule.Frequency {
	case model.FrequencyDaily:
		return rule.StartAt + int64(periodIndex)*86400

	case model.FrequencyWeekly:
		dow := rule.IssueDayOfWeek
		if dow == 0 {
			dow = isoWeekday(rule.StartAt)
		}
		first := firstWeekdayOnOrAfter(rule.StartAt, dow)
		return first + int64(periodIndex)*7*86400

	case model.FrequencyMonthly:
		dom := rule.IssueDayOfMonth
		if dom == 0 {
			dom = time.Unix(rule.StartAt, 0).UTC().Day()
		}
		first := firstMonthlyOnOrAfter(rule.StartAt, dom)
		return addMonthsClampDay(first, periodIndex, dom)

	case model.FrequencyYearly:
		moy := rule.IssueMonthOfYear
		dom := rule.IssueDayOfMonth
		if moy == 0 {
			moy = int(time.Unix(rule.StartAt, 0).UTC().Month())
		}
		if dom == 0 {
			dom = time.Unix(rule.StartAt, 0).UTC().Day()
		}
		first := firstYearlyOnOrAfter(rule.StartAt, moy, dom)
		return addYearsClampDay(first, periodIndex, moy, dom)
	}
	return 0
}

func isoWeekday(epoch int64) int {
	wd := int(time.Unix(epoch, 0).UTC().Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func daysInMonth(year int, m time.Month) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func firstWeekdayOnOrAfter(startAt int64, isoDow int) int64 {
	t := time.Unix(startAt, 0).UTC()
	for i := 0; i < 7; i++ {
		if isoWeekday(t.Unix()) == isoDow {
			return t.Unix()
		}
		t = t.Add(24 * time.Hour)
	}
	return t.Unix()
}

func firstMonthlyOnOrAfter(startAt int64, dom int) int64 {
	t := time.Unix(startAt, 0).UTC()
	y, m, d := t.Year(), t.Month(), t.Day()
	if d <= dom {
		actualDom := dom
		if cap := daysInMonth(y, m); actualDom > cap {
			actualDom = cap
		}
		return time.Date(y, m, actualDom, 0, 0, 0, 0, time.UTC).Unix()
	}
	next := time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC)
	actualDom := dom
	if cap := daysInMonth(next.Year(), next.Month()); actualDom > cap {
		actualDom = cap
	}
	return time.Date(next.Year(), next.Month(), actualDom, 0, 0, 0, 0, time.UTC).Unix()
}

func addMonthsClampDay(firstEmission int64, addMonths int, dom int) int64 {
	first := time.Unix(firstEmission, 0).UTC()
	target := time.Date(first.Year(), first.Month()+time.Month(addMonths), 1, 0, 0, 0, 0, time.UTC)
	actualDom := dom
	if cap := daysInMonth(target.Year(), target.Month()); actualDom > cap {
		actualDom = cap
	}
	return time.Date(target.Year(), target.Month(), actualDom, 0, 0, 0, 0, time.UTC).Unix()
}

func firstYearlyOnOrAfter(startAt int64, moy, dom int) int64 {
	t := time.Unix(startAt, 0).UTC()
	candYear := t.Year()
	sMonth, sDay := int(t.Month()), t.Day()
	if sMonth > moy || (sMonth == moy && sDay > dom) {
		candYear++
	}
	actualDom := dom
	if cap := daysInMonth(candYear, time.Month(moy)); actualDom > cap {
		actualDom = cap
	}
	return time.Date(candYear, time.Month(moy), actualDom, 0, 0, 0, 0, time.UTC).Unix()
}

func addYearsClampDay(first int64, addYears, moy, dom int) int64 {
	firstT := time.Unix(first, 0).UTC()
	targetYear := firstT.Year() + addYears
	actualDom := dom
	if cap := daysInMonth(targetYear, time.Month(moy)); actualDom > cap {
		actualDom = cap
	}
	return time.Date(targetYear, time.Month(moy), actualDom, 0, 0, 0, 0, time.UTC).Unix()
}
