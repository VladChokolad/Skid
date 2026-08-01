package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/VladChokolad/Skid/Backend/internal/calculator"
)

func (h *Handler) GetSettlementsHandler(w http.ResponseWriter, r *http.Request) {
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}
	payments, err := h.storage.GetPaymentsByPartyID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Невозможно получить данные из базы данных")
		return
	}
	purchases, err := h.storage.GetPurchasesByPartyID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Невозможно получить данные из базы данных")
		return
	}
	participants, err := h.storage.GetParticipantsByPartyID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Невозможно получить данные из базы данных")
		return
	}
	debts, err := h.storage.GetDebtsByPartyID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Невозможно получить данные из базы данных")
		return
	}

	var lightPayments []calculator.LightPayment
	var lightPurchases []calculator.LightPurchase
	var lightParticipants []calculator.LightParticipant
	var lightDebts []calculator.LightDebt

	for _, v := range payments {
		if !v.IsConfirmed {
			continue // пропускаем неподтверждённые
		}
		lightPayments = append(lightPayments, calculator.LightPayment{
			FromParticipantID: v.FromParticipantID,
			ToParticipantID:   v.ToParticipantID,
			Amount:            v.Amount,
		})
	}
	for _, v := range purchases {
		lightPurchases = append(lightPurchases, calculator.LightPurchase{
			BuyerID:   v.BuyerID,
			Price:     v.Price,
			SplitType: v.SplitType,
		})
	}

	for _, v := range participants {
		lightParticipants = append(lightParticipants, calculator.LightParticipant{
			ID:   v.ID,
			Name: v.Name,
		})
	}

	for _, v := range debts {
		lightDebts = append(lightDebts, calculator.LightDebt{
			PurchaseID:    v.PurchaseID,
			ParticipantID: v.ParticipantID,
			SplitValue:    v.SplitValue,
		})
	}

	importData := calculator.ImportToCalc{
		LightPayments:     lightPayments,
		LightPurchases:    lightPurchases,
		LightParticipants: lightParticipants,
		LightDebts:        lightDebts,
	}
	outputData, err := calculator.Calculation(importData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "ошибка расчёта")
		return
	}
	sendSuccessResponse(w, http.StatusOK, "Успешное присоединение к вечеринке", map[string]interface{}{
		"Calculation": outputData,
	})

}
