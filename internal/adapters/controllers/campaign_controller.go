package controllers

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"whatsapp-service/internal/usecases/campaigns"
	"whatsapp-service/internal/usecases/interfaces"
)

// CampaignController HTTP контроллер для работы с кампаниями
type CampaignController struct {
	createCampaignUC *campaigns.CreateCampaignUseCase
	startCampaignUC  *campaigns.StartCampaignUseCase
	cancelCampaignUC *campaigns.CancelCampaignUseCase
	campaignRepo     interfaces.CampaignRepository
}

// NewCampaignController создает новый экземпляр контроллера
func NewCampaignController(
	createCampaignUC *campaigns.CreateCampaignUseCase,
	startCampaignUC *campaigns.StartCampaignUseCase,
	cancelCampaignUC *campaigns.CancelCampaignUseCase,
	campaignRepo interfaces.CampaignRepository,
) *CampaignController {
	return &CampaignController{
		createCampaignUC: createCampaignUC,
		startCampaignUC:  startCampaignUC,
		cancelCampaignUC: cancelCampaignUC,
		campaignRepo:     campaignRepo,
	}
}

// CreateCampaign создает новую кампанию
func (c *CampaignController) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Получаем основные параметры
	name := r.FormValue("name")
	message := r.FormValue("message")
	initiator := r.FormValue("initiator")
	messagesPerHourStr := r.FormValue("messages_per_hour")

	if name == "" || message == "" {
		http.Error(w, "Name and message are required", http.StatusBadRequest)
		return
	}

	messagesPerHour := 60 // значение по умолчанию
	if messagesPerHourStr != "" {
		if mph, err := strconv.Atoi(messagesPerHourStr); err == nil && mph > 0 {
			messagesPerHour = mph
		}
	}

	// Получаем файл с номерами
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Получаем медиа файл (опционально)
	var mediaFileHeader *multipart.FileHeader
	if _, mediaHeader, err := r.FormFile("media"); err == nil {
		mediaFileHeader = mediaHeader
	}

	// Создаем запрос
	req := campaigns.CreateCampaignRequest{
		Name:              name,
		Message:           message,
		PhoneFile:         fileHeader,
		MediaFile:         mediaFileHeader,
		MessagesPerHour:   messagesPerHour,
		Initiator:         initiator,
		AdditionalNumbers: parseArrayParam(r, "additional_numbers"),
		ExcludeNumbers:    parseArrayParam(r, "exclude_numbers"),
	}

	// Выполняем use case
	response, err := c.createCampaignUC.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartCampaign запускает кампанию
func (c *CampaignController) StartCampaign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID из URL path
	campaignID := extractIDFromPath(r.URL.Path, "/campaigns/", "/start")
	if campaignID == "" {
		http.Error(w, "Campaign ID is required", http.StatusBadRequest)
		return
	}

	// Парсим тело запроса
	var requestBody struct {
		Async bool `json:"async"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		requestBody.Async = true // По умолчанию асинхронно
	}

	req := campaigns.StartCampaignRequest{
		CampaignID: campaignID,
		Async:      requestBody.Async,
	}

	response, err := c.startCampaignUC.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelCampaign отменяет кампанию
func (c *CampaignController) CancelCampaign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID из URL path
	campaignID := extractIDFromPath(r.URL.Path, "/campaigns/", "/cancel")
	if campaignID == "" {
		http.Error(w, "Campaign ID is required", http.StatusBadRequest)
		return
	}

	req := campaigns.CancelCampaignRequest{
		CampaignID: campaignID,
		Reason:     r.FormValue("reason"),
	}

	response, err := c.cancelCampaignUC.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCampaign получает кампанию по ID
func (c *CampaignController) GetCampaign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID из URL path
	campaignID := extractIDFromPath(r.URL.Path, "/campaigns/", "")
	if campaignID == "" {
		http.Error(w, "Campaign ID is required", http.StatusBadRequest)
		return
	}

	campaign, err := c.campaignRepo.GetByID(r.Context(), campaignID)
	if err != nil {
		http.Error(w, "Campaign not found", http.StatusNotFound)
		return
	}

	// Преобразуем в DTO
	response := map[string]interface{}{
		"id":                campaign.ID(),
		"name":              campaign.Name(),
		"message":           campaign.Message(),
		"status":            string(campaign.Status()),
		"total_count":       campaign.TotalCount(),
		"error_count":       campaign.ErrorCount(),
		"messages_per_hour": campaign.MessagesPerHour(),
		"initiator":         campaign.Initiator(),
		"created_at":        campaign.CreatedAt(),
		"progress":          campaign.GetProgress(),
	}

	if campaign.Media() != nil {
		response["media"] = map[string]interface{}{
			"filename":     campaign.Media().Filename(),
			"mime_type":    campaign.Media().MimeType(),
			"message_type": string(campaign.Media().MessageType()),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListCampaigns возвращает список кампаний
func (c *CampaignController) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим параметры пагинации
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // по умолчанию
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	campaigns, err := c.campaignRepo.List(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Failed to get campaigns", http.StatusInternalServerError)
		return
	}

	// Преобразуем в DTO
	var response []map[string]interface{}
	for _, campaign := range campaigns {
		item := map[string]interface{}{
			"id":                campaign.ID(),
			"name":              campaign.Name(),
			"message":           campaign.Message(),
			"status":            string(campaign.Status()),
			"total_count":       campaign.TotalCount(),
			"error_count":       campaign.ErrorCount(),
			"messages_per_hour": campaign.MessagesPerHour(),
			"initiator":         campaign.Initiator(),
			"created_at":        campaign.CreatedAt(),
			"progress":          campaign.GetProgress(),
		}

		if campaign.Media() != nil {
			item["media"] = map[string]interface{}{
				"filename":     campaign.Media().Filename(),
				"mime_type":    campaign.Media().MimeType(),
				"message_type": string(campaign.Media().MessageType()),
			}
		}

		response = append(response, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"campaigns": response,
		"limit":     limit,
		"offset":    offset,
		"total":     len(response),
	})
}

// parseArrayParam парсит массив параметров из формы
func parseArrayParam(r *http.Request, paramName string) []string {
	if values, ok := r.Form[paramName]; ok {
		var result []string
		for _, value := range values {
			if value != "" {
				result = append(result, value)
			}
		}
		return result
	}
	return nil
}

// extractIDFromPath извлекает ID из URL path
func extractIDFromPath(path, prefix, suffix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	idPart := strings.TrimPrefix(path, prefix)
	if suffix != "" {
		if !strings.HasSuffix(idPart, suffix) {
			return ""
		}
		idPart = strings.TrimSuffix(idPart, suffix)
	}

	// Убираем trailing slash если есть
	idPart = strings.TrimSuffix(idPart, "/")

	return idPart
}
