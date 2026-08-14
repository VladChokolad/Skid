package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/VladChokolad/Skid/Backend/internal/middleware"
	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

type UpdateUserRequest struct {
	Name  string  `json:"name"`
	Phone *string `json:"phone"`
}

func (h *Handler) GetMyUserOrAnonymousHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	isAnon := middleware.GetIsAnonymousFromContext(r.Context())

	if isAnon {
		user, err := h.storage.GetAnonymousByID(userID)
		if err != nil {
			sendErrorResponse(w, http.StatusNotFound, "Пользователь не найден")
			return
		}
		sendSuccessResponse(w, http.StatusOK, "Данные пользователя", user)
	} else {
		user, err := h.storage.GetUserByID(userID)
		if err != nil {
			sendErrorResponse(w, http.StatusNotFound, "Пользователь не найден")
			return
		}
		// Не отправляем хэш пароля
		sendSuccessResponse(w, http.StatusOK, "Данные пользователя", map[string]interface{}{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"phone":     user.Phone,
			"createdAt": user.CreatedAt,
		})
	}
}

func (h *Handler) UpdateMyUserOrAnonymousHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	isAnon := middleware.GetIsAnonymousFromContext(r.Context())

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	if req.Name == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Имя обязательно")
		return
	}

	if isAnon {
		anon := objects.AnonymousUser{
			ID:    userID,
			Name:  req.Name,
			Phone: req.Phone,
		}
		if err := h.storage.UpdateAnonymous(anon); err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при обновлении")
			return
		}
	} else {
		user := objects.User{
			ID:    userID,
			Name:  req.Name,
			Phone: req.Phone,
		}
		if err := h.storage.UpdateUser(user); err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при обновлении")
			return
		}
	}

	sendSuccessResponse(w, http.StatusOK, "Данные обновлены", nil)
}

func (h *Handler) DeleteMyUserOrAnonymousHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserOrAnonymousIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	isAnon := middleware.GetIsAnonymousFromContext(r.Context())

	if isAnon {
		if err := h.storage.DeleteAnonymous(userID); err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении")
			return
		}
	} else {
		if err := h.storage.DeleteUser(userID); err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при удалении")
			return
		}
	}

	// Удалить cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	sendSuccessResponse(w, http.StatusOK, "Аккаунт удалён", nil)
}

func (h *Handler) LogoutUserHandler(w http.ResponseWriter, r *http.Request) {
	// Удалить cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	sendSuccessResponse(w, http.StatusOK, "Выход выполнен", nil)
}
