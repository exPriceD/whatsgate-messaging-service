package handlers

import (
	"net/http"

	"whatsapp-service/internal/delivery/http/types"
	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/whatsgate/interfaces"

	"github.com/gin-gonic/gin"
)

// GetSettingsHandler возвращает текущие настройки WhatGate
// @Summary Get WhatGate settings
// @Description Возвращает текущие настройки WhatGate API
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} types.WhatGateSettings
// @Router /settings [get]
func (s *Server) GetSettingsHandler(c *gin.Context) {
	settings := s.whatsgateService.GetSettings()

	response := types.WhatGateSettings{
		WhatsappID: settings.WhatsappID,
		APIKey:     settings.APIKey,
		BaseURL:    settings.BaseURL,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSettingsHandler обновляет настройки WhatGate
// @Summary Update WhatGate settings
// @Description Обновляет настройки WhatGate API
// @Tags settings
// @Accept json
// @Produce json
// @Param settings body types.WhatGateSettings true "WhatGate settings"
// @Success 200 {object} types.WhatGateSettings
// @Failure 400 {object} types.ErrorResponse
// @Router /settings [put]
func (s *Server) UpdateSettingsHandler(c *gin.Context) {
	var request types.WhatGateSettings
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
		return
	}

	whatsgateSettings := &interfaces.Settings{
		WhatsappID: request.WhatsappID,
		APIKey:     request.APIKey,
		BaseURL:    request.BaseURL,
	}

	if err := s.whatsgateService.UpdateSettings(whatsgateSettings); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, request)
}

// ResetSettingsHandler сбрасывает настройки WhatGate
// @Summary Reset WhatGate settings
// @Description Сбрасывает настройки WhatGate API (удаляет сохраненные данные)
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} types.SuccessResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /settings/reset [delete]
func (s *Server) ResetSettingsHandler(c *gin.Context) {
	if err := s.whatsgateService.ResetSettings(); err != nil {
		c.Error(appErr.New("RESET_ERROR", "Failed to reset settings", err))
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Settings reset successfully",
	})
}
