package infra

import (
	"context"
	"whatsapp-service/1 old/internal/parser"
	domain "whatsapp-service/internal/bulk/domain"
	"whatsapp-service/internal/logger"
	usecase "whatsapp-service/internal/whatsgate/usecase"
)

// WhatsGateClientAdapter — адаптер для клиента WhatsGate
type WhatsGateClientAdapter struct {
	Service *usecase.SettingsUsecase
}

func (a *WhatsGateClientAdapter) SendTextMessage(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error) {
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

func (a *WhatsGateClientAdapter) SendMediaMessage(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (domain.SingleSendResult, error) {
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

func (a *FileParserAdapter) CountRowsInExcel(filePath string) (int, error) {
	result, err := parser.ParsePhonesFromExcel(filePath, "Телефон", a.Logger)
	if err != nil {
		return 0, err
	}

	return len(result.ValidPhones), nil
}
