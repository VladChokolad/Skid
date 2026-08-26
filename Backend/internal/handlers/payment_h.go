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
	// Членство уже проверено RequirePartyMember
	party, _ := middleware.PartyFromContext(r.Context())

	payments, err := h.storage.GetPaymentsByPartyID(party.ID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении платежей")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Список платежей", payments)
}

func (h *Handler) CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	// Членство уже проверено RequirePartyMember
	party, fromParticipant, _ := middleware.PartyAndParticipantFromContext(r.Context())

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

	// Получатель должен быть участником этой же тусовки
	toParticipant, err := h.storage.GetParticipantByID(req.ToParticipantID)
	if err != nil || toParticipant.PartyID != party.ID {
		sendErrorResponse(w, http.StatusBadRequest, "Получатель не найден в этой тусовке")
		return
	}

	payment := objects.Payment{
		PartyID:           party.ID,
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
	// Членство уже проверено RequirePartyMember
	party, participant, _ := middleware.PartyAndParticipantFromContext(r.Context())

	paymentID, err := strconv.Atoi(chi.URLParam(r, "paymentID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID платежа")
		return
	}

	// Проверить что платёж принадлежит этой тусовке и адресован этому участнику
	payment, err := h.storage.GetPaymentByID(paymentID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Платёж не найден")
		return
	}
	if payment.PartyID != party.ID {
		sendErrorResponse(w, http.StatusForbidden, "Платёж не принадлежит этой тусовке")
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
