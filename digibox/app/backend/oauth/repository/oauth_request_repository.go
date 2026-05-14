package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"digibox-plugin/oauth/model"
)

type OauthRequestRepository struct {
	DB *sql.DB
}

func NewOauthRequestRepository(db *sql.DB) (*OauthRequestRepository, error) {
	r := &OauthRequestRepository{DB: db}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("oauth_requests migration: %w", err)
	}
	return r, nil
}

func (r *OauthRequestRepository) migrate() error {
	_, err := r.DB.Exec(`
		CREATE TABLE IF NOT EXISTS oauth_requests (
			id           TEXT PRIMARY KEY,
			client_id    TEXT NOT NULL DEFAULT '',
			state        TEXT NOT NULL DEFAULT '',
			requested_on INTEGER NOT NULL DEFAULT 0,
			status       TEXT NOT NULL DEFAULT ''
		)
	`)
	return err
}

func (r *OauthRequestRepository) Insert(req model.OauthRequest) (model.OauthRequest, error) {
	req.ID = uuid.NewString()
	_, err := r.DB.Exec(
		`INSERT INTO oauth_requests (id, client_id, state, requested_on, status)
		 VALUES (?, ?, ?, ?, ?)`,
		req.ID, req.ClientId, req.State, req.RequestedOn, req.Status,
	)
	if err != nil {
		return model.OauthRequest{}, err
	}
	return req, nil
}

func (r *OauthRequestRepository) FindById(id string) (*model.OauthRequest, error) {
	var req model.OauthRequest
	err := r.DB.QueryRow(
		`SELECT id, client_id, state, requested_on, status FROM oauth_requests WHERE id = ?`, id,
	).Scan(&req.ID, &req.ClientId, &req.State, &req.RequestedOn, &req.Status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kein Request mit ID %s gefunden", id)
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *OauthRequestRepository) FindByState(state string) (*model.OauthRequest, error) {
	var req model.OauthRequest
	err := r.DB.QueryRow(
		`SELECT id, client_id, state, requested_on, status FROM oauth_requests WHERE state = ?`, state,
	).Scan(&req.ID, &req.ClientId, &req.State, &req.RequestedOn, &req.Status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kein Request mit state %s gefunden", state)
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}
