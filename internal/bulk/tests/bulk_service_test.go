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
