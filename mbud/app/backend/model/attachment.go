package model

type Attachment struct {
	ID               string `json:"id"`
	InvoiceID        string `json:"invoiceId,omitempty"`
	SessionID        string `json:"sessionId,omitempty"`
	Mime             string `json:"mime"`
	OriginalFilename string `json:"originalFilename"`
	SizeBytes        int64  `json:"sizeBytes"`
	CreatedAt        int64  `json:"createdAt"`
}
