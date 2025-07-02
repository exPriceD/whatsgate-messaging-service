package whatsgate

import (
	"context"
	"io"
	"sync"
	"time"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/infrastructure/gateways/whatsgate/types"
	"whatsapp-service/internal/usecases/dto"
	"whatsapp-service/internal/usecases/interfaces"
)

const (
	// defaultCacheTTL — время жизни кэша для настроек шлюза.
	// В течение этого времени шлюз не будет обращаться в БД за настройками.
	defaultCacheTTL = 1 * time.Minute
)

// SettingsAwareGateway получает актуальные креды из репозитория в рантайме
// и кэширует их для повышения производительности.
type SettingsAwareGateway struct {
	repo           interfaces.WhatsGateSettingsRepository
	cachedGateway  interfaces.MessageGateway
	cacheTimestamp time.Time
	cacheTTL       time.Duration
	mu             sync.RWMutex
}

// NewSettingsAwareGateway создаёт ленивый кэширующий шлюз.
func NewSettingsAwareGateway(repo interfaces.WhatsGateSettingsRepository) *SettingsAwareGateway {
	return &SettingsAwareGateway{
		repo:     repo,
		cacheTTL: defaultCacheTTL,
	}
}

// buildOrGetFromCache получает шлюз из кэша или создает новый, если кэш устарел.
func (d *SettingsAwareGateway) buildOrGetFromCache(ctx context.Context) (interfaces.MessageGateway, error) {
	d.mu.RLock()
	if d.cachedGateway != nil && time.Since(d.cacheTimestamp) < d.cacheTTL {
		defer d.mu.RUnlock()
		return d.cachedGateway, nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cachedGateway != nil && time.Since(d.cacheTimestamp) < d.cacheTTL {
		return d.cachedGateway, nil
	}

	settings, err := d.repo.Get(ctx)
	if err != nil || settings == nil {
		return nil, err
	}
	cfg := &types.WhatsGateConfig{
		BaseURL:       settings.BaseURL,
		APIKey:        settings.APIKey,
		WhatsappID:    settings.WhatsappID,
		Timeout:       types.DefaultTimeout,
		RetryAttempts: types.DefaultRetryAttempts,
		RetryDelay:    types.DefaultRetryDelay,
		MaxFileSize:   types.MaxFileSizeBytes,
	}

	newGateway := NewWhatsGateGateway(cfg)

	d.cachedGateway = newGateway
	d.cacheTimestamp = time.Now()

	return newGateway, nil
}

// SendTextMessage реализует interfaces.MessageGateway.
func (d *SettingsAwareGateway) SendTextMessage(ctx context.Context, phone, message string, async bool) (*dto.MessageSendResult, error) {
	gw, err := d.buildOrGetFromCache(ctx)
	if err != nil {
		return &dto.MessageSendResult{PhoneNumber: phone, Success: false, Error: "settings not configured", Timestamp: time.Now()}, nil
	}
	return gw.SendTextMessage(ctx, phone, message, async)
}

// SendMediaMessage аналогичен SendTextMessage.
func (d *SettingsAwareGateway) SendMediaMessage(ctx context.Context, phone string, mt entities.MessageType, message, filename string, media io.Reader, mime string, async bool) (*dto.MessageSendResult, error) {
	gw, err := d.buildOrGetFromCache(ctx)
	if err != nil {
		return &dto.MessageSendResult{PhoneNumber: phone, Success: false, Error: "settings not configured", Timestamp: time.Now()}, nil
	}
	return gw.SendMediaMessage(ctx, phone, mt, message, filename, media, mime, async)
}

// TestConnection вызывает эндпоинт WhatsGate «ping» с текущими реквизитами.
func (d *SettingsAwareGateway) TestConnection(ctx context.Context) (*dto.ConnectionTestResult, error) {
	gw, err := d.buildOrGetFromCache(ctx)
	if err != nil {
		return &dto.ConnectionTestResult{Success: false, Error: "settings not configured"}, nil
	}
	return gw.TestConnection(ctx)
}
