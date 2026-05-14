package model

type OAuthClient struct {
	ID          string `json:"id"`
	ClientId    string `json:"clientId"`
	Secret      string `json:"secret"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Provider    string `json:"provider"`
	Builtin     bool   `json:"builtin"`
}

type OauthToken struct {
	ID           string `json:"id"`
	ClientId     string `json:"clientId"`
	Provider     string `json:"provider"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	Expiry       int64  `json:"expiry,omitempty"`
	RequestID    string `json:"requestId,omitempty"`
}

type OauthRequest struct {
	ID          string `json:"id"`
	ClientId    string `json:"clientId"`
	State       string `json:"state"`
	RequestedOn int64  `json:"requestedOn"`
	Status      string `json:"status"`
}

type OauthProfile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	DisplayName   string `json:"displayName"`
	GivenName     string `json:"givenName"`
	Surname       string `json:"surname"`
	ProfilePhoto  string `json:"profilePhoto"`
	Provider      string `json:"provider"`
	AccountType   string `json:"accountType"`
	Country       string `json:"country"`
	TokenId       string `json:"tokenId"`
}

type OauthCode struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type OauthRequestView struct {
	ID            string        `json:"id"`
	RequestedOn   int64         `json:"requestedOn"`
	OauthClientID string        `json:"oauthClientId,omitempty"` // internal OAuthClient.ID resolved from token
	OauthToken    OauthToken    `json:"oauthToken"`
	OauthProfile  *OauthProfile `json:"oauthProfile,omitempty"`
}
