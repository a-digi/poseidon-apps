package playerstate

import (
	"database/sql"
	"fmt"

	"music-plugin/response"
)

type PlayerStateService struct {
	repository *PlayerStateRepository
}

func NewPlayerStateService(db *sql.DB) (*PlayerStateService, error) {
	repo, err := NewPlayerStateRepository(db)
	if err != nil {
		return nil, fmt.Errorf("NewPlayerStateRepository: %w", err)
	}
	return &PlayerStateService{repository: repo}, nil
}

func (s *PlayerStateService) Get() string {
	state, err := s.repository.Get()
	if err != nil {
		return response.ErrorResponse(fmt.Sprintf("Fehler beim Laden des Player-States: %v", err))
	}
	return response.SuccessResponse(state)
}

func (s *PlayerStateService) Save(state PlayerState) string {
	saved, err := s.repository.Save(state)
	if err != nil {
		return response.ErrorResponse(fmt.Sprintf("Fehler beim Speichern des Player-States: %v", err))
	}
	return response.SuccessResponse(saved)
}
