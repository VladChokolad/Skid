package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/VladChokolad/Skid/Backend/internal/calculator"
)

func (h *Handler) CalculateResultHandler(w http.ResponseWriter, r *http.Request) {
	var req int
	//докодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	payments, err := h.storage.GetPaymentsByPartyID(req)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Невозможно получить данные из базы данных")
		return
	}
	purchases, err := h.storage.GetPurchasesByPartyID(req)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Невозможно получить данные из базы данных")
		return
	}
	participants, err := h.storage.GetParticipantsByPartyID(req)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Невозможно получить данные из базы данных")
		return
	}
	debts, err := h.storage.GetDebtsByPartyID(req)
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
