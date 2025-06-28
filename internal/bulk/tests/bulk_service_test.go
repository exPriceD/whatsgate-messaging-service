package bulk_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"whatsapp-service/internal/bulk/domain"
	"whatsapp-service/internal/bulk/mocks"
	"whatsapp-service/internal/bulk/usecase"
	"whatsapp-service/internal/logger"
)

// createMultipartFileHeader создаёт multipart форму в памяти и возвращает *multipart.FileHeader
func createMultipartFileHeader(filename, content string, t *testing.T) *multipart.FileHeader {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	partHeader.Set("Content-Type", "application/octet-stream")
	part, err := w.CreatePart(partHeader)
	require.NoError(t, err)
	_, err = io.Copy(part, strings.NewReader(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	r := multipart.NewReader(&b, w.Boundary())
	form, err := r.ReadForm(1024 * 1024)
	require.NoError(t, err)
	files := form.File["file"]
	require.NotEmpty(t, files)
	return files[0]
}

func setupTestService() (*usecase.BulkService, *mocks.MockWhatsGateClient, *mocks.MockFileParser, *mocks.MockBulkCampaignStorage, *mocks.MockBulkCampaignStatusStorage) {
	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})
	client := &mocks.MockWhatsGateClient{}
	parser := &mocks.MockFileParser{}
	campaignStorage := mocks.NewMockBulkCampaignStorage()
	statusStorage := mocks.NewMockBulkCampaignStatusStorage()

	service := &usecase.BulkService{
		Logger:          log,
		WhatsGateClient: client,
		FileParser:      parser,
		CampaignStorage: campaignStorage,
		StatusStorage:   statusStorage,
	}

	return service, client, parser, campaignStorage, statusStorage
}

func TestBulkService_HandleBulkSendMultipart_TextMessage(t *testing.T) {
	service, client, parser, campaignStorage, statusStorage := setupTestService()

	// Настройка моков
	parser.ParsePhonesFromExcelFunc = func(filePath string, columnName string) ([]string, error) {
		return []string{"71234567890", "79876543210"}, nil
	}

	client.SendTextMessageFunc = func(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error) {
		return domain.SingleSendResult{
			PhoneNumber: phoneNumber,
			Success:     true,
			Status:      "sent",
		}, nil
	}

	// Тестовые параметры
	params := domain.BulkSendParams{
		Message:         "Test message",
		Async:           false,
		MessagesPerHour: 10,
		NumbersFile:     createMultipartFileHeader("numbers.xlsx", "test", t),
		MediaFile:       nil,
	}

	// Выполнение теста
	result, err := service.HandleBulkSendMultipart(context.Background(), params)

	// Проверки
	require.NoError(t, err)
	assert.True(t, result.Started)
	assert.Equal(t, 2, result.Total)
	assert.Contains(t, result.Message, "started in background")

	// Проверяем, что кампания создана
	campaigns := campaignStorage.Campaigns
	assert.Len(t, campaigns, 1)
	for _, campaign := range campaigns {
		assert.Equal(t, "Test message", campaign.Message)
		assert.Equal(t, 2, campaign.Total)
		assert.Equal(t, "started", campaign.Status)
		assert.Equal(t, 10, campaign.MessagesPerHour)
	}

	// Проверяем, что статусы созданы
	statuses := statusStorage.Statuses
	assert.Len(t, statuses, 2)
	for _, status := range statuses {
		assert.Equal(t, "pending", status.Status)
		assert.Contains(t, []string{"71234567890", "79876543210"}, status.PhoneNumber)
	}
}

func TestBulkService_HandleBulkSendMultipart_MediaMessage(t *testing.T) {
	service, client, parser, campaignStorage, _ := setupTestService()

	// Настройка моков
	parser.ParsePhonesFromExcelFunc = func(filePath string, columnName string) ([]string, error) {
		return []string{"71234567890"}, nil
	}

	client.SendMediaMessageFunc = func(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (domain.SingleSendResult, error) {
		return domain.SingleSendResult{
			PhoneNumber: phoneNumber,
			Success:     true,
			Status:      "sent",
		}, nil
	}

	// Тестовые параметры с медиа
	params := domain.BulkSendParams{
		Message:         "Test media message",
		Async:           false,
		MessagesPerHour: 5,
		NumbersFile:     createMultipartFileHeader("numbers.xlsx", "test", t),
		MediaFile:       createMultipartFileHeader("test.jpg", "fake-image-data", t),
	}

	// Выполнение теста
	result, err := service.HandleBulkSendMultipart(context.Background(), params)

	// Проверки
	require.NoError(t, err)
	assert.True(t, result.Started)
	assert.Equal(t, 1, result.Total)

	// Проверяем, что кампания создана с медиа
	campaigns := campaignStorage.Campaigns
	assert.Len(t, campaigns, 1)
	for _, campaign := range campaigns {
		assert.Equal(t, "Test media message", campaign.Message)
		assert.Equal(t, 1, campaign.Total)
		assert.Equal(t, "started", campaign.Status)
		assert.Equal(t, 5, campaign.MessagesPerHour)
		assert.NotNil(t, campaign.MediaFilename)
		assert.NotNil(t, campaign.MediaMime)
		assert.NotNil(t, campaign.MediaType)
	}
}

func TestBulkService_HandleBulkSendMultipart_ValidationErrors(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	tests := []struct {
		name    string
		params  domain.BulkSendParams
		wantErr string
	}{
		{
			name: "missing_numbers_file",
			params: domain.BulkSendParams{
				Message:         "Test",
				Async:           false,
				MessagesPerHour: 10,
				NumbersFile:     nil,
			},
			wantErr: "BULK_FILE_PARSE_ERROR",
		},
		{
			name: "invalid_messages_per_hour",
			params: domain.BulkSendParams{
				Message:         "Test",
				Async:           false,
				MessagesPerHour: 0,
				NumbersFile:     createMultipartFileHeader("numbers.xlsx", "test", t),
			},
			wantErr: "BULK_RATE_LIMIT_EXCEEDED",
		},
		{
			name: "negative_messages_per_hour",
			params: domain.BulkSendParams{
				Message:         "Test",
				Async:           false,
				MessagesPerHour: -5,
				NumbersFile:     createMultipartFileHeader("numbers.xlsx", "test", t),
			},
			wantErr: "BULK_RATE_LIMIT_EXCEEDED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.HandleBulkSendMultipart(context.Background(), tt.params)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestBulkService_HandleBulkSendMultipart_ParserError(t *testing.T) {
	service, _, parser, _, _ := setupTestService()

	// Настройка мока для возврата ошибки
	parser.ParsePhonesFromExcelFunc = func(filePath string, columnName string) ([]string, error) {
		return nil, errors.New("parser error")
	}

	params := domain.BulkSendParams{
		Message:         "Test message",
		Async:           false,
		MessagesPerHour: 10,
		NumbersFile:     createMultipartFileHeader("numbers.xlsx", "test", t),
	}

	_, err := service.HandleBulkSendMultipart(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parser error")
}

func TestBulkService_HandleBulkSendMultipart_EmptyPhones(t *testing.T) {
	service, _, parser, _, _ := setupTestService()

	// Настройка мока для возврата пустого списка
	parser.ParsePhonesFromExcelFunc = func(filePath string, columnName string) ([]string, error) {
		return []string{}, nil
	}

	params := domain.BulkSendParams{
		Message:         "Test message",
		Async:           false,
		MessagesPerHour: 10,
		NumbersFile:     createMultipartFileHeader("numbers.xlsx", "test", t),
	}

	_, err := service.HandleBulkSendMultipart(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BULK_NO_VALID_NUMBERS")
}

func TestBulkService_HandleBulkSendMultipart_ClientError(t *testing.T) {
	service, client, parser, _, _ := setupTestService()

	// Настройка моков
	parser.ParsePhonesFromExcelFunc = func(filePath string, columnName string) ([]string, error) {
		return []string{"71234567890"}, nil
	}

	client.SendTextMessageFunc = func(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error) {
		return domain.SingleSendResult{}, errors.New("client error")
	}

	params := domain.BulkSendParams{
		Message:         "Test message",
		Async:           false,
		MessagesPerHour: 10,
		NumbersFile:     createMultipartFileHeader("numbers.xlsx", "test", t),
	}

	// Выполнение теста
	result, err := service.HandleBulkSendMultipart(context.Background(), params)

	// Проверки - сервис должен запуститься, несмотря на ошибки клиента
	require.NoError(t, err)
	assert.True(t, result.Started)
	assert.Equal(t, 1, result.Total)
}

func TestBulkService_HandleBulkSendMultipart_InvalidFileType(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	// Тест с неправильным типом файла
	params := domain.BulkSendParams{
		Message:         "Test message",
		Async:           false,
		MessagesPerHour: 10,
		NumbersFile:     "invalid-type", // Не *multipart.FileHeader
	}

	_, err := service.HandleBulkSendMultipart(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BULK_FILE_PARSE_ERROR")
}

func TestCancelCampaign_Success(t *testing.T) {
	service, _, _, campaignStorage, _ := setupTestService()

	// Создаем тестовую кампанию
	campaign := &domain.BulkCampaign{
		ID:     "test-campaign",
		Name:   "Test Campaign",
		Status: domain.CampaignStatusStarted,
	}
	campaignStorage.Create(campaign)

	// Отменяем кампанию
	err := service.CancelCampaign(context.Background(), "test-campaign")
	assert.NoError(t, err)

	// Проверяем, что статус изменился
	updatedCampaign, _ := campaignStorage.GetByID("test-campaign")
	assert.Equal(t, domain.CampaignStatusCancelled, updatedCampaign.Status)
}

func TestCancelCampaign_NotFound(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	// Пытаемся отменить несуществующую кампанию
	err := service.CancelCampaign(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "campaign not found")
}

func TestCancelCampaign_AlreadyFinished(t *testing.T) {
	service, _, _, campaignStorage, _ := setupTestService()

	// Создаем завершенную кампанию
	campaign := &domain.BulkCampaign{
		ID:     "finished-campaign",
		Name:   "Finished Campaign",
		Status: domain.CampaignStatusFinished,
	}
	campaignStorage.Create(campaign)

	// Пытаемся отменить завершенную кампанию
	err := service.CancelCampaign(context.Background(), "finished-campaign")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be cancelled")

	// Проверяем, что статус не изменился
	updatedCampaign, _ := campaignStorage.GetByID("finished-campaign")
	assert.Equal(t, domain.CampaignStatusFinished, updatedCampaign.Status)
}

func TestCancelCampaign_AlreadyCancelled(t *testing.T) {
	service, _, _, campaignStorage, _ := setupTestService()

	// Создаем уже отмененную кампанию
	campaign := &domain.BulkCampaign{
		ID:     "cancelled-campaign",
		Name:   "Cancelled Campaign",
		Status: domain.CampaignStatusCancelled,
	}
	campaignStorage.Create(campaign)

	// Пытаемся отменить уже отмененную кампанию
	err := service.CancelCampaign(context.Background(), "cancelled-campaign")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be cancelled")

	// Проверяем, что статус не изменился
	updatedCampaign, _ := campaignStorage.GetByID("cancelled-campaign")
	assert.Equal(t, domain.CampaignStatusCancelled, updatedCampaign.Status)
}

func TestCancelCampaign_UpdatesPendingStatuses(t *testing.T) {
	service, _, _, campaignStorage, statusStorage := setupTestService()

	// Создаем тестовую кампанию
	campaign := &domain.BulkCampaign{
		ID:     "test-campaign-statuses",
		Name:   "Test Campaign Statuses",
		Status: domain.CampaignStatusStarted,
	}
	campaignStorage.Create(campaign)

	// Создаем статусы для номеров
	statuses := []*domain.BulkCampaignStatus{
		{
			CampaignID:  "test-campaign-statuses",
			PhoneNumber: "1234567890",
			Status:      domain.CampaignStatusPending,
		},
		{
			CampaignID:  "test-campaign-statuses",
			PhoneNumber: "0987654321",
			Status:      domain.CampaignStatusPending,
		},
		{
			CampaignID:  "test-campaign-statuses",
			PhoneNumber: "5555555555",
			Status:      "sent", // Уже отправленное сообщение
		},
	}

	for _, status := range statuses {
		statusStorage.Create(status)
	}

	// Отменяем кампанию
	err := service.CancelCampaign(context.Background(), "test-campaign-statuses")
	assert.NoError(t, err)

	// Проверяем, что статус кампании изменился
	updatedCampaign, _ := campaignStorage.GetByID("test-campaign-statuses")
	assert.Equal(t, domain.CampaignStatusCancelled, updatedCampaign.Status)

	// Проверяем, что pending статусы изменились на cancelled
	allStatuses, _ := statusStorage.ListByCampaignID("test-campaign-statuses")
	assert.Len(t, allStatuses, 3)

	pendingCount := 0
	cancelledCount := 0
	sentCount := 0

	for _, status := range allStatuses {
		switch status.Status {
		case domain.CampaignStatusPending:
			pendingCount++
		case domain.CampaignStatusCancelled:
			cancelledCount++
		case "sent":
			sentCount++
		}
	}

	assert.Equal(t, 0, pendingCount, "Should be no pending statuses")
	assert.Equal(t, 2, cancelledCount, "Should have 2 cancelled statuses")
	assert.Equal(t, 1, sentCount, "Should have 1 sent status")
}
