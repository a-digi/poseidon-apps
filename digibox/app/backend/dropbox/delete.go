package dropbox

import (
	"encoding/json"
	"errors"

	"digibox-plugin/dropbox/api"
	"digibox-plugin/oauth/config"
	"digibox-plugin/oauth/model"
	repo "digibox-plugin/oauth/repository"
)

func DeleteDropboxItem(token *model.OauthToken, tokenRepo *repo.OauthTokenRepository, clientSecret string, path string) error {
	if token == nil {
		return errors.New("Kein Token übergeben")
	}
	body, _ := json.Marshal(map[string]interface{}{
		"path": path,
	})

	resp, err := api.ApiPostRequestWithBody(config.DropboxDeleteItemURL, token, tokenRepo, clientSecret, body)
	if err != nil {
		return err
	}
	if resp.Status != api.StatusSuccess {
		return errors.New(resp.ErrorMsg)
	}
	return nil
}
