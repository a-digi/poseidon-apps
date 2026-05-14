package downloads

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	downloadsmodel "digibox-plugin/downloads/model"
)

type DownloadsRepository struct {
	DB *sql.DB
}

func NewDownloadsRepository(db *sql.DB) (*DownloadsRepository, error) {
	r := &DownloadsRepository{DB: db}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("downloads migration: %w", err)
	}
	return r, nil
}

func (r *DownloadsRepository) migrate() error {
	_, err := r.DB.Exec(`
		CREATE TABLE IF NOT EXISTS downloads (
			id            TEXT PRIMARY KEY,
			platform      TEXT NOT NULL DEFAULT '',
			external_id   TEXT NOT NULL DEFAULT '',
			target_folder TEXT NOT NULL DEFAULT ''
		)
	`)
	return err
}

func (r *DownloadsRepository) Insert(platform, externalId, targetFolder string) error {
	_, err := r.DB.Exec(
		`INSERT INTO downloads (id, platform, external_id, target_folder) VALUES (?, ?, ?, ?)`,
		uuid.NewString(), platform, externalId, targetFolder,
	)
	return err
}

func (r *DownloadsRepository) FindByPlatformAndExternalId(platform, externalId string) (*downloadsmodel.DownloadItem, error) {
	var item downloadsmodel.DownloadItem
	err := r.DB.QueryRow(
		`SELECT id, platform, external_id, target_folder FROM downloads
		 WHERE platform = ? AND external_id = ?`, platform, externalId,
	).Scan(&item.ID, &item.Platform, &item.ExternalId, &item.TargetFolder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DownloadsRepository) DeleteByID(id string) error {
	_, err := r.DB.Exec(`DELETE FROM downloads WHERE id = ?`, id)
	return err
}
