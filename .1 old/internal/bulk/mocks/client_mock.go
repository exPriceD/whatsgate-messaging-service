package mocks

import (
	"context"
	"whatsapp-service/internal/bulk/domain"
)

type MockWhatsGateClient struct {
	SendTextMessageFunc  func(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error)
	SendMediaMessageFunc func(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (domain.SingleSendResult, error)
}

func (m *MockWhatsGateClient) SendTextMessage(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error) {
	if m.SendTextMessageFunc != nil {
		return m.SendTextMessageFunc(ctx, phoneNumber, text, async)
	}
	return domain.SingleSendResult{
		PhoneNumber: phoneNumber,
		Success:     true,
		Status:      "sent",
	}, nil
}

func (m *MockWhatsGateClient) SendMediaMessage(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (domain.SingleSendResult, error) {
	if m.SendMediaMessageFunc != nil {
		return m.SendMediaMessageFunc(ctx, phoneNumber, messageType, text, filename, fileData, mimeType, async)
	}
	return domain.SingleSendResult{
		PhoneNumber: phoneNumber,
		Success:     true,
		Status:      "sent",
	}, nil
}
