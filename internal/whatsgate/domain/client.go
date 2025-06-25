package domain

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	appErr "whatsapp-service/internal/errors"
)

// Client представляет клиент для работы с WhatGate API
type Client struct {
	baseURL    string
	whatsappID string
	apiKey     string
	httpClient *http.Client
}

// NewClient создает новый клиент WhatGate
func NewClient(baseURL, whatsappID, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		whatsappID: whatsappID,
		apiKey:     apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendMessageRequest представляет запрос на отправку сообщения
type SendMessageRequest struct {
	WhatsappID string    `json:"WhatsappID"`
	Async      bool      `json:"async"`
	Recipient  Recipient `json:"recipient"`
	Message    Message   `json:"message"`
}

// Recipient представляет получателя сообщения
type Recipient struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type,omitempty"`
	Number string `json:"number,omitempty"`
}

// Message представляет сообщение
type Message struct {
	Type  string `json:"type,omitempty"`
	Body  string `json:"body,omitempty"`
	Quote string `json:"quote,omitempty"`
	Media *Media `json:"media,omitempty"`
}

// Media представляет медиа-файл
type Media struct {
	MimeType string `json:"mimetype"`
	Data     string `json:"data"`
	Filename string `json:"filename"`
}

// SendMessageResponse представляет ответ на отправку сообщения
type SendMessageResponse struct {
	ID      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// SendTextMessage отправляет текстовое сообщение
func (c *Client) SendTextMessage(ctx context.Context, phoneNumber, text string, async bool) (*SendMessageResponse, error) {
	request := SendMessageRequest{
		WhatsappID: c.whatsappID,
		Async:      async,
		Recipient: Recipient{
			Number: phoneNumber,
		},
		Message: Message{
			Type: "text",
			Body: text,
		},
	}

	return c.sendMessage(ctx, request)
}

// SendMediaMessage отправляет сообщение с медиа-файлом
func (c *Client) SendMediaMessage(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (*SendMessageResponse, error) {
	encodedData := base64.StdEncoding.EncodeToString(fileData)

	request := SendMessageRequest{
		WhatsappID: c.whatsappID,
		Async:      async,
		Recipient: Recipient{
			Number: phoneNumber,
		},
		Message: Message{
			Type: messageType,
			Body: text,
			Media: &Media{
				MimeType: mimeType,
				Data:     encodedData,
				Filename: filename,
			},
		},
	}

	return c.sendMessage(ctx, request)
}

// sendMessage выполняет HTTP запрос к WhatGate API
func (c *Client) sendMessage(ctx context.Context, request SendMessageRequest) (*SendMessageResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, appErr.New("MARSHAL_ERROR", "failed to marshal request", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, appErr.New("REQUEST_ERROR", "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, appErr.New("NETWORK_ERROR", "failed to send request", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, appErr.New("READ_ERROR", "failed to read response", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return &SendMessageResponse{
			Status:  "success",
			Message: "Message sent successfully",
		}, nil
	case http.StatusUnauthorized:
		return nil, appErr.New("UNAUTHORIZED", "Your request was made with invalid credentials.", nil)
	case http.StatusInternalServerError:
		return nil, appErr.New("SERVER_ERROR", "Server error: "+string(body), nil)
	default:
		return nil, appErr.New("API_ERROR", "API request failed with status "+resp.Status, nil)
	}
}

// ValidatePhoneNumber проверяет корректность номера телефона
func ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return appErr.NewValidationError("phone number is required")
	}

	for _, char := range phone {
		if char < '0' || char > '9' {
			return appErr.NewValidationError("phone number must contain only digits")
		}
	}

	if len(phone) != 11 {
		return appErr.NewValidationError("phone number must be exactly 11 digits")
	}

	if phone[0] != '7' {
		return appErr.NewValidationError("phone number must start with 7")
	}

	return nil
}

// ValidateMessageType проверяет корректность типа сообщения
func ValidateMessageType(messageType string) error {
	validTypes := map[string]bool{
		"text":    true,
		"image":   true,
		"sticker": true,
		"doc":     true,
		"voice":   true,
	}

	if !validTypes[messageType] {
		return appErr.NewValidationError("invalid message type")
	}

	return nil
}
