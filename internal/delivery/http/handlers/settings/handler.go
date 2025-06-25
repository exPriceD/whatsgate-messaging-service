package settings

import (
	"net/http"
	appErr "whatsapp-service/internal/errors"
	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"
	"whatsapp-service/internal/whatsgate/interfaces"

	"github.com/gin-gonic/gin"
)

// GetSettingsHandler возвращает gin.HandlerFunc с внедрённым сервисом
func GetSettingsHandler(whatsgateService *whatsgateDomain.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings := whatsgateService.GetSettings()
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
		var request WhatGateSettings
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		whatsgateSettings := &interfaces.Settings{
			WhatsappID: request.WhatsappID,
			APIKey:     request.APIKey,
			BaseURL:    request.BaseURL,
		}
		if err := whatsgateService.UpdateSettings(whatsgateSettings); err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, request)
	}
}

// ResetSettingsHandler возвращает gin.HandlerFunc с внедрённым сервисом
func ResetSettingsHandler(whatsgateService *whatsgateDomain.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := whatsgateService.ResetSettings(); err != nil {
			c.Error(appErr.New("RESET_ERROR", "Failed to reset settings", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Settings reset successfully",
		})
	}
}
