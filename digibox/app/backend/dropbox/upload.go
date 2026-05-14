package dropbox

import (
	"encoding/json"
	"errors"
	"os"

	"digibox-plugin/dropbox/api"
	"digibox-plugin/oauth/config"
	"digibox-plugin/oauth/model"
	repo "digibox-plugin/oauth/repository"
)

func UploadDropboxFile(token *model.OauthToken, tokenRepo *repo.OauthTokenRepository, clientSecret string, sourcePath string, dropboxPath string) error {
	if token == nil {
		return errors.New("Kein Token übergeben")
	}
	f, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer f.Close()

	apiArg, _ := json.Marshal(map[string]interface{}{
		"path":       dropboxPath,
		"mode":       "add",
		"autorename": true,
		"mute":       false,
	})

	resp, err := api.ApiContentUpload(config.DropboxUploadFileURL, token, tokenRepo, clientSecret, string(apiArg), f)
	if err != nil {
		return err
	}
	if resp.Status != api.StatusSuccess {
		return errors.New(resp.ErrorMsg)
	}
	return nil
}
