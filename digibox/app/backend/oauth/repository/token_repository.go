package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"digibox-plugin/oauth/model"
)

type OauthTokenRepository struct {
	DB *sql.DB
}

func NewOauthTokenRepository(db *sql.DB) (*OauthTokenRepository, error) {
	r := &OauthTokenRepository{DB: db}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("oauth_tokens migration: %w", err)
	}
	return r, nil
}

func (r *OauthTokenRepository) migrate() error {
	_, err := r.DB.Exec(`
		CREATE TABLE IF NOT EXISTS oauth_tokens (
			id            TEXT PRIMARY KEY,
			client_id     TEXT NOT NULL DEFAULT '',
			provider      TEXT NOT NULL DEFAULT '',
			access_token  TEXT NOT NULL DEFAULT '',
			refresh_token TEXT NOT NULL DEFAULT '',
			expiry        INTEGER NOT NULL DEFAULT 0,
			request_id    TEXT NOT NULL DEFAULT ''
		)
	`)
	return err
}

func scan(row interface {
	Scan(...any) error
}) (model.OauthToken, error) {
	var t model.OauthToken
	err := row.Scan(&t.ID, &t.ClientId, &t.Provider, &t.AccessToken, &t.RefreshToken, &t.Expiry, &t.RequestID)
	return t, err
}

const tokenFields = `id, client_id, provider, access_token, refresh_token, expiry, request_id`

func (r *OauthTokenRepository) Insert(token model.OauthToken) (model.OauthToken, error) {
	token.ID = uuid.NewString()
	_, err := r.DB.Exec(
		`INSERT INTO oauth_tokens (id, client_id, provider, access_token, refresh_token, expiry, request_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		token.ID, token.ClientId, token.Provider, token.AccessToken, token.RefreshToken, token.Expiry, token.RequestID,
	)
	if err != nil {
		return model.OauthToken{}, err
	}
	return token, nil
}

func (r *OauthTokenRepository) FindAll() ([]model.OauthToken, error) {
	rows, err := r.DB.Query(`SELECT ` + tokenFields + ` FROM oauth_tokens`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []model.OauthToken{}
	for rows.Next() {
		t, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (r *OauthTokenRepository) FindByID(id string) (*model.OauthToken, error) {
	t, err := scan(r.DB.QueryRow(`SELECT `+tokenFields+` FROM oauth_tokens WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kein Token mit ID %s gefunden", id)
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *OauthTokenRepository) FindByClientIdAndProvider(clientId, provider string) (*model.OauthToken, error) {
	t, err := scan(r.DB.QueryRow(
		`SELECT `+tokenFields+` FROM oauth_tokens WHERE client_id = ? AND provider = ?`, clientId, provider))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kein Token für clientId %s / provider %s gefunden", clientId, provider)
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *OauthTokenRepository) FindByTokenId(tokenId string) (*model.OauthToken, error) {
	t, err := scan(r.DB.QueryRow(
		`SELECT `+tokenFields+` FROM oauth_tokens WHERE request_id = ?`, tokenId))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kein Token mit requestId %s gefunden", tokenId)
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *OauthTokenRepository) Update(token model.OauthToken) (model.OauthToken, error) {
	res, err := r.DB.Exec(
		`UPDATE oauth_tokens
		 SET client_id=?, provider=?, access_token=?, refresh_token=?, expiry=?, request_id=?
		 WHERE id=?`,
		token.ClientId, token.Provider, token.AccessToken, token.RefreshToken, token.Expiry, token.RequestID, token.ID,
	)
	if err != nil {
		return model.OauthToken{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.OauthToken{}, errors.New("token nicht gefunden")
	}
	return token, nil
}

func (r *OauthTokenRepository) DeleteByID(id string) error {
	token, err := r.FindByID(id)
	if err != nil {
		return fmt.Errorf("token mit ID %s nicht gefunden: %w", id, err)
	}

	if _, err := r.DB.Exec(`DELETE FROM oauth_tokens WHERE id = ?`, id); err != nil {
		return err
	}

	if token.RequestID != "" {
		if _, err := r.DB.Exec(`DELETE FROM oauth_requests WHERE id = ?`, token.RequestID); err != nil {
			return fmt.Errorf("token gelöscht, aber Fehler beim Löschen des Requests: %w", err)
		}
	}
	return nil
}
