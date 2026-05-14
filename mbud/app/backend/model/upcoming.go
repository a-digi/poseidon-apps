package model

type UpcomingInvoice struct {
	ID          string   `json:"id"`
	BusinessID  string   `json:"businessId"`
	Amount      float64  `json:"amount"`
	Currency    string   `json:"currency"`
	Description string   `json:"description,omitempty"`
	DueAt       int64    `json:"dueAt"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
	UserIDs     []string `json:"userIds,omitempty"`
	TagIDs      []string `json:"tagIds,omitempty"`
}
