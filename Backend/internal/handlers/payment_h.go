package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/VladChokolad/Skid/Backend/internal/middleware"
	"github.com/VladChokolad/Skid/Backend/internal/objects"
	"github.com/go-chi/chi/v5"
)

type PaymentRequest struct {
	ToParticipantID int     `json:"toParticipantID"`
	Amount          float64 `json:"amount"`
	Note            string  `json:"note"`
}

func (h *Handler) GetPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	payments, err := h.storage.GetPaymentsByPartyID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении платежей")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Список платежей", payments)
}

func (h *Handler) CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	isAnon := middleware.GetIsAnonymousFromContext(r.Context())

	// Найти participant отправителя
	var fromParticipant objects.Participant
	if isAnon {
		fromParticipant, err = h.storage.GetParticipantByAnonAndPartyID(userID, partyID)
	} else {
		fromParticipant, err = h.storage.GetParticipantByUserAndPartyID(userID, partyID)
	}
	if err != nil {
		sendErrorResponse(w, http.StatusForbidden, "Вы не участник этой тусовки")
		return
	}

	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.ToParticipantID == 0 {
		sendErrorResponse(w, http.StatusBadRequest, "Укажите получателя")
		return
	}
	if req.Amount <= 0 {
		sendErrorResponse(w, http.StatusBadRequest, "Сумма должна быть больше нуля")
		return
	}
	if fromParticipant.ID == req.ToParticipantID {
		sendErrorResponse(w, http.StatusBadRequest, "Нельзя отправить платёж самому себе")
		return
	}

	payment := objects.Payment{
		PartyID:           partyID,
		FromParticipantID: fromParticipant.ID,
		ToParticipantID:   req.ToParticipantID,
		Amount:            req.Amount,
		Note:              req.Note,
		IsConfirmed:       false,
	}

	paymentID, err := h.storage.CreatePayment(payment)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при создании платежа")
		return
	}

	sendSuccessResponse(w, http.StatusCreated, "Платёж создан", map[string]int{"id": paymentID})
}

func (h *Handler) ConfirmPaymentHandler(w http.ResponseWriter, r *http.Request) {
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	paymentID, err := strconv.Atoi(chi.URLParam(r, "paymentID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID платежа")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	isAnon := middleware.GetIsAnonymousFromContext(r.Context())

	// Найти participant получателя
	var participant objects.Participant
	if isAnon {
		participant, err = h.storage.GetParticipantByAnonAndPartyID(userID, partyID)
	} else {
		participant, err = h.storage.GetParticipantByUserAndPartyID(userID, partyID)
	}
	if err != nil {
		sendErrorResponse(w, http.StatusForbidden, "Вы не участник этой тусовки")
		return
	}

	// Проверить что платёж адресован этому участнику
	payment, err := h.storage.GetPaymentByID(paymentID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Платёж не найден")
		return
	}
	if payment.ToParticipantID != participant.ID {
		sendErrorResponse(w, http.StatusForbidden, "Только получатель может подтвердить платёж")
		return
	}

	if err := h.storage.ConfirmPayment(paymentID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при подтверждении платежа")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Платёж подтверждён", nil)
}
