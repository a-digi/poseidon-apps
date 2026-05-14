package model

type DownloadItem struct {
	ID           string `json:"id"`
	Platform     string `json:"platform"`
	ExternalId   string `json:"platformItemId"`
	TargetFolder string `json:"targetFolder"`
	Filename     string `json:"filename"`
}
