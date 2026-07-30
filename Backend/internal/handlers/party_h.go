package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/VladChokolad/Skid/Backend/internal/middleware"
	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

type PartyRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	PartyImage  *string `json:"partyImage"`
}

func (h *Handler) GetMyPartiesHandler(w http.ResponseWriter, r *http.Request) { //Показывает все вечеринки пользователя
	// 1. Получаем userID и флаг анонимности
	ID, ok := middleware.GetUserIDFromContext(r.Context())
	isAnonymous := middleware.GetIsAnonymousFromContext(r.Context())

	// 2. Проверяем, авторизован ли пользователь (если требуется)
	if !ok || ID == 0 {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var parties_list []objects.Party // или конкретный тип, возвращаемый хранилищем
	var err error

	if isAnonymous {
		parties_list, err = h.storage.GetPartiesByAnonID(ID)
	} else {
		parties_list, err = h.storage.GetPartiesByUserID(ID)
	}
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Не удалось получить данные")
		return
	}
	sendSuccessResponse(w, http.StatusCreated, "Вот ваши тусовки!", map[string]interface{}{
		"data": parties_list,
	})
}
func (h *Handler) GetPartyHandler(w http.ResponseWriter, r *http.Request) { //Показывает конкретную вечеринку полностью
}
func (h *Handler) JoinPartyHandler(w http.ResponseWriter, r *http.Request) { //Присоединяет человека к вечеринке
}
func (h *Handler) CreatePartyHandler(w http.ResponseWriter, r *http.Request) { //Создаёт вечеринку
	var req PartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	// Только зарегистрированные могут создавать тусовки
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	if middleware.GetIsAnonymousFromContext(r.Context()) {
		sendErrorResponse(w, http.StatusForbidden, "Анонимы не могут создавать тусовки")
		return
	}

	// Создаём тусовку
	party := objects.Party{
		Name:    *req.Name,
		OwnerID: userID,
	}
	if req.Description != nil {
		party.Description = req.Description
	}
	if req.PartyImage != nil {
		party.PartyImage = req.PartyImage
	}

	_, err := h.storage.CreateParty(party)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при создании тусовки")
		return
	}
	sendSuccessResponse(w, http.StatusCreated, "тусовка создана!", nil)

}
func (h *Handler) UpdatePartyHandler(w http.ResponseWriter, r *http.Request) {
	// partyID из URL
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	// Только зарегистрированные
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	// Проверить что пользователь владелец
	party, err := h.storage.GetPartyByID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Тусовка не найдена")
		return
	}
	if party.OwnerID != userID {
		sendErrorResponse(w, http.StatusForbidden, "Только владелец может изменять тусовку")
		return
	}

	// Декодировать тело
	var req PartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	// Обновить
	party.Name = *req.Name
	if req.Description != nil {
		party.Description = req.Description
	}
	if req.PartyImage != nil {
		party.PartyImage = req.PartyImage
	}

	if err := h.storage.UpdateParty(party); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при обновлении тусовки")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Тусовка обновлена", nil)
}

func (h *Handler) DeletePartyHandler(w http.ResponseWriter, r *http.Request) {
	// partyID из URL
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	// Только зарегистрированные
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	// Проверить что пользователь владелец
	party, err := h.storage.GetPartyByID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Тусовка не найдена")
		return
	}
	if party.OwnerID != userID {
		sendErrorResponse(w, http.StatusForbidden, "Только владелец может удалять тусовку")
		return
	}

	// Удалить
	if err := h.storage.DeleteParty(partyID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении тусовки")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Тусовка удалена", nil)
}
