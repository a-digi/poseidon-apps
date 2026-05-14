package builtin

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"digibox-plugin/oauth/model"
)

//go:embed clients.json
var clientsJSON []byte

func Load() ([]model.OAuthClient, error) {
	return load(clientsJSON)
}

func load(data []byte) ([]model.OAuthClient, error) {
	var clients []model.OAuthClient
	if err := json.Unmarshal(data, &clients); err != nil {
		return nil, fmt.Errorf("builtin clients: %w", err)
	}
	for i, c := range clients {
		if !strings.HasPrefix(c.ID, "builtin:") {
			return nil, fmt.Errorf("builtin client id %q must start with \"builtin:\"", c.ID)
		}
		clients[i].Builtin = true
	}
	return clients, nil
}
