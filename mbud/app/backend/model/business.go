package model

type Business struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TaxID     string `json:"taxId,omitempty"`
	Email     string `json:"email,omitempty"`
	Address   string `json:"address,omitempty"`
	Notes     string `json:"notes,omitempty"`
	LogoType  string `json:"logoType,omitempty"`
	Logo      string `json:"logo,omitempty"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`

	TagIDs []string `json:"tagIds,omitempty"`
}
