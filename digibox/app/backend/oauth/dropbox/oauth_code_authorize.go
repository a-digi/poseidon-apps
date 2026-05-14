package dropbox

import (
	"log"

	"digibox-plugin/dropbox"
	"digibox-plugin/oauth"
	"digibox-plugin/oauth/repository"
	"digibox-plugin/response"
)

func OauthCodeAuthorize(
	oauthRequestRepository *repository.OauthRequestRepository,
	oauthClientRepository *repository.OauthClientsRepository,
	oauthTokenRepository oauth.TokenRepository,
	profileRepo *repository.OauthProfileRepository,
	state string,
	code string,
) string {

	if state == "" || code == "" {
		return response.ErrorResponse("State und Code sind erforderlich")
	}

	req, err := oauthRequestRepository.FindByState(state)

	if err != nil {
		return response.ErrorResponse("Kein passender OauthRequest gefunden")
	}

	client, err := oauthClientRepository.FindByClientId(req.ClientId)
	if err != nil {
		return response.ErrorResponse("Fehler beim Suchen des Clients: " + err.Error())
	}

	tokenID, err := oauth.AuthorizeWithCode(*client, code, req.ID)
	if err != nil {
		return response.ErrorResponse("Fehler beim Token-Austausch: " + err.Error())
	}

	if client.Provider == "dropbox" {
		tokenRepo, ok := oauthTokenRepository.(*repository.OauthTokenRepository)
		if !ok {
			log.Printf("[OauthCodeAuthorize] TokenRepository ist nicht vom Typ *OauthTokenRepository, Dropbox-Profil kann nicht abgerufen werden.")
		} else if profileRepo == nil {
			log.Printf("[OauthCodeAuthorize] profileRepo ist nil, Dropbox-Profil kann nicht abgerufen werden.")
		} else {
			if _, err := dropbox.GetDropboxProfileByTokenID(tokenID, tokenRepo, profileRepo, nil); err != nil {
				log.Printf("[OauthCodeAuthorize] Fehler beim Abrufen des Dropbox-Profils: %v", err)
			}
		}
	}

	return response.SuccessResponse(tokenID)
}
