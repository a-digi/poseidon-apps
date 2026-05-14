package model

type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
	FrequencyYearly  Frequency = "yearly"
)

type RecurringInvoice struct {
	ID               string    `json:"id"`
	BusinessID       string    `json:"businessId"`
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	Description      string    `json:"description,omitempty"`
	Frequency        Frequency `json:"frequency"`
	StartAt          int64     `json:"startAt"`
	EndAt            int64     `json:"endAt,omitempty"`
	Active           bool      `json:"active"`
	IssueDayOfWeek   int       `json:"issueDayOfWeek,omitempty"`
	IssueDayOfMonth  int       `json:"issueDayOfMonth,omitempty"`
	IssueMonthOfYear int       `json:"issueMonthOfYear,omitempty"`
	CreatedAt        int64     `json:"createdAt"`
	UpdatedAt        int64     `json:"updatedAt"`
	UserIDs          []string  `json:"userIds,omitempty"`
	TagIDs           []string  `json:"tagIds,omitempty"`
}
