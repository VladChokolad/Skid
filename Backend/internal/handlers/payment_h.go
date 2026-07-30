package handlers

import (
	"net/http"
)

func (h *Handler) GetPaymentsHandler(w http.ResponseWriter, r *http.Request)    {}
func (h *Handler) CreatePaymentHandler(w http.ResponseWriter, r *http.Request)  {}
func (h *Handler) ConfirmPaymentHandler(w http.ResponseWriter, r *http.Request) {}
