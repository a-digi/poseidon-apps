package model

type UploadSession struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoiceId,omitempty"`
	Status    string `json:"status"` // "active" | "consumed" | "expired"
	CreatedAt int64  `json:"createdAt"`
	ExpiresAt int64  `json:"expiresAt"`
}
