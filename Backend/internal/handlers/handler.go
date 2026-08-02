package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/VladChokolad/Skid/Backend/internal/config"
	"github.com/VladChokolad/Skid/Backend/internal/storage"
)

type Handler struct {
	storage *storage.Storage // доступ к БД
	cfg     config.Config    // доступ к JWTSecret и другим настройкам
}

func NewHandler(storage *storage.Storage, cfg config.Config) *Handler {
	return &Handler{storage: storage, cfg: cfg}
}

// Для возвращения ошибки
type ErrorResponse struct {
	Error string `json:"error"`
}

// Для возвращения ответа
type SuccessResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Token   string      `json:"token,omitempty"`
}

func sendErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func sendSuccessResponse(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
		"data":    data,
	})
}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60,
	})
}
