package dropbox

type DropboxClient struct {
	accessToken string
}

type DropResult struct {
	Entries  []DropboxEntry `json:"entries"`
	Provider string         `json:"provider"`
}

type DropboxEntry struct {
	Tag          string         `json:"tag"`
	Name         string         `json:"name"`
	Id           string         `json:"id"`
	Path         string         `json:"path"`
	Children     []DropboxEntry `json:"children,omitempty"`
	TargetFolder string         `json:"targetFolder,omitempty"`
	Downloaded   bool           `json:"downloaded"`
}

type listFolderRequest struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

type listFolderResponse struct {
	Entries []DropboxEntry `json:"entries"`
	Cursor  string         `json:"cursor"`
	HasMore bool           `json:"has_more"`
}

type DropboxProfile struct {
	AccountId string `json:"account_id"`
	Name      struct {
		DisplayName string `json:"display_name"`
		Surname     string `json:"surname"`
		GivenName   string `json:"given_name"`
	} `json:"name"`
	Email string `json:"email"`
}

type DownloadProgress struct {
	Percent  int
	CustomId string
}
