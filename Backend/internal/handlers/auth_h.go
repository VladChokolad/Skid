package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
	"github.com/golang-jwt/jwt/v5"
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

// Для входа анонима по ссылке
type JoinRequest struct {
	InviteCode string `json:"inviteCode"` // обязательно
	Name       string `json:"name"`       // обязательно — минимум для отображения
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
	token, err := h.generateUserJWT(user.ID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при генерации токена")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Успешный вход", map[string]interface{}{
		"user":  user,
		"token": token,
	})

}

func (h *Handler) CreateAnonymousHandler(w http.ResponseWriter, r *http.Request) {
	var req JoinRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.InviteCode == "" || req.Name == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Invite-код и имя обязательны")
		return
	}

	// Ищем вечеринку по invite-коду
	party, err := h.storage.GetPartyByInviteCode(req.InviteCode)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "Вечеринка с таким invite-кодом не найдена")
		return
	}

	// Проверяем, не закрыта ли вечеринка
	if party.IsActive == false {
		sendErrorResponse(w, http.StatusBadRequest, "Вечеринка закрыта, присоединение невозможно")
		return
	}

	// Создаем анонимного пользователя
	anonymoususer := objects.AnonymousUser{
		Name: req.Name,
	}

	anonymoususerID, err := h.storage.CreateAnonymousUser(anonymoususer)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при присоединении к вечеринке")
		return
	}

	// Генерируем JWT токен для анонима
	token, err := h.generateAnonJWT(anonymoususerID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Ошибка при генерации токена")
		return
	}

	sendSuccessResponse(w, http.StatusOK, "Успешное присоединение к вечеринке", map[string]interface{}{
		"anonymoususer": anonymoususerID,
		"token":         token,
	})
}

func (h *Handler) generateUserJWT(ID int) (string, error) { // generateJWT генерирует JWT токен
	claims := jwt.MapClaims{
		"id":     ID,
		"isanon": false,
		"exp":    time.Now().Add(time.Hour * 12 * 1).Unix(), // токен действителен 12 часов
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

func (h *Handler) generateAnonJWT(ID int) (string, error) {
	claims := jwt.MapClaims{
		"id":     ID,
		"isanon": true,
		"exp":    time.Now().Add(time.Hour * 12 * 1).Unix(), // токен действителен 12 часов
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
