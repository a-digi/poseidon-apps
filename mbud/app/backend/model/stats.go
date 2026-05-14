package model

type TopBusiness struct {
	BusinessID string  `json:"businessId"`
	Amount     float64 `json:"amount"`
}

type CurrencyStats struct {
	Currency      string        `json:"currency"`
	Total         float64       `json:"total"`
	Count         int           `json:"count"`
	PaidAmount    float64       `json:"paidAmount"`
	PaidCount     int           `json:"paidCount"`
	UnpaidAmount  float64       `json:"unpaidAmount"`
	UnpaidCount   int           `json:"unpaidCount"`
	PendingAmount float64       `json:"pendingAmount"`
	PendingCount  int           `json:"pendingCount"`
	Average       float64       `json:"average"`
	MaxAmount     float64       `json:"maxAmount"`
	BusinessCount int           `json:"businessCount"`
	TopBusinesses []TopBusiness `json:"topBusinesses"`
	TopDayEpoch   int64         `json:"topDayEpoch"`
	TopDayAmount  float64       `json:"topDayAmount"`
}
