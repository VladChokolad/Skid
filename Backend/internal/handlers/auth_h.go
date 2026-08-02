package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/VladChokolad/Skid/Backend/internal/auth"
	"github.com/VladChokolad/Skid/Backend/internal/middleware"
	"github.com/VladChokolad/Skid/Backend/internal/objects"
	"golang.org/x/crypto/bcrypt"
)

// Для регистрации
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Для логина
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type AnonymousCreationRequest struct {
	Name string `json:"name"` // обязательно — минимум для отображения
}

func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	//докодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	//валидация пустых полей - нужно удалить
	if req.Name == "" || req.Email == "" || req.Password == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Все поля обязательны для заполнения")
		return
	}
	//Валидация длинны пароля - нужно удалить
	if len(req.Password) < 8 {
		sendErrorResponse(w, http.StatusBadRequest, "Пароль должен содержать минимум 8 символов")
		return
	}
	//Валидация оригинальности пароля
	_, err := h.storage.GetUserByEmail(req.Email)
	if err == nil {
		sendErrorResponse(w, http.StatusConflict, "Пользователь с таким email уже существует")
		return
	}
	//Хеширование пароля
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при хешировании пароля")
		return
	}
	//Создаём пользователя
	user := objects.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
	}
	//Отправляем пользователя в бд
	_, err = h.storage.CreateUser(user)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при создании пользователя")
		return
	}

	sendSuccessResponse(w, http.StatusCreated, "Пользователь успешно зарегистрирован", nil)
}

func (h *Handler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.Email == "" || req.Password == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Email и пароль обязательны")
		return
	}

	// Ищем пользователя по email
	user, err := h.storage.GetUserByEmail(req.Email)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}

	// Генерируем JWT токен
	token, err := auth.CreateToken(user.ID, false, 14, h.cfg.JWTSecret)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при генерации токена")
		return
	}
	middleware.SetGuestCookie(w, token)
	sendSuccessResponse(w, http.StatusOK, "Успешный вход", nil)
}
