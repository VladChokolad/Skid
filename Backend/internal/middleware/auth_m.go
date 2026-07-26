package middleware

import (
	"context"
	"net/http"

	"github.com/VladChokolad/Skid/Backend/internal/auth"
	"github.com/VladChokolad/Skid/Backend/internal/config"
)

type contextKey string

const (
	ContextUserID      contextKey = "userID"
	ContextIsAnonymous contextKey = "isAnonymous"
)

func AuthMiddleware(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// 1. Достать токен из cookie
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// 2. Проверить токен
			claims, err := auth.ParseToken(cookie.Value, cfg.JWTSecret)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// 3. Положить данные в контекст
			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextIsAnonymous, claims.IsAnonymous)

			// 4. Обновить токен (sliding expiration)
			newToken, err := auth.CreateToken(claims.UserID, claims.IsAnonymous, cfg.JWTSecret)
			if err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     "token",
					Value:    newToken,
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
					Path:     "/",
					MaxAge:   30 * 24 * 60 * 60,
				})
			}

			// 5. Передать дальше
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(ContextUserID).(int)
	return userID, ok
}

func GetIsAnonymousFromContext(ctx context.Context) bool {
	isAnonymous, _ := ctx.Value(ContextIsAnonymous).(bool)
	return isAnonymous
}
