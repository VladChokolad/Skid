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
	// Членство уже проверено RequirePartyMember
	party, _ := middleware.PartyFromContext(r.Context())

	participants, err := h.storage.GetParticipantsByPartyID(party.ID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении участников")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Список участников", participants)
}

func (h *Handler) CreateParticipantHandler(w http.ResponseWriter, r *http.Request) {
	// Права админа/владельца уже проверены RequirePartyMember + RequirePartyAdmin
	party, _ := middleware.PartyFromContext(r.Context())

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
		PartyID:       party.ID,
		Name:          req.Name,
		IsPlaceholder: true,
	}

	placeholderID, err := h.storage.CreateParticipant(placeholder)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при создании участника")
		return
	}

	// Пересчитать покупки "поровну всем" на новый состав тусовки — не фатально для запроса
	_ = h.recalcEqualSplitDebts(party.ID)

	sendSuccessResponse(w, http.StatusCreated, "Участник добавлен", map[string]int{"id": placeholderID})
}

func (h *Handler) DeleteParticipantHandler(w http.ResponseWriter, r *http.Request) {
	// Права админа/владельца уже проверены RequirePartyMember + RequirePartyAdmin
	party, _ := middleware.PartyFromContext(r.Context())

	participantID, err := strconv.Atoi(chi.URLParam(r, "participantID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID участника")
		return
	}

	// Убедиться что удаляемый участник принадлежит именно этой тусовке
	target, err := h.storage.GetParticipantByID(participantID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Участник не найден")
		return
	}
	if target.PartyID != party.ID {
		sendErrorResponse(w, http.StatusForbidden, "Участник не принадлежит этой тусовке")
		return
	}

	// Сначала снести долги участника (FK на participant_id без каскада —
	// иначе удаление участника с существующими долгами упадёт с ошибкой)
	if err := h.storage.DeleteDebtsByParticipantID(participantID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении долгов участника")
		return
	}

	if err := h.storage.DeleteParticipant(participantID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении участника")
		return
	}

	// Пересчитать покупки "поровну всем" на новый состав тусовки — не фатально для запроса
	_ = h.recalcEqualSplitDebts(party.ID)

	sendSuccessResponse(w, http.StatusOK, "Участник удалён", nil)
}
