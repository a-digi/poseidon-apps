package dropbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"digibox-plugin/oauth/config"
)

func DownloadDropboxFile(accessToken string, pathOrId string) (io.ReadCloser, error) {
	url := config.DropboxDownloadFileURL
	client := &http.Client{}

	apiArg := map[string]string{"path": pathOrId}
	apiArgJson, _ := json.Marshal(apiArg)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Dropbox-API-Arg", string(apiArgJson))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Dropbox-Download-Fehler: %s (%s)", resp.Status, string(body))
	}
	return resp.Body, nil
}

func DownloadFromBody(body io.Reader, destPath string) error {
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, body)
	if err != nil {
		return err
	}
	return nil
}
