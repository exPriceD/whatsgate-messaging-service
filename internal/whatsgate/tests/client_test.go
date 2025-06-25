package whatsgate_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"
)

func TestNewClient(t *testing.T) {
	client := whatsgateDomain.NewClient("https://api.example.com", "test-id", "test-key")

	assert.NotNil(t, client)
}

func TestSendTextMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-key", r.Header.Get("X-Api-Key"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		bodyStr := string(body)
		assert.Contains(t, bodyStr, "71234567890")
		assert.Contains(t, bodyStr, "test message")
		assert.Contains(t, bodyStr, "text")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message_id": "123"}`))
	}))
	defer server.Close()

	client := whatsgateDomain.NewClient(server.URL, "test-id", "test-key")

	ctx := context.Background()
	response, err := client.SendTextMessage(ctx, "71234567890", "test message", false)
	require.NoError(t, err)
	assert.Equal(t, "success", response.Status)
}

func TestSendTextMessageAsync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message_id": "123"}`))
	}))
	defer server.Close()

	client := whatsgateDomain.NewClient(server.URL, "test-id", "test-key")

	ctx := context.Background()
	response, err := client.SendTextMessage(ctx, "71234567890", "test message", true)
	require.NoError(t, err)
	assert.Equal(t, "success", response.Status)
}

func TestSendMediaMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-key", r.Header.Get("X-Api-Key"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		bodyStr := string(body)
		assert.Contains(t, bodyStr, "71234567890")
		assert.Contains(t, bodyStr, "image")
		assert.Contains(t, bodyStr, "ZmFrZS1pbWFnZS1kYXRh")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message_id": "456"}`))
	}))
	defer server.Close()

	client := whatsgateDomain.NewClient(server.URL, "test-id", "test-key")

	imageData := []byte("fake-image-data")
	ctx := context.Background()
	response, err := client.SendMediaMessage(ctx, "71234567890", "image", "Test image", "test.jpg", imageData, "image/jpeg", false)
	require.NoError(t, err)
	assert.Equal(t, "success", response.Status)
}

func TestSendMessageWithoutAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "", r.Header.Get("X-Api-Key"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message_id": "789"}`))
	}))
	defer server.Close()

	client := whatsgateDomain.NewClient(server.URL, "test-id", "")

	ctx := context.Background()
	response, err := client.SendTextMessage(ctx, "71234567890", "test message", false)
	require.NoError(t, err)
	assert.Equal(t, "success", response.Status)
}

func TestSendMessageWithAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized", "message": "Invalid API key"}`))
	}))
	defer server.Close()

	client := whatsgateDomain.NewClient(server.URL, "test-id", "invalid-key")

	ctx := context.Background()
	_, err := client.SendTextMessage(ctx, "71234567890", "test message", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNAUTHORIZED")
}

func TestSendMessageWithServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal Server Error"}`))
	}))
	defer server.Close()

	client := whatsgateDomain.NewClient(server.URL, "test-id", "test-key")

	ctx := context.Background()
	_, err := client.SendTextMessage(ctx, "71234567890", "test message", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SERVER_ERROR")
}

func TestSendMessageWithBadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad Request", "message": "Invalid phone number"}`))
	}))
	defer server.Close()

	client := whatsgateDomain.NewClient(server.URL, "test-id", "test-key")

	ctx := context.Background()
	_, err := client.SendTextMessage(ctx, "invalid-phone", "test message", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestSendMessageWithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Задержка больше таймаута
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := whatsgateDomain.NewClient(server.URL, "test-id", "test-key")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := client.SendTextMessage(ctx, "71234567890", "test message", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NETWORK_ERROR")
}

func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		isValid bool
	}{
		{"valid_phone", "71234567890", true},
		{"valid_phone_with_different_digits", "79876543210", true},
		{"empty_phone", "", false},
		{"too_short_phone", "7123456789", false},
		{"too_long_phone", "712345678901", false},
		{"phone_starting_with_8", "81234567890", false},
		{"phone_starting_with_9", "91234567890", false},
		{"phone_with_plus", "+71234567890", false},
		{"phone_with_spaces", "7 123 456 78 90", false},
		{"phone_with_dashes", "7-123-456-78-90", false},
		{"phone_with_letters", "7a234567890", false},
		{"phone_with_special_chars", "7@234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := whatsgateDomain.ValidatePhoneNumber(tt.phone)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateMessageType(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		isValid bool
	}{
		{"valid_text", "text", true},
		{"valid_image", "image", true},
		{"valid_sticker", "sticker", true},
		{"valid_doc", "doc", true},
		{"valid_voice", "voice", true},
		{"invalid_type", "invalid", false},
		{"empty_type", "", false},
		{"invalid_type_uppercase", "TEXT", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := whatsgateDomain.ValidateMessageType(tt.msgType)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
