package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
	"github.com/VladChokolad/Skid/Backend/internal/storage"
)

const (
	ContextParty       contextKey = "party"
	ContextParticipant contextKey = "participant"
)

// RequirePartyMember проверяет, что текущий пользователь (или аноним) состоит
// в тусовке, чей ID взят из URL. Кладёт в контекст саму тусовку и участника,
// чтобы хендлеры и следующие мидлвейры не делали повторные запросы в БД.
// Должен идти в цепочке после AuthMiddleware.
func RequirePartyMember(s *storage.Storage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			partyID, err := strconv.Atoi(chi.URLParam(r, "partyID"))
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "Неверный ID тусовки")
				return
			}

			party, err := s.GetPartyByID(partyID)
			if err != nil {
				writeJSONError(w, http.StatusNotFound, "Тусовка не найдена")
				return
			}

			userID, ok := GetUserOrAnonymousIDFromContext(r.Context())
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "Не авторизован")
				return
			}

			var participant objects.Participant
			if GetIsAnonymousFromContext(r.Context()) {
				participant, err = s.GetParticipantByAnonAndPartyID(userID, partyID)
			} else {
				participant, err = s.GetParticipantByUserAndPartyID(userID, partyID)
			}
			if err != nil {
				writeJSONError(w, http.StatusForbidden, "Вы не участник этой тусовки")
				return
			}

			ctx := context.WithValue(r.Context(), ContextParty, party)
			ctx = context.WithValue(ctx, ContextParticipant, participant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePartyAdmin пропускает только владельца тусовки или участника с IsAdmin.
// Должен идти в цепочке после RequirePartyMember.
func RequirePartyAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			party, participant, ok := PartyAndParticipantFromContext(r.Context())
			userID, _ := GetUserOrAnonymousIDFromContext(r.Context())
			if !ok || (party.OwnerID != userID && !participant.IsAdmin) {
				writeJSONError(w, http.StatusForbidden, "Недостаточно прав")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePartyOwner пропускает только владельца тусовки.
// Должен идти в цепочке после RequirePartyMember.
func RequirePartyOwner() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			party, ok := PartyFromContext(r.Context())
			userID, _ := GetUserOrAnonymousIDFromContext(r.Context())
			if !ok || party.OwnerID != userID {
				writeJSONError(w, http.StatusForbidden, "Только владелец может выполнить это действие")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func PartyFromContext(ctx context.Context) (objects.Party, bool) {
	party, ok := ctx.Value(ContextParty).(objects.Party)
	return party, ok
}

func ParticipantFromContext(ctx context.Context) (objects.Participant, bool) {
	participant, ok := ctx.Value(ContextParticipant).(objects.Participant)
	return participant, ok
}

func PartyAndParticipantFromContext(ctx context.Context) (objects.Party, objects.Participant, bool) {
	party, ok1 := PartyFromContext(ctx)
	participant, ok2 := ParticipantFromContext(ctx)
	return party, participant, ok1 && ok2
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
