package whatsgate

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/infrastructure/gateways/whatsapp/whatsgate/types"

	"github.com/stretchr/testify/require"
)

// stubServer возвращает httptest сервер, который обрабатывает /send и /status
func stubServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(ts.Close)
	return ts
}

func newGateway(baseURL string) *WhatsGateGateway {
	cfg := &types.WhatsGateConfig{
		BaseURL:       baseURL,
		APIKey:        "test-api-key",
		WhatsappID:    "test-wa-id",
		Timeout:       2 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    10 * time.Millisecond,
		MaxFileSize:   types.MaxFileSizeBytes,
	}
	return NewWhatsGateGateway(cfg)
}

func TestSendTextMessage(t *testing.T) {
	testCases := []struct {
		name          string
		phoneNumber   string
		message       string
		mockHandler   func(t *testing.T, w http.ResponseWriter, r *http.Request)
		expectSuccess bool
		expectErrorIn string
	}{
		{
			name:        "success_text_message",
			phoneNumber: "79161234567",
			message:     "hello",
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/send", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				err := json.NewEncoder(w).Encode(map[string]string{"status": "sent", "id": "msg123"})
				require.NoError(t, err)
			},
			expectSuccess: true,
		},
		{
			name:          "invalid_phone_number",
			phoneNumber:   "123",
			message:       "hi",
			mockHandler:   nil, // Валидация происходит до вызова сервера
			expectSuccess: false,
			expectErrorIn: "invalid phone number",
		},
		{
			name:        "unauthorized_error_from_server",
			phoneNumber: "79161234567",
			message:     "msg",
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				// Имитируем более реалистичный ответ API при ошибке
				err := json.NewEncoder(w).Encode(types.SendMessageResponse{
					Status:  "error",
					Message: "Invalid API key",
				})
				require.NoError(t, err)
			},
			expectSuccess: false,
			expectErrorIn: "Invalid API key", // Теперь ожидаем осмысленное сообщение
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var serverURL string
			if tc.mockHandler != nil {
				server := stubServer(t, func(w http.ResponseWriter, r *http.Request) {
					tc.mockHandler(t, w, r)
				})
				serverURL = server.URL
			}

			gw := newGateway(serverURL)
			res, err := gw.SendTextMessage(context.Background(), tc.phoneNumber, tc.message, false)

			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, tc.expectSuccess, res.Success)
			if tc.expectErrorIn != "" {
				require.Contains(t, res.Error, tc.expectErrorIn)
			}
		})
	}
}

func TestSendMediaMessage_TooLarge(t *testing.T) {
	gw := newGateway("http://localhost")
	big := bytes.Repeat([]byte("a"), int(types.MaxFileSizeBytes)+1)
	res, err := gw.SendMediaMessage(context.Background(), "79161234567", campaign.MessageTypeImage, "photo", "big.jpg", bytes.NewReader(big), "image/jpeg", false)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.False(t, res.Success, "expected failure for big file")
	require.Contains(t, res.Error, "file size")
}

func TestTestConnection(t *testing.T) {
	testCases := []struct {
		name          string
		retryAttempts int
		mockHandler   func(t *testing.T, w http.ResponseWriter, r *http.Request)
		expectSuccess bool
		expectErrorIn string
	}{
		{
			name:          "success_on_first_try",
			retryAttempts: 1,
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				err := json.NewEncoder(w).Encode(types.TestConnectionResponse{Result: "success", Data: true})
				require.NoError(t, err)
			},
			expectSuccess: true,
		},
		{
			name:          "success_on_second_try_after_500",
			retryAttempts: 2,
			mockHandler: func() func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				calls := 0
				return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
					calls++
					if calls == 1 {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					err := json.NewEncoder(w).Encode(types.TestConnectionResponse{Result: "success", Data: true})
					require.NoError(t, err)
				}
			}(),
			expectSuccess: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var serverURL string
			if tc.mockHandler != nil {
				server := stubServer(t, func(w http.ResponseWriter, r *http.Request) {
					tc.mockHandler(t, w, r)
				})
				serverURL = server.URL
			}

			gw := newGateway(serverURL)
			gw.config.RetryAttempts = tc.retryAttempts
			gw.config.RetryDelay = 1 * time.Millisecond

			res, err := gw.TestConnection(context.Background())

			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, tc.expectSuccess, res.Success)
			if tc.expectErrorIn != "" {
				require.Contains(t, res.Error, tc.expectErrorIn)
			}
		})
	}
}
