package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"digibox-plugin/oauth/model"
	repo "digibox-plugin/oauth/repository"
)

func NewDropboxAPIClient() *http.Client {
	return &http.Client{}
}

func ApiPostRequest(method, url string, token *model.OauthToken, tokenRepo *repo.OauthTokenRepository, clientSecret string) (*DropboxAPIResult, error) {
	if token == nil {
		return nil, errors.New("Kein Token übergeben")
	}
	if token.Expiry > 0 && time.Now().Unix() > token.Expiry-60 {
		refreshed, err := refreshDropboxToken(tokenRepo, token.ID, token.ClientId, clientSecret)
		if err != nil {
			return nil, err
		}
		token = refreshed
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := NewDropboxAPIClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return buildDropboxAPIResult(resp, body), nil
}

func ApiPostEmptyBodyRequest(url, accessToken string) (*DropboxAPIResult, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := NewDropboxAPIClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return buildDropboxAPIResult(resp, body), nil
}

func ApiPostRequestWithBody(url string, token *model.OauthToken, tokenRepo *repo.OauthTokenRepository, clientSecret string, body []byte) (*DropboxAPIResult, error) {
	if token == nil {
		return nil, errors.New("Kein Token übergeben")
	}
	if token.Expiry > 0 && time.Now().Unix() > token.Expiry-60 {
		refreshed, err := refreshDropboxToken(tokenRepo, token.ID, token.ClientId, clientSecret)
		if err != nil {
			return nil, err
		}
		token = refreshed
	}

	client := NewDropboxAPIClient()
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return buildDropboxAPIResult(resp, respBody), nil
}

func ApiContentUpload(url string, token *model.OauthToken, tokenRepo *repo.OauthTokenRepository, clientSecret string, apiArg string, body io.Reader) (*DropboxAPIResult, error) {
	if token == nil {
		return nil, errors.New("Kein Token übergeben")
	}
	if token.Expiry > 0 && time.Now().Unix() > token.Expiry-60 {
		refreshed, err := refreshDropboxToken(tokenRepo, token.ID, token.ClientId, clientSecret)
		if err != nil {
			return nil, err
		}
		token = refreshed
	}

	client := NewDropboxAPIClient()
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", apiArg)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return buildDropboxAPIResult(resp, respBody), nil
}

func refreshDropboxToken(tokenRepo *repo.OauthTokenRepository, tokenID string, clientId string, clientSecret string) (*model.OauthToken, error) {
	token, err := tokenRepo.FindByID(tokenID)
	if err != nil {
		return nil, err
	}
	if token.RefreshToken == "" {
		return nil, errors.New("Kein Refresh-Token vorhanden")
	}

	form := url.Values{}
	form.Set("refresh_token", token.RefreshToken)
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", clientId)
	form.Set("client_secret", clientSecret)

	resp, err := http.PostForm("https://api.dropbox.com/oauth2/token", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("Dropbox Refresh-Token-Request fehlgeschlagen: " + string(body))
	}

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	newAccessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		return nil, errors.New("Kein access_token in Dropbox-Antwort: " + string(body))
	}
	expiry := int64(0)
	if exp, ok := tokenResp["expires_in"]; ok {
		switch v := exp.(type) {
		case float64:
			expiry = time.Now().Unix() + int64(v)
		case int64:
			expiry = time.Now().Unix() + v
		}
	}

	token.AccessToken = newAccessToken
	token.Expiry = expiry
	updatedToken, err := tokenRepo.Update(*token)
	if err != nil {
		return nil, err
	}

	return &updatedToken, nil
}
