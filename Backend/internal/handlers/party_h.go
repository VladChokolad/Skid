package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/VladChokolad/Skid/Backend/internal/middleware"
	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

type PartyRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	PartyImage  *string `json:"partyImage"`
}

type JoinRequest struct {
	InviteCode    string
	PlaceHolderID *int
}

func (h *Handler) GetMyPartiesHandler(w http.ResponseWriter, r *http.Request) { //Показывает все вечеринки пользователя
	// 1. Получаем userID и флаг анонимности
	ID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	isAnonymous := middleware.GetIsAnonymousFromContext(r.Context())

	// 2. Проверяем, авторизован ли пользователь (если требуется)
	if !ok || ID == 0 {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var parties_list []objects.Party // или конкретный тип, возвращаемый хранилищем
	var err error

	if isAnonymous {
		parties_list, err = h.storage.GetPartiesByAnonID(ID)
	} else {
		parties_list, err = h.storage.GetPartiesByUserID(ID)
	}
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Не удалось получить данные")
		return
	}
	sendSuccessResponse(w, http.StatusCreated, "Вот ваши тусовки!", map[string]interface{}{
		"data": parties_list,
	})
}

func (h *Handler) GetPartyHandler(w http.ResponseWriter, r *http.Request) {
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	party, err := h.storage.GetPartyByID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Тусовка не найдена")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Тусовка", party)
}

func (h *Handler) PreviewJoinHandler(w http.ResponseWriter, r *http.Request) {
	inviteCode := chi.URLParam(r, "inviteCode")
	if inviteCode == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Invite-код обязателен")
		return
	}

	// Найти тусовку
	party, err := h.storage.GetPartyByInviteCode(inviteCode)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Тусовка не найдена")
		return
	}
	if !party.IsActive {
		sendErrorResponse(w, http.StatusBadRequest, "Тусовка закрыта")
		return
	}

	// Получить всех участников
	participants, err := h.storage.GetParticipantsByPartyID(party.ID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении участников")
		return
	}

	// Отфильтровать только placeholders
	var placeholders []map[string]interface{}
	for _, p := range participants {
		if p.IsPlaceholder {
			placeholders = append(placeholders, map[string]interface{}{
				"id":   p.ID,
				"name": p.Name,
			})
		}
	}

	// Если пустой — вернуть пустой слайс а не nil
	if placeholders == nil {
		placeholders = []map[string]interface{}{}
	}

	sendSuccessResponse(w, http.StatusOK, "Информация о тусовке", map[string]interface{}{
		"partyID":      party.ID,
		"partyName":    party.Name,
		"placeholders": placeholders,
	})
}

func (h *Handler) JoinPartyHandler(w http.ResponseWriter, r *http.Request) {
	inviteCode := chi.URLParam(r, "inviteCode")
	if inviteCode == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Invite-код обязателен")
		return
	}

	// Получить userID из контекста
	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	isAnon := middleware.GetIsAnonymousFromContext(r.Context())

	// Найти тусовку
	party, err := h.storage.GetPartyByInviteCode(inviteCode)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Тусовка не найдена")
		return
	}
	if !party.IsActive {
		sendErrorResponse(w, http.StatusBadRequest, "Тусовка закрыта")
		return
	}

	// Проверить что пользователь ещё не в тусовке
	if isAnon {
		_, err = h.storage.GetParticipantByAnonAndPartyID(userID, party.ID)
	} else {
		_, err = h.storage.GetParticipantByUserAndPartyID(userID, party.ID)
	}
	if err == nil {
		sendErrorResponse(w, http.StatusConflict, "Вы уже участник этой тусовки")
		return
	}

	// Получить имя пользователя
	var name string
	if isAnon {
		anon, err := h.storage.GetAnonymousByID(userID)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении данных")
			return
		}
		name = anon.Name
	} else {
		user, err := h.storage.GetUserByID(userID)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при получении данных")
			return
		}
		name = user.Name
	}

	// Декодировать тело — placeholderID опционален
	var req struct {
		PlaceholderID *int `json:"placeholderID"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Если выбрал placeholder — занять его место
	if req.PlaceholderID != nil {
		placeholder, err := h.storage.GetParticipantByID(*req.PlaceholderID)
		if err != nil {
			sendErrorResponse(w, http.StatusNotFound, "Placeholder не найден")
			return
		}
		if !placeholder.IsPlaceholder {
			sendErrorResponse(w, http.StatusBadRequest, "Этот участник не является placeholder")
			return
		}
		if placeholder.PartyID != party.ID {
			sendErrorResponse(w, http.StatusForbidden, "Placeholder не принадлежит этой тусовке")
			return
		}

		placeholder.UserOrAnonymousID = &userID
		placeholder.IsAnonymous = isAnon
		placeholder.IsPlaceholder = false
		placeholder.Name = name

		if err := h.storage.UpdateParticipant(placeholder); err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при занятии места")
			return
		}

		sendSuccessResponse(w, http.StatusOK, "Вы заняли место участника", map[string]interface{}{
			"partyID":       party.ID,
			"participantID": placeholder.ID,
		})
		return
	}

	// Иначе создать нового участника
	participant := objects.Participant{
		PartyID:           party.ID,
		UserOrAnonymousID: &userID,
		Name:              name,
		IsAnonymous:       isAnon,
	}

	participantID, err := h.storage.CreateParticipant(participant)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при добавлении участника")
		return
	}

	sendSuccessResponse(w, http.StatusCreated, "Вы присоединились к тусовке", map[string]interface{}{
		"partyID":       party.ID,
		"participantID": participantID,
	})
}

func (h *Handler) CreatePartyHandler(w http.ResponseWriter, r *http.Request) { //Создаёт вечеринку
	var req PartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	// Только зарегистрированные могут создавать тусовки
	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok || userID == 0 {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	if middleware.GetIsAnonymousFromContext(r.Context()) {
		sendErrorResponse(w, http.StatusForbidden, "Анонимы не могут создавать тусовки")
		return
	}

	// Создаём тусовку
	party := objects.Party{
		Name:    *req.Name,
		OwnerID: userID,
	}
	if req.Description != nil {
		party.Description = req.Description
	}
	if req.PartyImage != nil {
		party.PartyImage = req.PartyImage
	}

	_, err := h.storage.CreateParty(party)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при создании тусовки")
		return
	}
	sendSuccessResponse(w, http.StatusCreated, "тусовка создана!", nil)

}

func (h *Handler) UpdatePartyHandler(w http.ResponseWriter, r *http.Request) {
	// partyID из URL
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	// Только зарегистрированные
	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	// Проверить что пользователь владелец
	party, err := h.storage.GetPartyByID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Тусовка не найдена")
		return
	}
	if party.OwnerID != userID {
		sendErrorResponse(w, http.StatusForbidden, "Только владелец может изменять тусовку")
		return
	}

	// Декодировать тело
	var req PartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	// Обновить
	party.Name = *req.Name
	if req.Description != nil {
		party.Description = req.Description
	}
	if req.PartyImage != nil {
		party.PartyImage = req.PartyImage
	}

	if err := h.storage.UpdateParty(party); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при обновлении тусовки")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Тусовка обновлена", nil)
}

func (h *Handler) DeletePartyHandler(w http.ResponseWriter, r *http.Request) {
	// partyID из URL
	partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный ID тусовки")
		return
	}

	// Только зарегистрированные
	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	// Проверить что пользователь владелец
	party, err := h.storage.GetPartyByID(partyID)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Тусовка не найдена")
		return
	}
	if party.OwnerID != userID {
		sendErrorResponse(w, http.StatusForbidden, "Только владелец может удалять тусовку")
		return
	}

	// Удалить
	if err := h.storage.DeleteParty(partyID); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении тусовки")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Тусовка удалена", nil)
}
