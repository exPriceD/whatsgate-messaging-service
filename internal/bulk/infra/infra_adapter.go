package infra

import (
	"context"
	domain "whatsapp-service/internal/bulk/domain"
	"whatsapp-service/internal/logger"
	"whatsapp-service/internal/parser"
	usecase "whatsapp-service/internal/whatsgate/usecase"
)

// WhatGateClientAdapter — адаптер для реального клиента WhatGate
// Использует существующий SettingsService
type WhatGateClientAdapter struct {
	Service *usecase.SettingsUsecase
}

func (a *WhatGateClientAdapter) SendTextMessage(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error) {
	client, err := a.Service.GetClient()
	if err != nil {
		return domain.SingleSendResult{PhoneNumber: phoneNumber, Success: false, Error: err.Error()}, err
	}

	resp, err := client.SendTextMessage(ctx, phoneNumber, text, async)
	if err != nil {
		return domain.SingleSendResult{PhoneNumber: phoneNumber, Success: false, Error: err.Error()}, err
	}

	return domain.SingleSendResult{PhoneNumber: phoneNumber, Success: true, Status: resp.Status}, nil
}

func (a *WhatGateClientAdapter) SendMediaMessage(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (domain.SingleSendResult, error) {
	client, err := a.Service.GetClient()
	if err != nil {
		return domain.SingleSendResult{PhoneNumber: phoneNumber, Success: false, Error: err.Error()}, err
	}

	resp, err := client.SendMediaMessage(ctx, phoneNumber, messageType, text, filename, fileData, mimeType, async)
	if err != nil {
		return domain.SingleSendResult{PhoneNumber: phoneNumber, Success: false, Error: err.Error()}, err
	}

	return domain.SingleSendResult{PhoneNumber: phoneNumber, Success: true, Status: resp.Status}, nil
}

// FileParserAdapter — адаптер для реального парсера Excel-файлов
type FileParserAdapter struct {
	Logger logger.Logger
}

func (a *FileParserAdapter) ParsePhonesFromExcel(filePath string, columnName string) ([]string, error) {
	result, err := parser.ParsePhonesFromExcel(filePath, columnName, a.Logger)
	if err != nil {
		return nil, err
	}

	return result.ValidPhones, nil
}
