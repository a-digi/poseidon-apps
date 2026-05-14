package oauth

import (
	"errors"
	"fmt"

	"digibox-plugin/oauth/model"
)

type GoogleDriveProviderHandler struct {
	BaseProviderHandler
}

func (h *GoogleDriveProviderHandler) CreateAuthorizationLink(client model.OAuthClient, state string) (string, error) {
	if client.Provider == string(ProviderGoogleDrive) {
		if client.ID == "" {
			return "", errors.New("Client ID fehlt für Google Drive")
		}
		return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&response_type=code&scope=https://www.googleapis.com/auth/drive&state=%s", client.ID, state), nil
	}

	if h.next != nil {
		return h.next.CreateAuthorizationLink(client, state)
	}

	return "", errors.New("Unbekannter Provider")
}

func (h *GoogleDriveProviderHandler) AuthorizeWithCode(client model.OAuthClient, code string, requestId string) (string, error) {
	if client.Provider == string(ProviderGoogleDrive) {
		if client.ID == "" || client.Secret == "" {
			return "", errors.New("Client ID oder Secret fehlt für Google Drive")
		}
		return "(Simulierter Google Drive Access Token)", nil
	}

	if h.next != nil {
		return h.next.AuthorizeWithCode(client, code, requestId)
	}

	return "", errors.New("Unbekannter Provider")
}
