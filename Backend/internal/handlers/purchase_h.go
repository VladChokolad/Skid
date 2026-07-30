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
	// 1. Получить partyID из URL
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	// 2. Получить покупки из БД
	purchases, err := h.storage.GetPurchasesByPartyID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении покупок")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Список покупок", purchases)
}

func (h *Handler) CreatePurchaseHandler(w http.ResponseWriter, r *http.Request) {
	// 1. partyID из URL
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	// 2. userID из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	// 3. Найти participant
	isAnon := middleware.GetIsAnonymousFromContext(r.Context())
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

	// 6. Создать покупку
	purchase := objects.Purchase{
		PartyID:        partyID,
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
	// 1. purchaseID из URL
	purchaseID, err := strconv.Atoi(chi.URLParam(r, "purchaseID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID покупки")
		return
	}

	// 2. Декодировать тело
	var req PurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	// 3. Валидация
	if req.Name == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Название обязательно")
		return
	}
	if req.Price <= 0 {
		sendErrorResponse(w, http.StatusBadRequest, "Цена должна быть больше нуля")
		return
	}

	// 4. Обновить покупку
	purchase := objects.Purchase{
		ID:             purchaseID,
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
	// 1. purchaseID из URL
	purchaseID, err := strconv.Atoi(chi.URLParam(r, "purchaseID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID покупки")
		return
	}

	// 2. Удалить покупку
	if err := h.storage.DeletePurchase(purchaseID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении покупки")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Покупка удалена", nil)
}
