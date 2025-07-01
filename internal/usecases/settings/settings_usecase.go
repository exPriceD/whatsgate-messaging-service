package settings

import (
	"context"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/usecases/interfaces"
)

// Service provides high-level operations for WhatsGate settings.
type Service struct {
	repo interfaces.WhatsGateSettingsRepository
}

func NewService(repo interfaces.WhatsGateSettingsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context) (*entities.WhatsGateSettings, error) {
	return s.repo.Get(ctx)
}

func (s *Service) Update(ctx context.Context, whatsappID, apiKey, baseURL string) (*entities.WhatsGateSettings, error) {
	st := entities.NewWhatsGateSettings(whatsappID, apiKey, baseURL)
	if err := s.repo.Save(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *Service) Reset(ctx context.Context) error {
	return s.repo.Reset(ctx)
}
