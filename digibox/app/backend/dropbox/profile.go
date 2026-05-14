package dropbox

import (
	"encoding/json"
	"errors"
	"log"

	dropboxapi "digibox-plugin/dropbox/api"
	"digibox-plugin/events"
	"digibox-plugin/oauth/model"
	"digibox-plugin/oauth/repository"
)

func GetDropboxProfileByTokenID(
	tokenID string,
	tokenRepo *repository.OauthTokenRepository,
	profileRepo *repository.OauthProfileRepository,
	bus *events.EventBus,
) (*DropboxProfile, error) {
	token, err := tokenRepo.FindByID(tokenID)
	if err != nil || token == nil {
		return nil, errors.New("Token nicht gefunden")
	}
	result, err := dropboxapi.ApiPostEmptyBodyRequest(string(DropboxGetCurrentAccount), token.AccessToken)
	if err != nil {
		return nil, err
	}

	if result.Status != dropboxapi.StatusSuccess {
		return nil, errors.New("Dropbox API Fehler received: " + result.ErrorMsg)
	}

	var profile DropboxProfile
	if err := json.Unmarshal(result.Body, &profile); err != nil {
		return nil, err
	}

	oauthProfile := model.OauthProfile{
		ID:          profile.AccountId,
		Email:       profile.Email,
		DisplayName: profile.Name.DisplayName,
		GivenName:   profile.Name.GivenName,
		Surname:     profile.Name.Surname,
		Provider:    "dropbox",
		TokenId:     tokenID,
	}
	if profileRepo != nil {
		if _, err := profileRepo.Insert(oauthProfile); err != nil {
			log.Printf("[DropboxProfile] Fehler beim Speichern des OauthProfiles: %v", err)
		} else {
			payload, _ := json.Marshal(map[string]string{
				"tokenId":  oauthProfile.TokenId,
				"email":    profile.Email,
				"clientId": token.ClientId,
			})
			if bus != nil {
				bus.Publish("dropboxProfileUpdated", string(payload))
			}
		}
	}
	return &profile, nil
}
