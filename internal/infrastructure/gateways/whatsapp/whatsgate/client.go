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
	"whatsapp-service/internal/entities/campaign"
	types2 "whatsapp-service/internal/infrastructure/gateways/whatsapp/whatsgate/types"
	"whatsapp-service/internal/usecases/dto"
)

// Регулярка для проверки номера (российский формат 7XXXXXXXXXX)
var phoneRegex = regexp.MustCompile(`^7\d{10}$`)

// WhatsGateGateway — «голый» HTTP-клиент What​sGate.
// Он не умеет сам добывать ключи: конфиг передаётся при создании.
// Рекомендуется оборачивать его в SettingsAwareGateway для поддержки
// «горячих» изменений настроек.
type WhatsGateGateway struct {
	config *types2.WhatsGateConfig
	client *http.Client
}

// NewWhatsGateGateway возвращает готовый к работе шлюз WhatsGate.
// Функция автоматически подставляет значения по умолчанию, если они
// не заданы в конфиге (таймауты, ретраи, лимит размера файла).
func NewWhatsGateGateway(config *types2.WhatsGateConfig) *WhatsGateGateway {
	if config.Timeout == 0 {
		config.Timeout = types2.DefaultTimeout
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = types2.DefaultRetryAttempts
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = types2.DefaultRetryDelay
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = types2.MaxFileSizeBytes
	}

	return &WhatsGateGateway{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// SendTextMessage отправляет текстовое сообщение
func (g *WhatsGateGateway) SendTextMessage(ctx context.Context, phoneNumber, message string, async bool) (*dto.MessageSendResult, error) {
	if err := g.validatePhoneNumber(phoneNumber); err != nil {
		return &dto.MessageSendResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("invalid phone number: %v", err),
			Timestamp:   time.Now(),
		}, nil
	}

	if strings.TrimSpace(message) == "" {
		return &dto.MessageSendResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       "message cannot be empty",
			Timestamp:   time.Now(),
		}, nil
	}

	request := types2.SendMessageRequest{
		WhatsappID: g.config.WhatsappID,
		Async:      async,
		Recipient: types2.Recipient{
			Number: phoneNumber,
		},
		Message: types2.Message{
			Type: types2.MessageTypeText,
			Body: message,
		},
	}

	return g.sendMessageWithRetry(ctx, request, phoneNumber)
}

func (g *WhatsGateGateway) TestConnection(ctx context.Context) (*dto.ConnectionTestResult, error) {
	request := types2.TestConnectionRequest{
		WhatsappID: g.config.WhatsappID,
		Number:     "79317019910",
	}

	return g.testConnectionWithRetry(ctx, request)
}

// SendMediaMessage отправляет медиа-сообщение
func (g *WhatsGateGateway) SendMediaMessage(ctx context.Context, phoneNumber string, messageType campaign.MessageType, message string, filename string, mediaData io.Reader, mimeType string, async bool) (*dto.MessageSendResult, error) {
	if err := g.validatePhoneNumber(phoneNumber); err != nil {
		return &dto.MessageSendResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("invalid phone number: %v", err),
			Timestamp:   time.Now(),
		}, nil
	}

	if err := g.validateMessageType(string(messageType)); err != nil {
		return &dto.MessageSendResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("invalid message type: %v", err),
			Timestamp:   time.Now(),
		}, nil
	}

	fileData, err := io.ReadAll(mediaData)
	if err != nil {
		return &dto.MessageSendResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("failed to read media data: %v", err),
			Timestamp:   time.Now(),
		}, nil
	}

	if int64(len(fileData)) > g.config.MaxFileSize {
		return &dto.MessageSendResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("file size exceeds limit: %d bytes", g.config.MaxFileSize),
			Timestamp:   time.Now(),
		}, nil
	}

	encodedData := base64.StdEncoding.EncodeToString(fileData)

	request := types2.SendMessageRequest{
		WhatsappID: g.config.WhatsappID,
		Async:      async,
		Recipient: types2.Recipient{
			Number: phoneNumber,
		},
		Message: types2.Message{
			Type: string(messageType),
			Body: message,
			Media: &types2.Media{
				MimeType: mimeType,
				Data:     encodedData,
				Filename: filename,
			},
		},
	}

	return g.sendMessageWithRetry(ctx, request, phoneNumber)
}

// sendMessageWithRetry отправляет сообщение с повторными попытками
func (g *WhatsGateGateway) sendMessageWithRetry(ctx context.Context, request types2.SendMessageRequest, phoneNumber string) (*dto.MessageSendResult, error) {
	var lastResult types2.MessageResult

	for attempt := 1; attempt <= g.config.RetryAttempts; attempt++ {
		result, err := g.sendMessage(ctx, request, phoneNumber)
		if err != nil {
			lastResult = types2.MessageResult{
				PhoneNumber: phoneNumber,
				Success:     false,
				Error:       fmt.Sprintf("attempt %d failed: %v", attempt, err),
				Timestamp:   time.Now().Format(time.RFC3339),
			}
		} else {
			lastResult = result
		}

		if lastResult.Success || !g.isRetryableError(lastResult.Error) {
			break
		}

		if attempt < g.config.RetryAttempts {
			select {
			case <-ctx.Done():
				lastResult = types2.MessageResult{
					PhoneNumber: phoneNumber,
					Success:     false,
					Error:       "context cancelled during retry",
					Timestamp:   time.Now().Format(time.RFC3339),
				}
				goto endRetry
			case <-time.After(g.config.RetryDelay):
			}
		}
	}

endRetry:
	ts, _ := time.Parse(time.RFC3339, lastResult.Timestamp)
	return &dto.MessageSendResult{
		PhoneNumber: lastResult.PhoneNumber,
		Success:     lastResult.Success,
		MessageID:   lastResult.Status,
		Error:       lastResult.Error,
		Timestamp:   ts,
	}, nil
}

// testConnectionWithRetry проверяет соединение с повторными попытками
func (g *WhatsGateGateway) testConnectionWithRetry(ctx context.Context, request types2.TestConnectionRequest) (*dto.ConnectionTestResult, error) {
	var lastInfraResult types2.TestConnectionResult
	var lastError error

	for attempt := 1; attempt <= g.config.RetryAttempts; attempt++ {
		result, err := g.testConnection(ctx, request)
		if err != nil {
			lastInfraResult = types2.TestConnectionResult{
				Success:   false,
				Error:     fmt.Sprintf("attempt %d failed: %v", attempt, err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
			lastError = err
		} else {
			lastInfraResult = result
			lastError = nil // Сбрасываем ошибку при успехе
		}

		if lastInfraResult.Success || !g.isRetryableError(lastInfraResult.Error) {
			break
		}

		if attempt < g.config.RetryAttempts {
			select {
			case <-ctx.Done():
				lastInfraResult = types2.TestConnectionResult{
					Success:   false,
					Error:     "context cancelled during retry",
					Timestamp: time.Now().Format(time.RFC3339),
				}
				lastError = ctx.Err()
				goto endRetry
			case <-time.After(g.config.RetryDelay):
			}
		}
	}

endRetry:
	// Если была системная ошибка, а не ошибка API, пробрасываем ее
	if lastError != nil && !lastInfraResult.Success {
		// Но сначала конвертируем то, что есть
		return &dto.ConnectionTestResult{
			Success: false,
			Error:   lastInfraResult.Error,
		}, lastError
	}

	// Конвертируем финальный результат в DTO
	return &dto.ConnectionTestResult{
		Success: lastInfraResult.Success,
		Error:   lastInfraResult.Error,
	}, nil
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
		types2.MessageTypeText:    {},
		types2.MessageTypeImage:   {},
		types2.MessageTypeDoc:     {},
		types2.MessageTypeVoice:   {},
		types2.MessageTypeSticker: {},
	}

	if _, exists := allowedTypes[messageType]; !exists {
		return fmt.Errorf("unsupported message type: %s", messageType)
	}

	return nil
}

// sendMessage общий метод для отправки сообщений
func (g *WhatsGateGateway) sendMessage(ctx context.Context, request types2.SendMessageRequest, phoneNumber string) (types2.MessageResult, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return types2.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("failed to marshal request: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.config.BaseURL+"/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return types2.MessageResult{
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
		return types2.MessageResult{
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
		return types2.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       fmt.Sprintf("failed to read response: %v", err),
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	// Обработка ответа
	if resp.StatusCode != http.StatusOK {
		var errorMsg string
		if resp.StatusCode >= 500 {
			errorMsg = fmt.Sprintf("server error: API returned HTTP %d - %s", resp.StatusCode, string(body))
		} else {
			var errResp types2.SendMessageResponse
			_ = json.Unmarshal(body, &errResp) // Игнорируем ошибку, если тело пустое
			errorMsg = fmt.Sprintf("API client error: HTTP %d. Status: %s. Message: %s", resp.StatusCode, errResp.Status, errResp.Message)
		}

		return types2.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Status:      "failed",
			Error:       errorMsg,
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	var response types2.SendMessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return types2.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     true,
			Status:      "sent",
			Timestamp:   time.Now().Format(time.RFC3339),
		}, nil
	}

	return types2.MessageResult{
		PhoneNumber: phoneNumber,
		Success:     true,
		Status:      "sent",
		Timestamp:   time.Now().Format(time.RFC3339),
	}, nil
}

// TestConnection проверяет соединение с API
func (g *WhatsGateGateway) testConnection(ctx context.Context, request types2.TestConnectionRequest) (types2.TestConnectionResult, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return types2.TestConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to marshal request: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.config.BaseURL+"/check", bytes.NewBuffer(jsonData))
	if err != nil {
		return types2.TestConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to create request: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", g.config.APIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return types2.TestConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("network error: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types2.TestConnectionResult{Success: false, Error: "failed to read response body"}, fmt.Errorf("read response: %w", err)
	}

	// Сначала проверяем статус HTTP
	if resp.StatusCode != http.StatusOK {
		var errorMsg string
		if resp.StatusCode >= 500 {
			errorMsg = fmt.Sprintf("server error: API returned HTTP %d - %s", resp.StatusCode, string(body))
		} else {
			errorMsg = fmt.Sprintf("API client error: HTTP %d - %s", resp.StatusCode, string(body))
		}
		return types2.TestConnectionResult{Success: false, Error: errorMsg, Timestamp: time.Now().Format(time.RFC3339)}, nil
	}

	var response types2.TestConnectionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return types2.TestConnectionResult{Success: false, Error: "failed to decode response", Timestamp: time.Now().Format(time.RFC3339)}, fmt.Errorf("decode response: %w", err)
	}

	// Теперь проверяем содержимое ответа
	isSuccess := response.Result == "success" && response.Data == true
	errorMessage := ""
	if !isSuccess {
		errorMessage = fmt.Sprintf("Connection test failed: %s", response.Result)
	}

	return types2.TestConnectionResult{
		Success:   isSuccess,
		Error:     errorMessage,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}
