package playerstate

type PlayerState struct {
	SelectedPlaylistID string `json:"selectedPlaylistId"`
	PlayMode           string `json:"playMode"`
	CurrentItemID      string `json:"currentItemId"`
}
