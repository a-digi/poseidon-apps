package model

type Invoice struct {
	ID          string  `json:"id"`
	BusinessID  string  `json:"businessId"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description,omitempty"`
	IssuedAt    int64   `json:"issuedAt"`
	DueAt       int64   `json:"dueAt"`
	Paid        bool    `json:"paid"`
	PaidAt      int64   `json:"paidAt,omitempty"`
	CreatedAt   int64   `json:"createdAt"`
	UpdatedAt   int64   `json:"updatedAt"`

	RecurringIDs []string `json:"recurringIds,omitempty"`
	UpcomingIDs  []string `json:"upcomingIds,omitempty"`
	UserIDs      []string `json:"userIds,omitempty"`
	TagIDs       []string `json:"tagIds,omitempty"`
}
