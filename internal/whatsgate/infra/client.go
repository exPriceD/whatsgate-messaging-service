package infra

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	"whatsapp-service/internal/whatsgate/domain"

	"go.uber.org/zap"
)

type ClientImpl struct {
	baseURL    string
	whatsappID string
	apiKey     string
	httpClient *http.Client
	Logger     logger.Logger
}

func NewClient(baseURL, whatsappID, apiKey string, log logger.Logger) *ClientImpl {
	return &ClientImpl{
		baseURL:    baseURL,
		whatsappID: whatsappID,
		apiKey:     apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Logger: log,
	}
}

func (c *ClientImpl) SendTextMessage(ctx context.Context, phoneNumber, text string, async bool) (*domain.SendMessageResponse, error) {
	request := domain.SendMessageRequest{
		WhatsappID: c.whatsappID,
		Async:      async,
		Recipient: domain.Recipient{
			Number: phoneNumber,
		},
		Message: domain.Message{
			Type: "text",
			Body: text,
		},
	}
	return c.sendMessage(ctx, request)
}

func (c *ClientImpl) SendMediaMessage(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (*domain.SendMessageResponse, error) {
	encodedData := base64.StdEncoding.EncodeToString(fileData)
	request := domain.SendMessageRequest{
		WhatsappID: c.whatsappID,
		Async:      async,
		Recipient: domain.Recipient{
			Number: phoneNumber,
		},
		Message: domain.Message{
			Type: messageType,
			Body: text,
			Media: &domain.Media{
				MimeType: mimeType,
				Data:     encodedData,
				Filename: filename,
			},
		},
	}
	return c.sendMessage(ctx, request)
}

func (c *ClientImpl) sendMessage(ctx context.Context, request domain.SendMessageRequest) (*domain.SendMessageResponse, error) {
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

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	latency := time.Since(start)
	if err != nil {
		return nil, appErr.New("NETWORK_ERROR", "failed to send request", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, appErr.New("READ_ERROR", "failed to read response", err)
	}

	c.Logger.Info("WhatGate API request",
		zap.String("url", req.URL.String()),
		zap.String("whatsapp_id", c.whatsappID),
		zap.String("recipient", request.Recipient.Number),
		zap.String("status", resp.Status),
		zap.Duration("latency", latency),
	)

	switch resp.StatusCode {
	case http.StatusOK:
		c.Logger.Info("Message sent successfully to WhatGate API",
			zap.String("recipient", request.Recipient.Number),
			zap.String("status", resp.Status),
		)
		return &domain.SendMessageResponse{
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

var allowedTypes = map[string]struct{}{
	"text":    {},
	"image":   {},
	"sticker": {},
	"doc":     {},
	"voice":   {},
}

func ValidateMessageType(messageType string) error {
	if _, ok := allowedTypes[messageType]; !ok {
		return appErr.NewValidationError("invalid message type")
	}
	return nil
}
