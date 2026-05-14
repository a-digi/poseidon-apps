package dropbox

import (
	"encoding/json"
	"errors"
	"time"

	"digibox-plugin/dropbox/api"
	"digibox-plugin/oauth/config"
	"digibox-plugin/oauth/model"
	repo "digibox-plugin/oauth/repository"
)

func ListDropboxFiles(token *model.OauthToken, tokenRepo *repo.OauthTokenRepository, clientSecret string, path string) (DropResult, error) {
	if token == nil {
		return DropResult{}, errors.New("Kein Token übergeben")
	}
	requestBody, _ := json.Marshal(map[string]interface{}{
		"path":      path,
		"recursive": false,
	})

	if token.Expiry > 0 && time.Now().Unix() > token.Expiry-60 {
		_, err := api.ApiPostRequestWithBody(config.DropboxListFolderURL, token, tokenRepo, clientSecret, requestBody)
		if err != nil {
			return DropResult{}, err
		}
	}

	resp, err := api.ApiPostRequestWithBody(config.DropboxListFolderURL, token, tokenRepo, clientSecret, requestBody)
	if err != nil {
		return DropResult{}, err
	}

	if resp.Status != api.StatusSuccess {
		return DropResult{}, errors.New(resp.ErrorMsg)
	}

	var apiResp struct {
		Entries []map[string]interface{} `json:"entries"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return DropResult{}, err
	}

	entries := make([]DropboxEntry, 0, len(apiResp.Entries))
	for _, e := range apiResp.Entries {
		entry := DropboxEntry{
			Tag:  e[".tag"].(string),
			Name: e["name"].(string),
			Id:   e["id"].(string),
			Path: e["path_display"].(string),
		}
		entries = append(entries, entry)
	}

	return DropResult{
		Entries:  entries,
		Provider: "dropbox",
	}, nil
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
