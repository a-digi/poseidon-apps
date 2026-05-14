package oauth

import (
	"encoding/json"

	"digibox-plugin/oauth/model"
	"digibox-plugin/oauth/repository"
	"digibox-plugin/response"
)

func ListOAuthClients(repo *repository.OauthClientsRepository) string {
	clients, err := repo.FindAll()
	if err != nil {
		return response.ErrorResponse(err.Error())
	}

	return response.SuccessResponse(clients)
}

func AddOAuthClient(oauthClientRepository *repository.OauthClientsRepository, clientJson string) string {
	var client model.OAuthClient
	if err := json.Unmarshal([]byte(clientJson), &client); err != nil {
		return response.ErrorResponse(err.Error())
	}

	if client.Name == "" || client.ClientId == "" || client.Secret == "" {
		return response.ErrorResponse("name, clientId und secret sind erforderlich")
	}

	_, err := oauthClientRepository.Insert(client)

	if err != nil {
		return response.ErrorResponse(err.Error())
	}

	return response.SuccessResponse("Client hinzugefügt")
}

func DeleteOAuthClient(oauthClientRepository *repository.OauthClientsRepository, id string) string {
	err := oauthClientRepository.Delete(id)
	if err != nil {
		return response.ErrorResponse(err.Error())
	}

	return response.SuccessResponse("Client gelöscht")
}

func GetActiveProviders(oauthClientRepository *repository.OauthClientsRepository) string {
	clients, err := oauthClientRepository.FindAll()

	if err != nil {
		return response.ErrorResponse(err.Error())
	}

	return response.SuccessResponse(clients)
}
