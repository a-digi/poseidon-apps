package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"digibox-plugin/oauth/model"
)

type OauthProfileRepository struct {
	DB *sql.DB
}

func NewOauthProfileRepository(db *sql.DB) (*OauthProfileRepository, error) {
	r := &OauthProfileRepository{DB: db}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("oauth_profiles migration: %w", err)
	}
	return r, nil
}

func (r *OauthProfileRepository) migrate() error {
	_, err := r.DB.Exec(`
		CREATE TABLE IF NOT EXISTS oauth_profiles (
			id            TEXT PRIMARY KEY,
			email         TEXT NOT NULL DEFAULT '',
			email_verified INTEGER NOT NULL DEFAULT 0,
			display_name  TEXT NOT NULL DEFAULT '',
			given_name    TEXT NOT NULL DEFAULT '',
			surname       TEXT NOT NULL DEFAULT '',
			profile_photo TEXT NOT NULL DEFAULT '',
			provider      TEXT NOT NULL DEFAULT '',
			account_type  TEXT NOT NULL DEFAULT '',
			country       TEXT NOT NULL DEFAULT '',
			token_id      TEXT NOT NULL DEFAULT ''
		)
	`)
	return err
}

func (r *OauthProfileRepository) Insert(profile model.OauthProfile) (model.OauthProfile, error) {
	profile.ID = uuid.NewString()
	emailVerified := 0
	if profile.EmailVerified {
		emailVerified = 1
	}
	_, err := r.DB.Exec(
		`INSERT INTO oauth_profiles
			(id, email, email_verified, display_name, given_name, surname, profile_photo, provider, account_type, country, token_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		profile.ID, profile.Email, emailVerified, profile.DisplayName, profile.GivenName,
		profile.Surname, profile.ProfilePhoto, profile.Provider, profile.AccountType,
		profile.Country, profile.TokenId,
	)
	if err != nil {
		return model.OauthProfile{}, err
	}
	return profile, nil
}

func (r *OauthProfileRepository) FindByTokenId(tokenId string) (*model.OauthProfile, error) {
	var p model.OauthProfile
	var emailVerified int
	err := r.DB.QueryRow(
		`SELECT id, email, email_verified, display_name, given_name, surname, profile_photo, provider, account_type, country, token_id
		 FROM oauth_profiles WHERE token_id = ?`, tokenId,
	).Scan(&p.ID, &p.Email, &emailVerified, &p.DisplayName, &p.GivenName,
		&p.Surname, &p.ProfilePhoto, &p.Provider, &p.AccountType, &p.Country, &p.TokenId)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kein Profil mit tokenId %s gefunden", tokenId)
	}
	if err != nil {
		return nil, err
	}
	p.EmailVerified = emailVerified == 1
	return &p, nil
}
