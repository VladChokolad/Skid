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
	Name           string     `json:"name"`
	Description    *string    `json:"description"`
	PurchaseIconID *int       `json:"purchaseIconId"`
	Price          float64    `json:"price"`
	SplitType      int        `json:"splitType"`
	DateofPurchase *time.Time `json:"dateOfPurchase"`
	Debtors        []int      `json:"debtors"`
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

	// 6. Создать покупку
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

	// 6. Обновить покупку (PartyID и BuyerID сохраняем из исходной записи)
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

	// 4. Удалить покупку
	if err := h.storage.DeletePurchase(purchaseID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении покупки")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Покупка удалена", nil)
}
