package whatsgate

import (
	"context"
	"io"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/infrastructure/gateways/whatsgate/types"
	"whatsapp-service/internal/usecases/interfaces"
)

// SettingsAwareGateway получает актуальные креды из репозитория в рантайме.
type SettingsAwareGateway struct {
	repo interfaces.WhatsGateSettingsRepository
}

// NewSettingsAwareGateway создаёт ленивый шлюз, «знающий» о репозитории
// настроек.  Каждый вызов сначала забирает запись whatsgate_settings из
// БД (можно кешировать/добавить TTL, если понадобится), затем формирует
// внутренний WhatsGateGateway и делегирует ему отправку.
func NewSettingsAwareGateway(repo interfaces.WhatsGateSettingsRepository) interfaces.MessageGateway {
	return &SettingsAwareGateway{repo: repo}
}

func (d *SettingsAwareGateway) buildGateway(ctx context.Context) (interfaces.MessageGateway, error) {
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
	return NewWhatsGateGateway(cfg), nil
}

// SendTextMessage реализует interfaces.MessageGateway.
// Перед реальной отправкой формируется внутренний клиент с текущими
// параметрами What​sGate.  Если настройка ещё не сохранена — возвращаем
// ошибку в поле Error и Success=false.
func (d *SettingsAwareGateway) SendTextMessage(ctx context.Context, phone, message string, async bool) (types.MessageResult, error) {
	gw, err := d.buildGateway(ctx)
	if err != nil {
		return types.MessageResult{PhoneNumber: phone, Success: false, Error: "settings not configured"}, nil
	}
	return gw.SendTextMessage(ctx, phone, message, async)
}

// SendMediaMessage аналогичен SendTextMessage, но отправляет медиа-файл.
// Все ограничения (размер файла, MIME) проверяются внутри базового
// WhatsGateGateway.
func (d *SettingsAwareGateway) SendMediaMessage(ctx context.Context, phone string, mt entities.MessageType, message, filename string, media io.Reader, mime string, async bool) (types.MessageResult, error) {
	gw, err := d.buildGateway(ctx)
	if err != nil {
		return types.MessageResult{PhoneNumber: phone, Success: false, Error: "settings not configured"}, nil
	}
	return gw.SendMediaMessage(ctx, phone, mt, message, filename, media, mime, async)
}

// TestConnection вызывает эндпоинт What​sGate «ping» с текущими
// реквизитами и возвращает результат.
func (d *SettingsAwareGateway) TestConnection(ctx context.Context) (types.TestConnectionResult, error) {
	gw, err := d.buildGateway(ctx)
	if err != nil {
		return types.TestConnectionResult{Success: false, Error: "settings not configured"}, nil
	}
	return gw.TestConnection(ctx)
}
