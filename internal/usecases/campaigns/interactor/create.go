package interactor

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/usecases/campaigns/dto"
	retailcrmDTO "whatsapp-service/internal/usecases/retailcrm/dto"
)

// Константы для create операций
const (
	MaxCampaignNameLength = 100
	MinCampaignNameLength = 3
	MaxMessageLength      = 4096
	MaxMessagesPerHour    = 3600
	MaxAdditionalNumbers  = 1000
	MaxExcludeNumbers     = 1000
)

// Кастомные ошибки для create операций
var (
	ErrCampaignNameTooShort     = fmt.Errorf("campaign name too short: minimum %d characters", MinCampaignNameLength)
	ErrCampaignNameTooLong      = fmt.Errorf("campaign name too long: maximum %d characters", MaxCampaignNameLength)
	ErrMessageTooLong           = fmt.Errorf("message too long: maximum %d characters", MaxMessageLength)
	ErrTooManyAdditionalNumbers = fmt.Errorf("too many additional numbers: maximum %d", MaxAdditionalNumbers)
	ErrTooManyExcludeNumbers    = fmt.Errorf("too many exclude numbers: maximum %d", MaxExcludeNumbers)
)

// Create выполняет создание кампании
func (ci *CampaignInteractor) Create(ctx context.Context, req dto.CreateCampaignRequest) (*dto.CreateCampaignResponse, error) {
	if err := ci.checkActiveCampaigns(ctx); err != nil {
		return nil, err
	}

	if err := ci.validateCreateRequest(req); err != nil {
		return nil, err
	}

	campaignEntity := campaign.NewCampaign(req.Name, req.Message, req.MessagesPerHour, req.SelectedCategoryName)
	if req.Initiator != "" {
		campaignEntity.SetInitiator(req.Initiator)
	}

	phoneProcessingResult, err := ci.processPhoneNumbers(req)
	if err != nil {
		return nil, err
	}

	// Если есть фильтрация по категории, создаем кампанию со статусом "filtering"
	if req.SelectedCategoryName != "" {
		campaignEntity.SetStatus(campaign.CampaignStatusFiltering)
		// Устанавливаем временное значение total_count = 0
		campaignEntity.Metrics().Total = 0
	} else {
		// Если фильтрации нет, добавляем номера сразу
		if err := ci.addNumbersToCampaign(campaignEntity, phoneProcessingResult); err != nil {
			return nil, err
		}
	}

	if err := ci.processMediaFile(campaignEntity, req.MediaFile); err != nil {
		return nil, err
	}

	if err := ci.saveCampaignWithStatuses(ctx, campaignEntity); err != nil {
		return nil, err
	}

	if req.SelectedCategoryName != "" {
		copyPhones := func(src []*campaign.PhoneNumber) []*campaign.PhoneNumber {
			dst := make([]*campaign.PhoneNumber, len(src))
			copy(dst, src)
			return dst
		}

		asyncResult := &PhoneProcessingResult{
			FilePhones:       copyPhones(phoneProcessingResult.FilePhones),
			AdditionalPhones: copyPhones(phoneProcessingResult.AdditionalPhones),
			ExcludePhones:    copyPhones(phoneProcessingResult.ExcludePhones),
			InvalidCount:     phoneProcessingResult.InvalidCount,
			TotalTargets:     phoneProcessingResult.TotalTargets,
		}

		go ci.processCategoryFilteringAsync(campaignEntity.ID(), asyncResult, req.SelectedCategoryName, req.AutoStartAfterFilter)
	}

	return ci.buildCreateResponse(campaignEntity, phoneProcessingResult), nil
}

// processCategoryFilteringAsync асинхронно обрабатывает фильтрацию по категории
func (ci *CampaignInteractor) processCategoryFilteringAsync(campaignID string, result *PhoneProcessingResult, categoryName string, autoStartAfterFilter bool) {
	ci.logger.Info("campaign interactor: starting async category filtering",
		"campaign_id", campaignID,
		"category_name", categoryName,
		"total_numbers", len(result.FilePhones)+len(result.AdditionalPhones),
	)

	ctx := context.Background()

	if err := ci.filterByCategory(ctx, result, categoryName); err != nil {
		ci.logger.Error("campaign interactor: async category filtering failed",
			"error", err,
			"campaign_id", campaignID,
			"category_name", categoryName,
		)

		if err := ci.campaignRepo.UpdateStatus(ctx, campaignID, campaign.CampaignStatusFailed); err != nil {
			ci.logger.Error("campaign interactor: failed to update campaign status to failed",
				"error", err,
				"campaign_id", campaignID,
			)
		}
		return
	}

	campaignEntity, err := ci.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		ci.logger.Error("campaign interactor: failed to get campaign for async update",
			"error", err,
			"campaign_id", campaignID,
		)
		return
	}

	if err := ci.addNumbersToCampaign(campaignEntity, result); err != nil {
		if err == campaign.ErrNoPhoneNumbers {
			ci.logger.Info("campaign interactor: no phone numbers found after category filtering",
				"campaign_id", campaignID,
				"category_name", categoryName,
			)
			campaignEntity.Metrics().Total = 0
			result.TotalTargets = 0
			campaignEntity.SetStatus(campaign.CampaignStatusFailed)
			ci.logger.Info("campaign interactor: setting campaign status to failed",
				"campaign_id", campaignID,
				"status", campaignEntity.Status(),
			)
		} else {
			ci.logger.Error("campaign interactor: failed to add filtered numbers to campaign",
				"error", err,
				"campaign_id", campaignID,
			)
			return
		}
	} else {
		campaignEntity.SetStatus(campaign.CampaignStatusPending)
		ci.logger.Info("campaign interactor: setting campaign status to pending",
			"campaign_id", campaignID,
			"status", campaignEntity.Status(),
			"total_targets", result.TotalTargets,
		)
	}

	if err := ci.campaignRepo.Update(ctx, campaignEntity); err != nil {
		ci.logger.Error("campaign interactor: failed to update campaign after filtering",
			"error", err,
			"campaign_id", campaignID,
		)
		return
	}

	if err := ci.campaignRepo.UpdateStatus(ctx, campaignID, campaignEntity.Status()); err != nil {
		ci.logger.Error("campaign interactor: failed to update campaign status after filtering",
			"error", err,
			"campaign_id", campaignID,
			"status", campaignEntity.Status(),
		)
		return
	}

	// Сохраняем статусы номеров (только если есть номера)
	if result.TotalTargets > 0 {
		targetNumbers := campaignEntity.Audience().AllTargets()
		statuses := make([]*campaign.CampaignPhoneStatus, 0, len(targetNumbers))

		for _, phone := range targetNumbers {
			status := campaign.NewCampaignStatus(campaignID, phone.Value())
			statuses = append(statuses, status)
		}

		if err := ci.saveCampaignStatuses(ctx, statuses); err != nil {
			ci.logger.Error("campaign interactor: failed to save campaign statuses after filtering",
				"error", err,
				"campaign_id", campaignID,
			)
			return
		}

		for _, status := range statuses {
			campaignEntity.Delivery().Add(status)
		}
	} else {
		ci.logger.Info("campaign interactor: no phone numbers to save statuses for",
			"campaign_id", campaignID,
		)
	}

	ci.logger.Info("campaign interactor: async category filtering completed successfully",
		"campaign_id", campaignID,
		"category_name", categoryName,
		"filtered_total", result.TotalTargets,
	)

	if autoStartAfterFilter && result.TotalTargets > 0 && campaignEntity.Status() == campaign.CampaignStatusPending {
		ci.logger.Info("campaign interactor: auto-starting campaign after filtering",
			"campaign_id", campaignID,
			"auto_start_after_filter", autoStartAfterFilter,
		)

		startReq := dto.StartCampaignRequest{
			CampaignID: campaignID,
		}

		if _, err := ci.Start(ctx, startReq); err != nil {
			ci.logger.Error("campaign interactor: failed to auto-start campaign after filtering",
				"error", err,
				"campaign_id", campaignID,
			)
		} else {
			ci.logger.Info("campaign interactor: campaign auto-started successfully after filtering",
				"campaign_id", campaignID,
			)
		}
	}
}

// filterByCategory фильтрует номера по выбранной категории
func (ci *CampaignInteractor) filterByCategory(ctx context.Context, result *PhoneProcessingResult, categoryName string) error {
	ci.logger.Info("campaign interactor: filtering phone numbers by category",
		"category_name", categoryName,
		"total_numbers", len(result.FilePhones)+len(result.AdditionalPhones),
	)

	allPhones := make([]string, 0, len(result.FilePhones)+len(result.AdditionalPhones))

	for _, phone := range result.FilePhones {
		allPhones = append(allPhones, phone.Value())
	}

	for _, phone := range result.AdditionalPhones {
		allPhones = append(allPhones, phone.Value())
	}

	ci.logger.Info("campaign interactor: sending phones for filtering",
		"total_phones", len(allPhones),
		"file_phones", len(result.FilePhones),
		"additional_phones", len(result.AdditionalPhones),
		"phone_numbers", allPhones,
	)

	if len(allPhones) == 0 {
		ci.logger.Warn("campaign interactor: no phone numbers to filter")
		return nil
	}

	filterRequest := retailcrmDTO.FilterCustomersByCategoryRequest{
		PhoneNumbers:         allPhones,
		SelectedCategoryName: categoryName,
	}

	filterResponse, err := ci.retailCRMUseCase.FilterCustomersByCategory(ctx, filterRequest)
	if err != nil {
		ci.logger.Error("campaign interactor: failed to filter customers by category",
			"error", err,
			"category_name", categoryName,
		)
		return fmt.Errorf("failed to filter customers by category: %w", err)
	}

	shouldSendCount := filterResponse.ShouldSendCount
	filterResults := filterResponse.Results

	ci.logger.Info("campaign interactor: category filtering completed",
		"category_name", categoryName,
		"total_numbers", len(allPhones),
		"should_send_count", shouldSendCount,
	)

	filteredFilePhones := make([]*campaign.PhoneNumber, 0)
	filteredAdditionalPhones := make([]*campaign.PhoneNumber, 0)

	filteredPhonesMap := make(map[string]bool)
	for _, filterResult := range filterResults {
		ci.logger.Debug("campaign interactor: processing filter result",
			"phone_number", filterResult.PhoneNumber,
			"should_send", filterResult.ShouldSend,
		)
		if filterResult.ShouldSend {
			filteredPhonesMap[filterResult.PhoneNumber] = true
			ci.logger.Debug("campaign interactor: found matching phone number",
				"phone_number", filterResult.PhoneNumber,
				"should_send", filterResult.ShouldSend,
			)
		}
	}

	ci.logger.Info("campaign interactor: filtering results summary",
		"total_results", len(filterResults),
		"should_send_count", shouldSendCount,
		"filtered_phones_map_size", len(filteredPhonesMap),
		"filtered_phones", getMapKeys(filteredPhonesMap),
	)

	for _, phone := range result.FilePhones {
		phoneValue := phone.Value()
		if filteredPhonesMap[phoneValue] {
			filteredFilePhones = append(filteredFilePhones, phone)
			ci.logger.Debug("campaign interactor: added file phone to filtered list",
				"phone_number", phoneValue,
			)
		} else {
			ci.logger.Debug("campaign interactor: file phone not in filtered list",
				"phone_number", phoneValue,
			)
		}
	}

	for _, phone := range result.AdditionalPhones {
		phoneValue := phone.Value()
		if filteredPhonesMap[phoneValue] {
			filteredAdditionalPhones = append(filteredAdditionalPhones, phone)
			ci.logger.Debug("campaign interactor: added additional phone to filtered list",
				"phone_number", phoneValue,
			)
		} else {
			ci.logger.Debug("campaign interactor: additional phone not in filtered list",
				"phone_number", phoneValue,
			)
		}
	}

	result.FilePhones = filteredFilePhones
	result.AdditionalPhones = filteredAdditionalPhones
	// TotalTargets будет рассчитан в addNumbersToCampaign после применения исключения номеров

	ci.logger.Info("campaign interactor: phone numbers filtered by category",
		"category_name", categoryName,
		"original_total", len(allPhones),
		"filtered_file_phones", len(filteredFilePhones),
		"filtered_additional_phones", len(filteredAdditionalPhones),
	)

	return nil
}

func (ci *CampaignInteractor) checkActiveCampaigns(ctx context.Context) error {
	activeCampaigns, err := ci.campaignRepo.GetActiveCampaigns(ctx)
	if err != nil {
		return fmt.Errorf("failed to check active campaigns: %w", err)
	}

	if len(activeCampaigns) > 0 {
		return campaign.ErrCampaignAlreadyRunning
	}

	return nil
}

// PhoneProcessingResult содержит результаты обработки номеров
type PhoneProcessingResult struct {
	FilePhones       []*campaign.PhoneNumber
	AdditionalPhones []*campaign.PhoneNumber
	ExcludePhones    []*campaign.PhoneNumber
	InvalidCount     int
	TotalTargets     int
}

// processPhoneNumbers обрабатывает все телефонные номера из запроса
func (ci *CampaignInteractor) processPhoneNumbers(req dto.CreateCampaignRequest) (*PhoneProcessingResult, error) {
	result := &PhoneProcessingResult{}

	if req.PhoneFile != nil {
		filePhones, err := ci.parsePhoneFile(req.PhoneFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse phone file: %w", err)
		}
		result.FilePhones = filePhones
	}

	result.AdditionalPhones, result.InvalidCount = ci.parsePhoneStrings(req.AdditionalNumbers)

	result.ExcludePhones, _ = ci.parsePhoneStrings(req.ExcludeNumbers)

	return result, nil
}

// parsePhoneFile парсит номера из файла
func (ci *CampaignInteractor) parsePhoneFile(file *multipart.FileHeader) ([]*campaign.PhoneNumber, error) {
	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open phone file: %w", err)
	}
	defer f.Close()

	phones, err := ci.fileParser.ParsePhoneNumbers(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse phone numbers: %w", err)
	}

	result := make([]*campaign.PhoneNumber, len(phones))
	for i, phone := range phones {
		result[i] = &phone
	}

	return result, nil
}

// parsePhoneStrings парсит номера из массива строк
func (ci *CampaignInteractor) parsePhoneStrings(phoneStrings []string) ([]*campaign.PhoneNumber, int) {
	var phones []*campaign.PhoneNumber
	var invalidCount int

	for _, phoneStr := range phoneStrings {
		phone, err := campaign.NewPhoneNumber(phoneStr)
		if err != nil {
			invalidCount++
			continue
		}
		phones = append(phones, phone)
	}

	return phones, invalidCount
}

// addNumbersToCampaign добавляет номера в кампанию
func (ci *CampaignInteractor) addNumbersToCampaign(campaignEntity *campaign.Campaign, result *PhoneProcessingResult) error {
	if len(result.FilePhones) > 0 {
		if err := campaignEntity.AddPhoneNumbers(result.FilePhones); err != nil {
			return fmt.Errorf("failed to add file phone numbers: %w", err)
		}
	}

	if len(result.AdditionalPhones) > 0 {
		campaignEntity.AddAdditionalNumbers(result.AdditionalPhones)
	}

	if len(result.ExcludePhones) > 0 {
		campaignEntity.AddExcludedNumbers(result.ExcludePhones)
	}

	targetNumbers := campaignEntity.Audience().AllTargets()
	if len(targetNumbers) == 0 {
		return campaign.ErrNoPhoneNumbers
	}

	campaignEntity.Metrics().Total = len(targetNumbers)
	result.TotalTargets = len(targetNumbers)

	return nil
}

// processMediaFile обрабатывает медиа-файл
func (ci *CampaignInteractor) processMediaFile(c *campaign.Campaign, mediaFile *multipart.FileHeader) error {
	if mediaFile == nil {
		return nil
	}

	mediaData, err := ci.parseMediaFile(mediaFile)
	if err != nil {
		return fmt.Errorf("failed to parse media file: %w", err)
	}

	media := campaign.NewMedia(mediaFile.Filename, mediaFile.Header.Get("Content-Type"), mediaData)
	if !media.IsValid() {
		return fmt.Errorf("invalid media file: unsupported format")
	}

	c.SetMedia(media)
	return nil
}

// saveCampaignWithStatuses сохраняет кампанию и создает статусы в транзакции
func (ci *CampaignInteractor) saveCampaignWithStatuses(ctx context.Context, campaignEntity *campaign.Campaign) error {
	if err := ci.campaignRepo.Save(ctx, campaignEntity); err != nil {
		ci.logger.Error("Failed to save campaign to DB", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignEntity.ID(),
		})
		return fmt.Errorf("failed to save campaign: %w", err)
	}

	// Для кампаний с фильтрацией номера будут сохранены после завершения фильтрации
	if campaignEntity.Status() == campaign.CampaignStatusFiltering {
		return nil
	}

	targetNumbers := campaignEntity.Audience().AllTargets()
	statuses := make([]*campaign.CampaignPhoneStatus, 0, len(targetNumbers))

	for _, phone := range targetNumbers {
		status := campaign.NewCampaignStatus(campaignEntity.ID(), phone.Value())
		statuses = append(statuses, status)
	}

	if err := ci.saveCampaignStatuses(ctx, statuses); err != nil {
		return fmt.Errorf("failed to save campaign statuses: %w", err)
	}

	for _, status := range statuses {
		campaignEntity.Delivery().Add(status)
	}

	return nil
}

// saveCampaignStatuses сохраняет статусы пакетно
func (ci *CampaignInteractor) saveCampaignStatuses(ctx context.Context, statuses []*campaign.CampaignPhoneStatus) error {
	// Если нет статусов для сохранения, это нормально
	if len(statuses) == 0 {
		return nil
	}

	var lastErr error
	successCount := 0

	for i, status := range statuses {
		if i%100 == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		if err := ci.campaignRepo.SavePhoneStatus(ctx, status); err != nil {
			ci.logger.Error("Failed to save campaign status", map[string]interface{}{
				"error":      err.Error(),
				"campaignID": status.CampaignID(),
				"phone":      status.PhoneNumber(),
			})
			lastErr = err
			continue
		}
		successCount++
	}

	if successCount == 0 && lastErr != nil {
		return fmt.Errorf("failed to save any campaign statuses: %w", lastErr)
	}

	if successCount < len(statuses) {
		ci.logger.Warn("Some campaign statuses failed to save", map[string]interface{}{
			"total":   len(statuses),
			"success": successCount,
			"failed":  len(statuses) - successCount,
		})
	}

	return nil
}

// buildCreateResponse строит ответ на запрос создания кампании
func (ci *CampaignInteractor) buildCreateResponse(campaignEntity *campaign.Campaign, result *PhoneProcessingResult) *dto.CreateCampaignResponse {
	response := &dto.CreateCampaignResponse{
		Campaign:      campaignEntity,
		ValidPhones:   result.TotalTargets,
		InvalidPhones: result.InvalidCount,
		TotalNumbers:  result.TotalTargets,
		Warnings:      make([]string, 0),
	}

	// Если кампания в статусе filtering, показываем информацию о фильтрации
	if campaignEntity.Status() == campaign.CampaignStatusFiltering {
		response.TotalNumbers = 0 // Показываем 0, так как фильтрация еще не завершена
		response.Warnings = append(response.Warnings,
			fmt.Sprintf("Фильтрация по категории '%s' выполняется в фоне", campaignEntity.CategoryName()))
	} else {
		// Для кампаний без фильтрации показываем обычную статистику
		if len(result.ExcludePhones) > 0 {
			response.Warnings = append(response.Warnings,
				fmt.Sprintf("Исключено %d номеров", len(result.ExcludePhones)))
		}

		if result.InvalidCount > 0 {
			response.Warnings = append(response.Warnings,
				fmt.Sprintf("Пропущено %d невалидных номеров", result.InvalidCount))
		}
	}

	ci.logger.Info("Successfully created campaign", map[string]interface{}{
		"campaignID":   campaignEntity.ID(),
		"name":         campaignEntity.Name(),
		"status":       campaignEntity.Status(),
		"totalNumbers": response.TotalNumbers,
		"invalidCount": result.InvalidCount,
		"excludeCount": len(result.ExcludePhones),
	})

	return response
}

// validateCreateRequest проверяет валидность запроса с детальной валидацией
func (ci *CampaignInteractor) validateCreateRequest(req dto.CreateCampaignRequest) error {
	if req.Name == "" {
		return campaign.ErrCampaignNameRequired
	}
	if len(req.Name) < MinCampaignNameLength {
		return ErrCampaignNameTooShort
	}
	if len(req.Name) > MaxCampaignNameLength {
		return ErrCampaignNameTooLong
	}

	if req.Message == "" {
		return campaign.ErrCampaignMessageRequired
	}
	if len(req.Message) > MaxMessageLength {
		return ErrMessageTooLong
	}

	if req.PhoneFile == nil && len(req.AdditionalNumbers) == 0 {
		return campaign.ErrNoPhoneNumbers
	}

	if req.MessagesPerHour < 0 || req.MessagesPerHour > MaxMessagesPerHour {
		return campaign.ErrInvalidMessagesPerHour
	}

	if len(req.AdditionalNumbers) > MaxAdditionalNumbers {
		return ErrTooManyAdditionalNumbers
	}
	if len(req.ExcludeNumbers) > MaxExcludeNumbers {
		return ErrTooManyExcludeNumbers
	}

	return nil
}

// parseMediaFile парсит медиа-файл из multipart
func (ci *CampaignInteractor) parseMediaFile(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// getMapKeys возвращает ключи из map для логирования
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
