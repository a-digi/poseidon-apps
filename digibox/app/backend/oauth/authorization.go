package oauth

import (
	"digibox-plugin/oauth/model"
)

type Provider string

const (
	ProviderDropbox     Provider = "dropbox"
	ProviderGoogleDrive Provider = "googledrive"
)

type OAuthProviderHandler interface {
	SetNext(next OAuthProviderHandler)
	CreateAuthorizationLink(client model.OAuthClient, state string) (string, error)
	AuthorizeWithCode(client model.OAuthClient, code string, requestId string) (string, error)
}

type BaseProviderHandler struct {
	next OAuthProviderHandler
}

func (b *BaseProviderHandler) SetNext(next OAuthProviderHandler) {
	b.next = next
}

func buildOAuthProviderChain(tokenRepo TokenRepository) OAuthProviderHandler {
	dropbox := NewDropboxProviderHandler(tokenRepo)
	google := &GoogleDriveProviderHandler{}
	dropbox.SetNext(google)
	return dropbox
}

var providerChain OAuthProviderHandler

func InitProviderChain(tokenRepo TokenRepository) {
	providerChain = buildOAuthProviderChain(tokenRepo)
}

func CreateAuthorizationLink(client model.OAuthClient, state string) (string, error) {
	return providerChain.CreateAuthorizationLink(client, state)
}

func AuthorizeWithCode(client model.OAuthClient, code string, requestId string) (string, error) {
	return providerChain.AuthorizeWithCode(client, code, requestId)
}
