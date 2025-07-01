package whatsgate

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"whatsapp-service/internal/infrastructure/gateways/whatsgate/types"

	"whatsapp-service/internal/entities"
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
	return NewWhatsGateGateway(cfg).(*WhatsGateGateway)
}

func TestSendTextMessage_Success(t *testing.T) {
	server := stubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/send" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "id": "123"})
	})

	gw := newGateway(server.URL)

	res, _ := gw.SendTextMessage(context.Background(), "79161234567", "hello", false)

	if !res.Success {
		t.Fatalf("expected success, got %+v", res)
	}
}

func TestSendTextMessage_InvalidPhone(t *testing.T) {
	gw := newGateway("http://localhost")
	res, _ := gw.SendTextMessage(context.Background(), "123", "hi", false)
	if res.Success {
		t.Error("expected failure for invalid phone")
	}
	if !strings.Contains(res.Error, "invalid") {
		t.Errorf("unexpected error: %s", res.Error)
	}
}

func TestSendMediaMessage_TooLarge(t *testing.T) {
	gw := newGateway("http://localhost")
	big := bytes.Repeat([]byte("a"), int(types.MaxFileSizeBytes)+1)
	res, _ := gw.SendMediaMessage(context.Background(), "79161234567", entities.MessageTypeImage, "photo", "big.jpg", bytes.NewReader(big), "image/jpeg", false)
	if res.Success {
		t.Error("expected failure for big file")
	}
	if !strings.Contains(res.Error, "file size") {
		t.Errorf("unexpected error: %s", res.Error)
	}
}

func TestTestConnection_RetrySuccess(t *testing.T) {
	calls := 0
	server := stubServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("oops"))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	gw := newGateway(server.URL)
	gw.config.RetryAttempts = 2
	gw.config.RetryDelay = 1 * time.Millisecond

	res, _ := gw.TestConnection(context.Background())
	if !res.Success {
		t.Errorf("expected success after retry, got %+v", res)
	}
}

func TestSendTextMessage_Unauthorized(t *testing.T) {
	server := stubServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	})

	gw := newGateway(server.URL)

	res, _ := gw.SendTextMessage(context.Background(), "79161234567", "msg", false)
	if res.Success {
		t.Error("should fail with unauthorized")
	}
	if !strings.Contains(res.Error, "unauthorized") {
		t.Errorf("unexpected error: %s", res.Error)
	}
}
