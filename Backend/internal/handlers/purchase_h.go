package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/VladChokolad/Skid/Backend/internal/middleware"
	"github.com/VladChokolad/Skid/Backend/internal/objects"
	"github.com/go-chi/chi/v5"
)

type PurchaseRequest struct {
	Name           string          `json:"name"`
	Description    *string         `json:"description"`
	PurchaseIconID *int            `json:"purchaseIconId"`
	Price          float64         `json:"price"`
	SplitType      int             `json:"splitType"`
	DateofPurchase *time.Time      `json:"dateOfPurchase"`
	Debtors        []int           `json:"debtors"`       // id участников — для splitType 0/1
	DebtorAmounts  map[int]float64 `json:"debtorAmounts"` // id участника -> сумма/вес — для splitType 2/3
}

func (h *Handler) EchoHandler(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)
	sendSuccessResponse(w, http.StatusOK, "Я получил", body)
}

func (h *Handler) GetPurchasesHandler(w http.ResponseWriter, r *http.Request) {
	// Членство уже проверено RequirePartyMember
	party, _ := middleware.PartyFromContext(r.Context())

	purchases, err := h.storage.GetPurchasesByPartyID(party.ID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении покупок")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Список покупок", purchases)
}

func (h *Handler) CreatePurchaseHandler(w http.ResponseWriter, r *http.Request) {
	// Членство уже проверено RequirePartyMember
	party, participant, _ := middleware.PartyAndParticipantFromContext(r.Context())

	// Декодировать тело
	var req PurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	// 5. Валидация
	if req.Name == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Название обязательно")
		return
	}
	if req.Price <= 0 {
		sendErrorResponse(w, http.StatusBadRequest, "Цена должна быть больше нуля")
		return
	}

	// 6. Посчитать долги участников (до создания покупки — чтобы не создавать
	// покупку без корректных долгов при ошибке валидации разбивки)
	participants, err := h.storage.GetParticipantsByPartyID(party.ID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении участников")
		return
	}
	debts, err := buildDebts(0, req.Price, req.SplitType, participants, req.Debtors, req.DebtorAmounts)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 7. Создать покупку
	purchase := objects.Purchase{
		PartyID:        party.ID,
		BuyerID:        participant.ID,
		Name:           req.Name,
		Description:    req.Description,
		PurchaseIconID: req.PurchaseIconID,
		Price:          req.Price,
		SplitType:      req.SplitType,
		DateofPurchase: req.DateofPurchase,
	}

	purchaseID, err := h.storage.CreatePurchase(purchase)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при создании покупки")
		return
	}

	// 8. Сохранить долги
	for _, d := range debts {
		d.PurchaseID = purchaseID
		if _, err := h.storage.CreateDebt(d); err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при сохранении долгов")
			return
		}
	}

	sendSuccessResponse(w, http.StatusCreated, "Покупка создана", map[string]int{"id": purchaseID})
}

func (h *Handler) UpdatePurchaseHandler(w http.ResponseWriter, r *http.Request) {
	// Членство уже проверено RequirePartyMember
	party, participant, _ := middleware.PartyAndParticipantFromContext(r.Context())

	// 1. purchaseID из URL
	purchaseID, err := strconv.Atoi(chi.URLParam(r, "purchaseID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID покупки")
		return
	}

	// 2. Убедиться что покупка принадлежит этой тусовке
	existing, err := h.storage.GetPurchaseByID(purchaseID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Покупка не найдена")
		return
	}
	if existing.PartyID != party.ID {
		sendErrorResponse(w, http.StatusForbidden, "Покупка не принадлежит этой тусовке")
		return
	}
	// 3. Изменять может только автор покупки или админ/владелец тусовки
	if existing.BuyerID != participant.ID && !participant.IsAdmin {
		sendErrorResponse(w, http.StatusForbidden, "Изменить покупку может только автор или администратор тусовки")
		return
	}

	// 4. Декодировать тело
	var req PurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	// 5. Валидация
	if req.Name == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Название обязательно")
		return
	}
	if req.Price <= 0 {
		sendErrorResponse(w, http.StatusBadRequest, "Цена должна быть больше нуля")
		return
	}

	// 6. Пересчитать долги под новую цену/тип разбивки
	participants, err := h.storage.GetParticipantsByPartyID(party.ID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении участников")
		return
	}
	debts, err := buildDebts(purchaseID, req.Price, req.SplitType, participants, req.Debtors, req.DebtorAmounts)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 7. Обновить покупку (PartyID и BuyerID сохраняем из исходной записи)
	purchase := objects.Purchase{
		ID:             purchaseID,
		PartyID:        existing.PartyID,
		BuyerID:        existing.BuyerID,
		Name:           req.Name,
		Description:    req.Description,
		PurchaseIconID: req.PurchaseIconID,
		Price:          req.Price,
		SplitType:      req.SplitType,
		DateofPurchase: req.DateofPurchase,
	}

	if err := h.storage.UpdatePurchase(purchase); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при обновлении покупки")
		return
	}

	// 8. Заменить старые долги новыми
	if err := h.storage.DeleteDebtsByPurchaseID(purchaseID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при обновлении долгов")
		return
	}
	for _, d := range debts {
		if _, err := h.storage.CreateDebt(d); err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при сохранении долгов")
			return
		}
	}

	sendSuccessResponse(w, http.StatusOK, "Покупка обновлена", nil)
}

func (h *Handler) DeletePurchaseHandler(w http.ResponseWriter, r *http.Request) {
	// Членство уже проверено RequirePartyMember
	party, participant, _ := middleware.PartyAndParticipantFromContext(r.Context())

	// 1. purchaseID из URL
	purchaseID, err := strconv.Atoi(chi.URLParam(r, "purchaseID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID покупки")
		return
	}

	// 2. Убедиться что покупка принадлежит этой тусовке
	existing, err := h.storage.GetPurchaseByID(purchaseID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Покупка не найдена")
		return
	}
	if existing.PartyID != party.ID {
		sendErrorResponse(w, http.StatusForbidden, "Покупка не принадлежит этой тусовке")
		return
	}
	// 3. Удалить может только автор покупки или админ/владелец тусовки
	if existing.BuyerID != participant.ID && !participant.IsAdmin {
		sendErrorResponse(w, http.StatusForbidden, "Удалить покупку может только автор или администратор тусовки")
		return
	}

	// 4. Сначала удалить связанные долги (FK на purchase_id), потом саму покупку
	if err := h.storage.DeleteDebtsByPurchaseID(purchaseID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении долгов")
		return
	}
	if err := h.storage.DeletePurchase(purchaseID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении покупки")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Покупка удалена", nil)
}
