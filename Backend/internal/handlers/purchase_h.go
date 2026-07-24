package handlers

import (
	"net/http"
	"time"
)

// Для создания/изменения тусовки
type PurchaseRequest struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	PartyID        int    // Вечеринка в которой сделана покупка
	BuyerID        int    // Participant который заплатил
	Name           string
	Description    *string
	PurchaseIconID *int       // Иконка покупки — для визуального отображения в списке
	Price          float64    // Полная сумма покупки в рублях
	SplitType      int        // 0 - поровну всем, 1 - поровну выбранным, 2 - индивидуальные суммы, 3 - индивидуальные доли
	DateofPurchase *time.Time // Когда совершена покупка — может отличаться от CreatedAt
	CreatedAt      time.Time
}

func (h *Handler) CreatePurchaseHandler(w http.ResponseWriter, r *http.Request) {}
func (h *Handler) UpdatePurchaseHandler(w http.ResponseWriter, r *http.Request) {}
func (h *Handler) DeletePurchaseHandler(w http.ResponseWriter, r *http.Request) {}
