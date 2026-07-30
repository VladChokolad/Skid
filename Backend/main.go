package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/VladChokolad/Skid/Backend/internal/config"
	"github.com/VladChokolad/Skid/Backend/internal/handlers"
	"github.com/VladChokolad/Skid/Backend/internal/middleware"
	"github.com/VladChokolad/Skid/Backend/internal/storage"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	db, err := storage.Connect(cfg)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	h := handlers.NewHandler(db, cfg)

	r := chi.NewRouter()

	// Глобальные middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.CORSMiddleware(cfg))

	// Публичные маршруты
	r.HandleFunc("/echo", h.EchoHandler)
	r.Post("/auth/register", h.RegisterUserHandler)
	r.Post("/auth/login", h.LoginUserHandler)
	r.Post("/auth/join", h.CreateAnonymousHandler)

	// Защищённые маршруты
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Get("/profile", h.GetMyUserOrAnonymousHandler)
		r.Get("/profile", h.UpdateMyUserOrAnonymousHandler)
		r.Get("/profile", h.DeleteMyUserOrAnonymousHandler)

		r.Get("/parties", h.GetMyPartiesHandler)
		r.Post("/parties", h.CreatePartyHandler) //недоступно для анонимов
		r.Post("/parties", h.JoinPartyHandler)

		r.Route("/parties/{partyID}", func(r chi.Router) {
			r.Get("/", h.GetPartyHandler)
			r.Put("/", h.UpdatePartyHandler)
			r.Delete("/", h.DeletePartyHandler)

			r.Get("/participants", h.GetParticipantsHandler)
			r.Post("/participants", h.CreateParticipantHandler)
			r.Put("/participants/{participantID}/replace", h.ReplaceParticipantsHandler)
			r.Delete("/participants/{participantID}", h.DeleteParticipantHandler)

			r.Get("/purchases", h.GetPurchasesHandler)
			r.Post("/purchases", h.CreatePurchaseHandler)
			r.Put("/purchases/{purchaseID}", h.UpdatePurchaseHandler)
			r.Delete("/purchases/{purchaseID}", h.DeletePurchaseHandler)

			r.Get("/payments", h.GetPaymentsHandler)
			r.Post("/payments", h.CreatePaymentHandler)
			r.Post("/payments/{paymentID}/confirm", h.ConfirmPaymentHandler)

			r.Get("/settlements", h.GetSettlementsHandler)
		})
	})

	log.Println("Сервер запущен на порту:", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatal(err)
	}
}
