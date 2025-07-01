package campaigns

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/entities/errors"
	"whatsapp-service/internal/usecases/interfaces"
)

// CreateCampaignUseCase обрабатывает создание кампаний массовой рассылки
type CreateCampaignUseCase struct {
	campaignRepo       interfaces.CampaignRepository
	campaignStatusRepo interfaces.CampaignStatusRepository
	fileParser         interfaces.FileParser
}

// NewCreateCampaignUseCase создает новый экземпляр use case
func NewCreateCampaignUseCase(
	campaignRepo interfaces.CampaignRepository,
	campaignStatusRepo interfaces.CampaignStatusRepository,
	fileParser interfaces.FileParser,
) *CreateCampaignUseCase {
	return &CreateCampaignUseCase{
		campaignRepo:       campaignRepo,
		campaignStatusRepo: campaignStatusRepo,
		fileParser:         fileParser,
	}
}

// CreateCampaignRequest представляет запрос на создание кампании
type CreateCampaignRequest struct {
	Name              string                // Название кампании
	Message           string                // Текст сообщения
	PhoneFile         *multipart.FileHeader // Excel файл с номерами
	MediaFile         *multipart.FileHeader // Медиа-файл (опционально)
	AdditionalNumbers []string              // Дополнительные номера
	ExcludeNumbers    []string              // Номера для исключения
	MessagesPerHour   int                   // Лимит сообщений в час
	Initiator         string                // Инициатор кампании
	Async             bool                  // Асинхронное выполнение
}

// CreateCampaignResponse представляет ответ на создание кампании
type CreateCampaignResponse struct {
	Campaign       *entities.Campaign // Созданная кампания
	ValidPhones    int                // Количество валидных номеров
	InvalidPhones  int                // Количество невалидных номеров
	DuplicateCount int                // Количество дубликатов
	TotalNumbers   int                // Общее количество номеров после обработки
	Warnings       []string           // Предупреждения
}

// Execute выполняет создание кампании
func (uc *CreateCampaignUseCase) Execute(ctx context.Context, req CreateCampaignRequest) (*CreateCampaignResponse, error) {
	// 1. Проверяем активные кампании
	activeCampaigns, err := uc.campaignRepo.GetActiveCampaigns(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check active campaigns: %w", err)
	}

	if len(activeCampaigns) > 0 {
		return nil, errors.ErrCampaignAlreadyRunning
	}

	// 2. Валидация базовых параметров
	if err := uc.validateRequest(req); err != nil {
		return nil, err
	}

	// 3. Создаем кампанию
	campaignID := generateID()
	campaign := entities.NewCampaign(campaignID, req.Name, req.Message)

	// 4. Устанавливаем дополнительные параметры
	if req.Initiator != "" {
		campaign.SetInitiator(req.Initiator)
	}

	if req.MessagesPerHour > 0 {
		if err := campaign.SetMessagesPerHour(req.MessagesPerHour); err != nil {
			return nil, err
		}
	}

	// 5. Парсим номера из файла
	var filePhones []*entities.PhoneNumber
	var invalidCount int

	if req.PhoneFile != nil {
		// Открываем файл
		file, err := req.PhoneFile.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open phone file: %w", err)
		}
		defer file.Close()

		// Парсим номера
		phones, err := uc.fileParser.ParsePhoneNumbers(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse phone file: %w", err)
		}

		// Конвертируем в указатели
		filePhones = make([]*entities.PhoneNumber, len(phones))
		for i, phone := range phones {
			filePhones[i] = &phone
		}
	}

	// 6. Парсим дополнительные номера
	var additionalPhones []*entities.PhoneNumber
	for _, phoneStr := range req.AdditionalNumbers {
		phone, err := entities.NewPhoneNumber(phoneStr)
		if err != nil {
			invalidCount++
			continue
		}
		additionalPhones = append(additionalPhones, phone)
	}

	// 7. Парсим исключаемые номера
	var excludePhones []*entities.PhoneNumber
	for _, phoneStr := range req.ExcludeNumbers {
		phone, err := entities.NewPhoneNumber(phoneStr)
		if err != nil {
			continue // Игнорируем невалидные номера в исключениях
		}
		excludePhones = append(excludePhones, phone)
	}

	// 8. Добавляем номера в кампанию
	if len(filePhones) > 0 {
		campaign.AddPhoneNumbers(filePhones)
	}

	if len(additionalPhones) > 0 {
		campaign.AddAdditionalNumbers(additionalPhones)
	}

	if len(excludePhones) > 0 {
		campaign.AddExcludeNumbers(excludePhones)
	}

	// 9. Проверяем, что есть номера для отправки
	targetNumbers := campaign.GetAllTargetNumbers()
	if len(targetNumbers) == 0 {
		return nil, errors.ErrNoPhoneNumbers
	}

	// 10. Обрабатываем медиа-файл
	if req.MediaFile != nil {
		mediaData, err := uc.parseMediaFile(req.MediaFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse media file: %w", err)
		}
		campaign.SetMedia(req.MediaFile.Filename, req.MediaFile.Header.Get("Content-Type"), mediaData)
	}

	// 11. Сохраняем кампанию
	if err := uc.campaignRepo.Save(ctx, campaign); err != nil {
		return nil, fmt.Errorf("failed to save campaign: %w", err)
	}

	// 12. Создаем статусы для всех номеров
	for _, phone := range targetNumbers {
		status := entities.NewCampaignStatus(campaignID, phone.Value())
		if err := uc.campaignStatusRepo.Save(ctx, status); err != nil {
			// Логируем ошибку, но продолжаем
			continue
		}
		campaign.AddCampaignStatus(status)
	}

	// 13. Подготавливаем ответ
	response := &CreateCampaignResponse{
		Campaign:      campaign,
		ValidPhones:   len(targetNumbers),
		InvalidPhones: invalidCount,
		TotalNumbers:  len(targetNumbers),
		Warnings:      make([]string, 0),
	}

	// Добавляем предупреждения
	if len(excludePhones) > 0 {
		response.Warnings = append(response.Warnings,
			fmt.Sprintf("Исключено %d номеров", len(excludePhones)))
	}

	if invalidCount > 0 {
		response.Warnings = append(response.Warnings,
			fmt.Sprintf("Пропущено %d невалидных номеров", invalidCount))
	}

	return response, nil
}

// validateRequest проверяет валидность запроса
func (uc *CreateCampaignUseCase) validateRequest(req CreateCampaignRequest) error {
	if req.Name == "" {
		return errors.ErrCampaignNameRequired
	}

	if req.Message == "" {
		return errors.ErrCampaignMessageRequired
	}

	if req.PhoneFile == nil && len(req.AdditionalNumbers) == 0 {
		return errors.ErrNoPhoneNumbers
	}

	if req.MessagesPerHour < 0 || req.MessagesPerHour > 3600 {
		return errors.ErrInvalidMessagesPerHour
	}

	// Валидация файла
	if req.PhoneFile != nil {
		if !uc.fileParser.IsSupported(req.PhoneFile.Filename) {
			return fmt.Errorf("unsupported file type: %s", req.PhoneFile.Filename)
		}
	}

	return nil
}

// parseMediaFile парсит медиа-файл из multipart
func (uc *CreateCampaignUseCase) parseMediaFile(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Читаем содержимое файла
	data := make([]byte, file.Size)
	_, err = src.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// generateID генерирует уникальный идентификатор кампании
func generateID() string {
	return fmt.Sprintf("campaign_%d", time.Now().UnixNano())
}
