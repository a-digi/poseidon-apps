package dropbox

import (
	"log"
	"time"

	"digibox-plugin/oauth"
	"digibox-plugin/oauth/model"
	"digibox-plugin/oauth/repository"
	"digibox-plugin/platform"
	"digibox-plugin/response"
)

func CreateAuthorizationLink(
	oauthClientRepository *repository.OauthClientsRepository,
	oauthRequestRepository *repository.OauthRequestRepository,
	clientID string,
	state string,
) string {

	client, err := oauthClientRepository.FindByID(clientID)
	if err != nil {
		return response.ErrorResponse("Fehler beim Suchen des Clients: " + err.Error())
	}
	if client == nil {
		return response.ErrorResponse("Client mit dieser ID nicht gefunden")
	}

	if oauthRequestRepository != nil {
		req := model.OauthRequest{
			ID:          client.ID + "-" + state,
			ClientId:    client.ClientId,
			State:       state,
			RequestedOn: time.Now().Unix(),
			Status:      "pending",
		}
		_, err := oauthRequestRepository.Insert(req)
		if err != nil {
			log.Printf("[OAUTH] Fehler beim Speichern des OauthRequest: %v", err)
		}
	}
	link, err := oauth.CreateAuthorizationLink(*client, state)
	if err != nil {
		return response.ErrorResponse(err.Error())
	}
	platform.OpenBrowser(link)
	return response.SuccessResponse(link)
}
