package settings

import (
	"net/http"
	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"
	"whatsapp-service/internal/whatsgate/interfaces"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSettingsHandler возвращает gin.HandlerFunc с внедрённым сервисом
func GetSettingsHandler(whatsgateService *whatsgateDomain.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		settings := whatsgateService.GetSettings()
		log.Info("Get settings", zap.String("whatsapp_id", settings.WhatsappID))
		response := WhatGateSettings{
			WhatsappID: settings.WhatsappID,
			APIKey:     settings.APIKey,
			BaseURL:    settings.BaseURL,
		}
		c.JSON(http.StatusOK, response)
	}
}

// UpdateSettingsHandler возвращает gin.HandlerFunc с внедрённым сервисом
func UpdateSettingsHandler(whatsgateService *whatsgateDomain.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		var request WhatGateSettings
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Invalid request body", zap.Error(err))
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		whatsgateSettings := &interfaces.Settings{
			WhatsappID: request.WhatsappID,
			APIKey:     request.APIKey,
			BaseURL:    request.BaseURL,
		}
		if err := whatsgateService.UpdateSettings(whatsgateSettings); err != nil {
			log.Error("Failed to update settings", zap.Error(err))
			c.Error(err)
			return
		}
		log.Info("Settings updated", zap.String("whatsapp_id", request.WhatsappID))
		c.JSON(http.StatusOK, request)
	}
}

// ResetSettingsHandler возвращает gin.HandlerFunc с внедрённым сервисом
func ResetSettingsHandler(whatsgateService *whatsgateDomain.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		if err := whatsgateService.ResetSettings(); err != nil {
			log.Error("Failed to reset settings", zap.Error(err))
			c.Error(appErr.New("RESET_ERROR", "Failed to reset settings", err))
			return
		}
		log.Info("Settings reset")
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Settings reset successfully",
		})
	}
}
