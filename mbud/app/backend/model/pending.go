package model

type PendingInvoice struct {
	Source      string   `json:"source"`
	SourceID    string   `json:"sourceId"`
	BusinessID  string   `json:"businessId"`
	Amount      float64  `json:"amount"`
	Currency    string   `json:"currency"`
	Description string   `json:"description,omitempty"`
	DueAt       int64    `json:"dueAt"`
	IssuedAt    int64    `json:"issuedAt"`
	TagIDs      []string `json:"tagIds,omitempty"`
	UserIDs     []string `json:"userIds,omitempty"`
}
