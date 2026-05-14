package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"digibox-plugin/oauth/model"
)

type OauthClientsRepository struct {
	DB       *sql.DB
	builtins []model.OAuthClient
}

func NewOauthClientsRepository(db *sql.DB, builtins []model.OAuthClient) (*OauthClientsRepository, error) {
	r := &OauthClientsRepository{DB: db, builtins: builtins}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("oauth_clients migration: %w", err)
	}
	return r, nil
}

func (r *OauthClientsRepository) migrate() error {
	_, err := r.DB.Exec(`
		CREATE TABLE IF NOT EXISTS oauth_clients (
			id          TEXT PRIMARY KEY,
			client_id   TEXT NOT NULL,
			name        TEXT NOT NULL DEFAULT '',
			secret      TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			provider    TEXT NOT NULL DEFAULT ''
		)
	`)
	return err
}

func (r *OauthClientsRepository) Insert(client model.OAuthClient) (model.OAuthClient, error) {
	client.ID = uuid.NewString()
	_, err := r.DB.Exec(
		`INSERT INTO oauth_clients (id, client_id, name, secret, description, provider)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		client.ID, client.ClientId, client.Name, client.Secret, client.Description, client.Provider,
	)
	if err != nil {
		return model.OAuthClient{}, err
	}
	return client, nil
}

func (r *OauthClientsRepository) FindAll() ([]model.OAuthClient, error) {
	rows, err := r.DB.Query(
		`SELECT id, client_id, name, secret, description, provider FROM oauth_clients`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []model.OAuthClient{}
	for rows.Next() {
		var c model.OAuthClient
		if err := rows.Scan(&c.ID, &c.ClientId, &c.Name, &c.Secret, &c.Description, &c.Provider); err != nil {
			return nil, err
		}
		if len(c.Secret) > 1 {
			c.Secret = string(c.Secret[0]) + "****" + string(c.Secret[len(c.Secret)-1])
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, b := range r.builtins {
		b.ClientId = ""
		b.Secret = ""
		result = append(result, b)
	}
	return result, nil
}

func (r *OauthClientsRepository) FindByID(id string) (*model.OAuthClient, error) {
	if strings.HasPrefix(id, "builtin:") {
		for _, b := range r.builtins {
			if b.ID == id {
				c := b
				return &c, nil
			}
		}
		return nil, fmt.Errorf("kein OAuthClient mit ID %s gefunden", id)
	}

	var c model.OAuthClient
	err := r.DB.QueryRow(
		`SELECT id, client_id, name, secret, description, provider FROM oauth_clients WHERE id = ?`, id,
	).Scan(&c.ID, &c.ClientId, &c.Name, &c.Secret, &c.Description, &c.Provider)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kein OAuthClient mit ID %s gefunden", id)
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *OauthClientsRepository) FindByClientId(clientId string) (*model.OAuthClient, error) {
	var c model.OAuthClient
	err := r.DB.QueryRow(
		`SELECT id, client_id, name, secret, description, provider FROM oauth_clients WHERE client_id = ?`, clientId,
	).Scan(&c.ID, &c.ClientId, &c.Name, &c.Secret, &c.Description, &c.Provider)
	if err == nil {
		return &c, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	for _, b := range r.builtins {
		if b.ClientId == clientId {
			cp := b
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("kein OAuthClient mit clientId %s gefunden", clientId)
}

func (r *OauthClientsRepository) Delete(id string) error {
	if strings.HasPrefix(id, "builtin:") {
		return errors.New("built-in clients cannot be deleted")
	}
	res, err := r.DB.Exec(`DELETE FROM oauth_clients WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("OAuthClient nicht gefunden")
	}
	return nil
}
