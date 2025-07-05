package settings

import (
	"net/http"
	httpTypes "whatsapp-service/internal/delivery/http/types"
	appErrors "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	"whatsapp-service/internal/whatsgate/domain"
	whatsgateService "whatsapp-service/internal/whatsgate/usecase"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSettingsHandler возвращает текущие настройки
// @Summary Получить настройки
// @Description Возвращает текущие настройки WhatsGate
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} SettingsResponse "Настройки получены"
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /settings [get]
func GetSettingsHandler(ws *whatsgateService.SettingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		settings := ws.GetSettings()
		log.Info("Get settings", zap.String("whatsapp_id", settings.WhatsappID))
		response := WhatsGateSettings{
			WhatsappID: settings.WhatsappID,
			APIKey:     settings.APIKey,
			BaseURL:    settings.BaseURL,
		}
		c.JSON(http.StatusOK, response)
	}
}

// UpdateSettingsHandler обновляет настройки
// @Summary Обновить настройки
// @Description Обновляет настройки WhatsGate
// @Tags settings
// @Accept json
// @Produce json
// @Param settings body UpdateSettingsRequest true "Новые настройки"
// @Success 200 {object} types.SuccessResponse "Настройки обновлены"
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /settings [put]
func UpdateSettingsHandler(ws *whatsgateService.SettingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		var request WhatsGateSettings
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Invalid request body", zap.Error(err))
			c.Error(appErrors.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		whatsgateSettings := &domain.Settings{
			WhatsappID: request.WhatsappID,
			APIKey:     request.APIKey,
			BaseURL:    request.BaseURL,
		}
		if err := ws.UpdateSettings(whatsgateSettings); err != nil {
			log.Error("Failed to update settings", zap.Error(err))
			c.Error(err)
			return
		}
		log.Info("Settings updated", zap.String("whatsapp_id", request.WhatsappID))
		c.JSON(http.StatusOK, request)
	}
}

// ResetSettingsHandler godoc
// @Summary Сбросить настройки WhatGate
// @Description Сбрасывает настройки WhatGate API (удаляет сохраненные данные)
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} types.SuccessResponse "OK"
// @Failure 500 {object} types.AppErrorResponse "Внутренняя ошибка сервера"
// @Router /settings/reset [delete]
func ResetSettingsHandler(ws *whatsgateService.SettingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		if err := ws.ResetSettings(); err != nil {
			log.Error("Failed to reset settings", zap.Error(err))
			c.Error(appErrors.New(appErrors.ErrorTypeConfiguration, "RESET_ERROR", "Failed to reset settings", err))
			return
		}
		log.Info("Settings reset")
		c.JSON(http.StatusOK, httpTypes.SuccessResponse{
			Success: true,
			Message: "Settings reset successfully",
		})
	}
}
