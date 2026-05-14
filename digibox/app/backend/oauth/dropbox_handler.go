package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"digibox-plugin/oauth/config"
	"digibox-plugin/oauth/model"
)

type DropboxProviderHandler struct {
	BaseProviderHandler
	TokenRepo TokenRepository
}

type TokenRepository interface {
	Insert(token model.OauthToken) (model.OauthToken, error)
}

func NewDropboxProviderHandler(tokenRepo TokenRepository) *DropboxProviderHandler {
	return &DropboxProviderHandler{TokenRepo: tokenRepo}
}

func (h *DropboxProviderHandler) CreateAuthorizationLink(client model.OAuthClient, state string) (string, error) {

	if client.Provider == string(ProviderDropbox) {
		if client.ClientId == "" {
			return "", errors.New("ClientId fehlt für Dropbox")
		}

		redirectUri := config.DropboxRedirectURI

		return fmt.Sprintf("https://www.dropbox.com/oauth2/authorize?client_id=%s&response_type=code&state=%s&redirect_uri=%s&token_access_type=offline", client.ClientId, state, redirectUri), nil
	}

	if h.next != nil {
		return h.next.CreateAuthorizationLink(client, state)
	}

	return "", errors.New("Unbekannter Provider")
}

func (h *DropboxProviderHandler) AuthorizeWithCode(client model.OAuthClient, code string, requestId string) (string, error) {
	if client.Provider != string(ProviderDropbox) {
		if h.next != nil {
			return h.next.AuthorizeWithCode(client, code, requestId)
		}

		return "", errors.New("Unbekannter Provider")
	}
	if client.ClientId == "" || client.Secret == "" {
		return "", errors.New("ClientId oder Secret fehlt für Dropbox")
	}

	redirectUri := config.DropboxRedirectURI

	params := url.Values{}
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	params.Set("client_id", client.ClientId)
	params.Set("client_secret", client.Secret)
	params.Set("redirect_uri", redirectUri)

	resp, err := http.PostForm("https://api.dropbox.com/oauth2/token", params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Dropbox Token-Request fehlgeschlagen: %s", string(body))
	}

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	token, ok := tokenResp["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("Kein access_token in Dropbox-Antwort: %s", string(body))
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
	refreshToken := ""
	if rt, ok := tokenResp["refresh_token"].(string); ok {
		refreshToken = rt
	}
	oauthToken := model.OauthToken{
		ID:           client.ID,
		ClientId:     client.ClientId,
		Provider:     client.Provider,
		AccessToken:  token,
		RefreshToken: refreshToken,
		Expiry:       expiry,
		RequestID:    requestId,
	}

	createdToken, err := h.TokenRepo.Insert(oauthToken)
	if err != nil {
		return "", fmt.Errorf("Fehler beim Speichern des Tokens: %w", err)
	}

	return createdToken.ID, nil
}
