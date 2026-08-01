package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/VladChokolad/Skid/Backend/internal/middleware"
	"github.com/VladChokolad/Skid/Backend/internal/objects"
	"github.com/go-chi/chi/v5"
)

type ParticipantRequest struct {
	Name string `json:"name"`
}

func (h *Handler) GetParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	participants, err := h.storage.GetParticipantsByPartyID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении участников")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Список участников", participants)
}

func (h *Handler) CreateParticipantHandler(w http.ResponseWriter, r *http.Request) {
	// Только админ или владелец может добавлять placeholder
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	// Проверить что пользователь участник и имеет права
	participant, err := h.storage.GetParticipantByUserAndPartyID(userID, partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusForbidden, "Вы не участник этой тусовки")
		return
	}
	if !participant.IsAdmin {
		party, err := h.storage.GetPartyByID(partyID)
		if err != nil || party.OwnerID != userID {
			sendErrorResponse(w, http.StatusForbidden, "Недостаточно прав")
			return
		}
	}

	var req ParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	if req.Name == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Имя обязательно")
		return
	}

	// Создаём placeholder
	placeholder := objects.Participant{
		PartyID:       partyID,
		Name:          req.Name,
		IsPlaceholder: true,
	}

	placeholderID, err := h.storage.CreateParticipant(placeholder)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при создании участника")
		return
	}

	sendSuccessResponse(w, http.StatusCreated, "Участник добавлен", map[string]int{"id": placeholderID})
}

func (h *Handler) ReplaceParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	// Аноним или user занимает место placeholder
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	participantID, err := strconv.Atoi(chi.URLParam(r, "participantID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID участника")
		return
	}

	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	isAnon := middleware.GetIsAnonymousFromContext(r.Context())

	// Найти placeholder
	placeholder, err := h.storage.GetParticipantByID(participantID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Участник не найден")
		return
	}
	if !placeholder.IsPlaceholder {
		sendErrorResponse(w, http.StatusBadRequest, "Этот участник не является placeholder")
		return
	}
	if placeholder.PartyID != partyID {
		sendErrorResponse(w, http.StatusForbidden, "Участник не принадлежит этой тусовке")
		return
	}

	// Обновить placeholder — привязать реального пользователя
	placeholder.UserOrAnonymousID = &userID
	placeholder.IsAnonymous = isAnon
	placeholder.IsPlaceholder = false

	if err := h.storage.UpdateParticipant(placeholder); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при обновлении участника")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Вы успешно заняли место участника", nil)
}

func (h *Handler) DeleteParticipantHandler(w http.ResponseWriter, r *http.Request) {
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	participantID, err := strconv.Atoi(chi.URLParam(r, "participantID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID участника")
		return
	}

	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	// Проверить права — только владелец или админ
	party, err := h.storage.GetPartyByID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Тусовка не найдена")
		return
	}

	participant, err := h.storage.GetParticipantByUserAndPartyID(userID, partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusForbidden, "Вы не участник этой тусовки")
		return
	}

	if party.OwnerID != userID && !participant.IsAdmin {
		sendErrorResponse(w, http.StatusForbidden, "Недостаточно прав")
		return
	}

	if err := h.storage.DeleteParticipant(participantID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении участника")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Участник удалён", nil)
}
