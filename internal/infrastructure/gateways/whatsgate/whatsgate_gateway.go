package whatsgate

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"whatsapp-service/internal/infrastructure/gateways/whatsgate/types"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/usecases/interfaces"
)

// Регулярка для проверки номера (российский формат 7XXXXXXXXXX)
var phoneRegex = regexp.MustCompile(`^7\d{10}$`)

// WhatsGateGateway — «голый» HTTP-клиент What​sGate.
// Он не умеет сам добывать ключи: конфиг передаётся при создании.
// Рекомендуется оборачивать его в SettingsAwareGateway для поддержки
// «горячих» изменений настроек.
type WhatsGateGateway struct {
	config *types.WhatsGateConfig
	client *http.Client
}

// NewWhatsGateGateway возвращает готовый к работе шлюз WhatsGate.
// Функция автоматически подставляет значения по умолчанию, если они
// не заданы в конфиге (таймауты, ретраи, лимит размера файла).
func NewWhatsGateGateway(config *types.WhatsGateConfig) interfaces.MessageGateway {
	if config.Timeout == 0 {
		config.Timeout = types.DefaultTimeout
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = types.DefaultRetryAttempts
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = types.DefaultRetryDelay
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = types.MaxFileSizeBytes
	}

	return &WhatsGateGateway{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// SendTextMessage отправляет текстовое сообщение
func (g *WhatsGateGateway) SendTextMessage(ctx context.Context, phoneNumber, message string, async bool) (types.MessageResult, error) {
	if err := g.validatePhoneNumber(phoneNumber); err != nil {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("invalid phone number: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	if strings.TrimSpace(message) == "" {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       "message cannot be empty",
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	request := types.SendMessageRequest{
		WhatsappID: g.config.WhatsappID,
		Async:      async,
		Recipient: types.Recipient{
			Number: phoneNumber,
		},
		Message: types.Message{
			Type: types.MessageTypeText,
			Body: message,
		},
	}

	return g.sendMessageWithRetry(ctx, request, phoneNumber)
}

func (g *WhatsGateGateway) TestConnection(ctx context.Context) (types.TestConnectionResult, error) {
	request := types.TestConnectionRequest{
		WhatsappID: g.config.WhatsappID,
		Number:     "79317019910",
	}

	return g.testConnectionWithRetry(ctx, request)
}

// SendMediaMessage отправляет медиа-сообщение
func (g *WhatsGateGateway) SendMediaMessage(ctx context.Context, phoneNumber string, messageType entities.MessageType, message string, filename string, mediaData io.Reader, mimeType string, async bool) (types.MessageResult, error) {
	if err := g.validatePhoneNumber(phoneNumber); err != nil {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("invalid phone number: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	if err := g.validateMessageType(string(messageType)); err != nil {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("invalid message type: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	fileData, err := io.ReadAll(mediaData)
	if err != nil {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("failed to read media data: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	if int64(len(fileData)) > g.config.MaxFileSize {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("file size exceeds limit: %d bytes", g.config.MaxFileSize),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	encodedData := base64.StdEncoding.EncodeToString(fileData)

	request := types.SendMessageRequest{
		WhatsappID: g.config.WhatsappID,
		Async:      async,
		Recipient: types.Recipient{
			Number: phoneNumber,
		},
		Message: types.Message{
			Type: string(messageType),
			Body: message,
			Media: &types.Media{
				MimeType: mimeType,
				Data:     encodedData,
				Filename: filename,
			},
		},
	}

	return g.sendMessageWithRetry(ctx, request, phoneNumber)
}

// sendMessageWithRetry отправляет сообщение с повторными попытками
func (g *WhatsGateGateway) sendMessageWithRetry(ctx context.Context, request types.SendMessageRequest, phoneNumber string) (types.MessageResult, error) {
	var lastResult types.MessageResult

	for attempt := 1; attempt <= g.config.RetryAttempts; attempt++ {
		result, err := g.sendMessage(ctx, request, phoneNumber)
		if err != nil {
			lastResult = types.MessageResult{
				PhoneNumber: phoneNumber,
				Success:     false,
				Error:       fmt.Sprintf("attempt %d failed: %v", attempt, err),
				Timestamp:   time.Now().Format(time.RFC3339),
			}
		} else {
			lastResult = result
		}

		if lastResult.Success || !g.isRetryableError(lastResult.Error) {
			return lastResult, nil
		}

		if attempt < g.config.RetryAttempts {
			select {
			case <-ctx.Done():
				return types.MessageResult{
					PhoneNumber: phoneNumber,
					Success:     false,
					Error:       "context cancelled during retry",
					Timestamp:   time.Now().Format(time.RFC3339),
				}, nil
			case <-time.After(g.config.RetryDelay):
			}
		}
	}

	return lastResult, nil
}

// testConnectionWithRetry проверяет соединение с повторными попытками
func (g *WhatsGateGateway) testConnectionWithRetry(ctx context.Context, request types.TestConnectionRequest) (types.TestConnectionResult, error) {
	var lastResult types.TestConnectionResult

	for attempt := 1; attempt <= g.config.RetryAttempts; attempt++ {
		result, err := g.testConnection(ctx, request)
		if err != nil {
			lastResult = types.TestConnectionResult{
				Success:   false,
				Error:     fmt.Sprintf("attempt %d failed: %v", attempt, err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
		} else {
			lastResult = result
		}

		if lastResult.Success || !g.isRetryableError(lastResult.Error) {
			return lastResult, nil
		}

		if attempt < g.config.RetryAttempts {
			select {
			case <-ctx.Done():
				return types.TestConnectionResult{
					Success:   false,
					Error:     "context cancelled during retry",
					Timestamp: time.Now().Format(time.RFC3339),
				}, nil
			case <-time.After(g.config.RetryDelay):
			}
		}
	}

	return lastResult, nil
}

// isRetryableError определяет, стоит ли повторять запрос при данной ошибке
func (g *WhatsGateGateway) isRetryableError(errorMsg string) bool {
	retryableErrors := []string{
		"network error",
		"timeout",
		"connection refused",
		"server error",
		"temporary failure",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errorMsg), retryable) {
			return true
		}
	}

	return false
}

// validatePhoneNumber валидирует номер телефона
func (g *WhatsGateGateway) validatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}

	if !phoneRegex.MatchString(phoneNumber) {
		return fmt.Errorf("phone number must be in format 7XXXXXXXXXX (11 digits starting with 7)")
	}

	return nil
}

// validateMessageType валидирует тип сообщения
func (g *WhatsGateGateway) validateMessageType(messageType string) error {
	allowedTypes := map[string]struct{}{
		types.MessageTypeText:    {},
		types.MessageTypeImage:   {},
		types.MessageTypeDoc:     {},
		types.MessageTypeVoice:   {},
		types.MessageTypeSticker: {},
	}

	if _, exists := allowedTypes[messageType]; !exists {
		return fmt.Errorf("unsupported message type: %s", messageType)
	}

	return nil
}

// sendMessage общий метод для отправки сообщений
func (g *WhatsGateGateway) sendMessage(ctx context.Context, request types.SendMessageRequest, phoneNumber string) (types.MessageResult, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("failed to marshal request: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.config.BaseURL+"/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("failed to create request: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", g.config.APIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("network error: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("failed to read response: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var response types.SendMessageResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return types.MessageResult{
				PhoneNumber: phoneNumber,
				Success:     true,
				Status:      "sent",
				Timestamp:   time.Now().Format(time.RFC3339),
			}, nil
		}

		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     true,
			Status:      "sent",
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil

	case http.StatusUnauthorized:
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       "unauthorized: invalid credentials",
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil

	case http.StatusInternalServerError:
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("server error: %s", string(body)),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil

	default:
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("API error: HTTP %d - %s", resp.StatusCode, string(body)),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}
}

// TestConnection проверяет соединение с API
func (g *WhatsGateGateway) testConnection(ctx context.Context, request types.TestConnectionRequest) (types.TestConnectionResult, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return types.TestConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to marshal request: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.config.BaseURL+"/check", bytes.NewBuffer(jsonData))
	if err != nil {
		return types.TestConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to create request: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", g.config.APIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return types.TestConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("network error: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.TestConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to read response: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var response types.TestConnectionResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return types.TestConnectionResult{
				Success:   true,
				Timestamp: time.Now().Format(time.RFC3339),
			}, nil
		}

		return types.TestConnectionResult{
			Success:   true,
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil

	case http.StatusUnauthorized:
		return types.TestConnectionResult{
			Success:   false,
			Error:     "unauthorized: invalid credentials",
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil

	case http.StatusInternalServerError:
		return types.TestConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("server error: %s", string(body)),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil

	default:
		return types.TestConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("API error: HTTP %d - %s", resp.StatusCode, string(body)),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}
}
