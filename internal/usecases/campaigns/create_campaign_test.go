package campaigns

import (
	"context"
	"errors"
	"io"
	"testing"

	"whatsapp-service/internal/entities"
	entityErrors "whatsapp-service/internal/entities/errors"
	"whatsapp-service/internal/infrastructure/parsers/types"
)

// MockCampaignRepository мок репозитория кампаний
type MockCampaignRepository struct {
	SaveFunc    func(ctx context.Context, campaign *entities.Campaign) error
	GetByIDFunc func(ctx context.Context, id string) (*entities.Campaign, error)
	UpdateFunc  func(ctx context.Context, campaign *entities.Campaign) error
	campaigns   map[string]*entities.Campaign
}

// NewMockCampaignRepository создает новый мок репозитория
func NewMockCampaignRepository() *MockCampaignRepository {
	return &MockCampaignRepository{
		campaigns: make(map[string]*entities.Campaign),
	}
}

// Save сохраняет кампанию
func (m *MockCampaignRepository) Save(ctx context.Context, campaign *entities.Campaign) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, campaign)
	}
	m.campaigns[campaign.ID()] = campaign
	return nil
}

// GetByID получает кампанию по ID
func (m *MockCampaignRepository) GetByID(ctx context.Context, id string) (*entities.Campaign, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	campaign, exists := m.campaigns[id]
	if !exists {
		return nil, entityErrors.ErrCampaignNotFound
	}
	return campaign, nil
}

// Update обновляет кампанию
func (m *MockCampaignRepository) Update(ctx context.Context, campaign *entities.Campaign) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, campaign)
	}
	m.campaigns[campaign.ID()] = campaign
	return nil
}

// List возвращает список кампаний
func (m *MockCampaignRepository) List(ctx context.Context, limit, offset int) ([]*entities.Campaign, error) {
	return nil, nil
}

// Delete удаляет кампанию
func (m *MockCampaignRepository) Delete(ctx context.Context, id string) error {
	delete(m.campaigns, id)
	return nil
}

// IncrementProcessedCount увеличивает счетчик обработанных
func (m *MockCampaignRepository) IncrementProcessedCount(ctx context.Context, id string) error {
	return nil
}

// IncrementErrorCount увеличивает счетчик ошибок
func (m *MockCampaignRepository) IncrementErrorCount(ctx context.Context, id string) error {
	return nil
}

// UpdateStatus обновляет статус
func (m *MockCampaignRepository) UpdateStatus(ctx context.Context, id string, status entities.CampaignStatus) error {
	return nil
}

// UpdateProcessedCount обновляет счетчик обработанных
func (m *MockCampaignRepository) UpdateProcessedCount(ctx context.Context, id string, processedCount int) error {
	return nil
}

// GetActiveCampaigns возвращает активные кампании
func (m *MockCampaignRepository) GetActiveCampaigns(ctx context.Context) ([]*entities.Campaign, error) {
	return nil, nil
}

// InitTable инициализирует таблицы
func (m *MockCampaignRepository) InitTable(ctx context.Context) error {
	return nil
}

// MockCampaignStatusRepository мок репозитория статусов
type MockCampaignStatusRepository struct {
	SaveFunc func(ctx context.Context, status *entities.CampaignPhoneStatus) error
	statuses map[string]*entities.CampaignPhoneStatus
}

// NewMockCampaignStatusRepository создает новый мок репозитория статусов
func NewMockCampaignStatusRepository() *MockCampaignStatusRepository {
	return &MockCampaignStatusRepository{
		statuses: make(map[string]*entities.CampaignPhoneStatus),
	}
}

// Save сохраняет статус
func (m *MockCampaignStatusRepository) Save(ctx context.Context, status *entities.CampaignPhoneStatus) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, status)
	}
	m.statuses[status.ID()] = status
	return nil
}

// GetByID получает статус по ID
func (m *MockCampaignStatusRepository) GetByID(ctx context.Context, id string) (*entities.CampaignPhoneStatus, error) {
	status, exists := m.statuses[id]
	if !exists {
		return nil, entityErrors.ErrCampaignNotFound
	}
	return status, nil
}

// Update обновляет статус
func (m *MockCampaignStatusRepository) Update(ctx context.Context, status *entities.CampaignPhoneStatus) error {
	m.statuses[status.ID()] = status
	return nil
}

// Delete удаляет статус
func (m *MockCampaignStatusRepository) Delete(ctx context.Context, id string) error {
	delete(m.statuses, id)
	return nil
}

// ListByCampaignID возвращает статусы по ID кампании
func (m *MockCampaignStatusRepository) ListByCampaignID(ctx context.Context, campaignID string) ([]*entities.CampaignPhoneStatus, error) {
	var result []*entities.CampaignPhoneStatus
	for _, status := range m.statuses {
		if status.CampaignID() == campaignID {
			result = append(result, status)
		}
	}
	return result, nil
}

// UpdateStatusesByCampaignID обновляет статусы по ID кампании
func (m *MockCampaignStatusRepository) UpdateStatusesByCampaignID(ctx context.Context, campaignID string, oldStatus, newStatus entities.CampaignStatusType) error {
	return nil
}

// MarkAsSent помечает как отправленный
func (m *MockCampaignStatusRepository) MarkAsSent(ctx context.Context, id string) error {
	return nil
}

// MarkAsFailed помечает как неудачный
func (m *MockCampaignStatusRepository) MarkAsFailed(ctx context.Context, id string, errorMsg string) error {
	return nil
}

// MarkAsCancelled помечает как отмененный
func (m *MockCampaignStatusRepository) MarkAsCancelled(ctx context.Context, id string) error {
	return nil
}

// GetSentNumbersByCampaignID возвращает отправленные номера
func (m *MockCampaignStatusRepository) GetSentNumbersByCampaignID(ctx context.Context, campaignID string) ([]string, error) {
	return nil, nil
}

// GetFailedStatusesByCampaignID возвращает неудачные статусы
func (m *MockCampaignStatusRepository) GetFailedStatusesByCampaignID(ctx context.Context, campaignID string) ([]*entities.CampaignPhoneStatus, error) {
	return nil, nil
}

// CountStatusesByCampaignID подсчитывает статусы
func (m *MockCampaignStatusRepository) CountStatusesByCampaignID(ctx context.Context, campaignID string, status entities.CampaignStatusType) (int, error) {
	return 0, nil
}

// InitTable инициализирует таблицы
func (m *MockCampaignStatusRepository) InitTable(ctx context.Context) error {
	return nil
}

// MockFileParser мок файлового парсера
type MockFileParser struct {
	phones []entities.PhoneNumber
	err    error
}

// NewMockFileParser создает новый мок парсера
func NewMockFileParser() *MockFileParser {
	phone, _ := entities.NewPhoneNumber("79161234567")
	return &MockFileParser{
		phones: []entities.PhoneNumber{*phone},
		err:    nil,
	}
}

func (m *MockFileParser) ParsePhoneNumbers(content io.Reader) ([]entities.PhoneNumber, error) {
	return m.phones, m.err
}

func (m *MockFileParser) ParsePhoneNumbersDetailed(content io.Reader, columnName string) (*types.ParseResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &types.ParseResult{
		ValidPhones:     m.phones,
		InvalidPhones:   []types.InvalidPhone{},
		DuplicatePhones: []types.DuplicatePhone{},
		Statistics: types.ParseStatistics{
			TotalRows:   len(m.phones),
			ValidCount:  len(m.phones),
			UniqueCount: len(m.phones),
		},
		Warnings: []string{},
	}, nil
}

func (m *MockFileParser) SupportedExtensions() map[string]struct{} {
	return map[string]struct{}{
		".xlsx": {},
		".xls":  {},
	}
}

func (m *MockFileParser) IsSupported(filename string) bool {
	return true
}

// TestCreateCampaignUseCase_Execute проверяет успешное создание кампании
func TestCreateCampaignUseCase_Execute(t *testing.T) {
	mockRepo := NewMockCampaignRepository()
	mockStatusRepo := NewMockCampaignStatusRepository()
	mockParser := NewMockFileParser()
	useCase := NewCreateCampaignUseCase(mockRepo, mockStatusRepo, mockParser)

	req := CreateCampaignRequest{
		Name:              "Тестовая рассылка",
		Message:           "Привет от нашей компании!",
		AdditionalNumbers: []string{"79099876543"},
		MessagesPerHour:   100,
	}

	response, err := useCase.Execute(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Campaign.Name() != "Тестовая рассылка" {
		t.Errorf("Expected campaign name 'Тестовая рассылка', got %s", response.Campaign.Name())
	}

	if response.Campaign.Message() != "Привет от нашей компании!" {
		t.Errorf("Expected campaign message 'Привет от нашей компании!', got %s", response.Campaign.Message())
	}

	if response.Campaign.Status() != entities.CampaignStatusPending {
		t.Errorf("Expected campaign status 'pending', got %s", response.Campaign.Status())
	}

	if response.ValidPhones != 1 {
		t.Errorf("Expected 1 valid phone, got %d", response.ValidPhones)
	}

	if len(mockRepo.campaigns) != 1 {
		t.Errorf("Expected 1 campaign in repository, got %d", len(mockRepo.campaigns))
	}
}

// TestCreateCampaignUseCase_Execute_WithInvalidRate проверяет обработку невалидного лимита сообщений
func TestCreateCampaignUseCase_Execute_WithInvalidRate(t *testing.T) {
	mockRepo := NewMockCampaignRepository()
	mockStatusRepo := NewMockCampaignStatusRepository()
	mockParser := NewMockFileParser()
	useCase := NewCreateCampaignUseCase(mockRepo, mockStatusRepo, mockParser)

	req := CreateCampaignRequest{
		Name:            "Тестовая рассылка",
		Message:         "Привет!",
		MessagesPerHour: -1,
	}

	response, err := useCase.Execute(context.Background(), req)

	if err == nil {
		t.Error("Expected error for invalid messages per hour, got nil")
	}

	if response != nil {
		t.Error("Expected nil response for invalid input")
	}
}

// TestCreateCampaignUseCase_Execute_RepositoryError проверяет обработку ошибки репозитория
func TestCreateCampaignUseCase_Execute_RepositoryError(t *testing.T) {
	mockRepo := NewMockCampaignRepository()
	mockRepo.SaveFunc = func(ctx context.Context, campaign *entities.Campaign) error {
		return entityErrors.ErrRepositoryError
	}

	mockStatusRepo := NewMockCampaignStatusRepository()
	mockParser := NewMockFileParser()
	useCase := NewCreateCampaignUseCase(mockRepo, mockStatusRepo, mockParser)

	req := CreateCampaignRequest{
		Name:              "Тестовая рассылка",
		Message:           "Привет!",
		AdditionalNumbers: []string{"79161234567"},
	}

	response, err := useCase.Execute(context.Background(), req)

	if !errors.Is(err, entityErrors.ErrRepositoryError) {
		t.Errorf("Expected repository error, got %v", err)
	}

	if response != nil {
		t.Error("Expected nil response for repository error")
	}
}
