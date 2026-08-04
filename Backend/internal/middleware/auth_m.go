package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"github.com/VladChokolad/Skid/Backend/internal/activity"
	"github.com/VladChokolad/Skid/Backend/internal/auth"
	"github.com/VladChokolad/Skid/Backend/internal/config"
	"github.com/VladChokolad/Skid/Backend/internal/storage"
)

type contextKey string

const (
	ContextUserOrAnonymousID contextKey = "userOrAnonymousID"
	ContextIsAnonymous       contextKey = "isAnonymous"
)

// GuestOrAuthMiddleware обрабатывает как зарегистрированных, так и анонимных пользователей.
// Если токен отсутствует или невалиден – создаёт нового анонима, выдаёт токен и пропускает запрос.
// Если токен есть и валиден – извлекает userID и isAnonymous из токена.
// Во всех случаях кладёт в контекст userID и isAnonymous.
func AuthMiddleware(cfg config.Config, s *storage.Storage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// 1. Пытаемся получить токен из cookie
			cookie, err := r.Cookie("token")
			if err != nil { // Мы не получили cookie value
				// создаём анонима
				anonID, err := s.CreateAnonymousUser()
				if err != nil {
					http.Error(w, "Failed to create guest", http.StatusInternalServerError)
					return
				}

				// Генерируем токен для нового анонима
				token, err := auth.CreateToken(anonID, true, 30, cfg.JWTSecret)
				if err != nil {
					http.Error(w, "Failed to generate token", http.StatusInternalServerError)
					return
				}
				SetGuestCookie(w, token)
				// Кладём данные в контекст и передаём управление дальше (аноним создан)
				ctx := context.WithValue(r.Context(), ContextUserOrAnonymousID, anonID)
				ctx = context.WithValue(ctx, ContextIsAnonymous, true)
				next.ServeHTTP(w, r.WithContext(ctx))
				return

			} //Мы получили cookie.Value
			// Парсим токен
			claims, err := auth.ParseToken(cookie.Value, cfg.JWTSecret)

			if errors.Is(err, jwt.ErrTokenExpired) {
				// Токен просрочен – возвращаем 401
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
			// Другие ошибки (невалидный токен) – тоже 401
			if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Токен валиден – обновляем sliding expiration
			if claims.IsAnonymous {
				newToken, err := auth.CreateToken(claims.UserOrAnonymousID, true, 30, cfg.JWTSecret)
				if err == nil {
					SetGuestCookie(w, newToken)
				}
			} else {
				newToken, err := auth.CreateToken(claims.UserOrAnonymousID, false, 14, cfg.JWTSecret)
				if err == nil {
					SetGuestCookie(w, newToken)
				}
			}
			// 3. Обновление last_activity для анонимов (с троттлингом)
			if claims.IsAnonymous && activity.ShouldUpdate(claims.UserOrAnonymousID) {
				go func(anonID int) {
					_ = s.UpdateAnonymousActivity(anonID)
				}(claims.UserOrAnonymousID)
			}

			// 4. Кладём данные в контекст
			ctx := context.WithValue(r.Context(), ContextUserOrAnonymousID, claims.UserOrAnonymousID)
			ctx = context.WithValue(ctx, ContextIsAnonymous, claims.IsAnonymous)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// setGuestCookie устанавливает cookie с токеном (единая функция для всех случаев).
func SetGuestCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 дней
	})
}

func GetUserOrAnonymousIDFromContext(ctx context.Context) (int, bool) {
	userOrAnonymousID, ok := ctx.Value(ContextUserOrAnonymousID).(int)
	return userOrAnonymousID, ok
}

func GetIsAnonymousFromContext(ctx context.Context) bool {
	isAnonymous, _ := ctx.Value(ContextIsAnonymous).(bool)
	return isAnonymous
}
