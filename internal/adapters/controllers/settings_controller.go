package controllers

import (
	"encoding/json"
	"net/http"

	"whatsapp-service/internal/usecases/settings"
)

type SettingsController struct {
	service *settings.Service
}

func NewSettingsController(svc *settings.Service) *SettingsController {
	return &SettingsController{service: svc}
}

func (c *SettingsController) GetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s, err := c.service.Get(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(s)
}

func (c *SettingsController) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		WhatsappID string `json:"whatsapp_id"`
		APIKey     string `json:"api_key"`
		BaseURL    string `json:"base_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	s, err := c.service.Update(ctx, payload.WhatsappID, payload.APIKey, payload.BaseURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(s)
}

func (c *SettingsController) ResetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := c.service.Reset(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
