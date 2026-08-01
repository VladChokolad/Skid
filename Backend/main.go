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
	r.HandleFunc("/echo", h.EchoHandler) //Тестовый - нужно удалить
	r.Post("/auth/register", h.RegisterUserHandler)
	r.Post("/auth/login", h.LoginUserHandler)

	// Защищённые маршруты
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg, db))
		//профиль
		r.Get("/profile", h.GetMyUserOrAnonymousHandler)
		r.Put("/profile", h.UpdateMyUserOrAnonymousHandler)
		r.Delete("/profile", h.DeleteMyUserOrAnonymousHandler)
		//тусовки
		r.Get("/parties", h.GetMyPartiesHandler)
		r.Post("/parties", h.CreatePartyHandler) //недоступно для анонимов
		//вступление в тусовку
		r.Get("/invite/{inviteCode}", h.PreviewJoinHandler)
		r.Post("/invite/{inviteCode}/join", h.JoinPartyHandler)

		r.Route("/parties/{partyID}", func(r chi.Router) {
			//тусовка
			r.Get("/", h.GetPartyHandler)
			r.Put("/", h.UpdatePartyHandler)
			r.Delete("/", h.DeletePartyHandler)
			//Участники
			r.Get("/participants", h.GetParticipantsHandler)
			r.Post("/participants", h.CreateParticipantHandler)
			r.Put("/participants/{participantID}/replace", h.ReplaceParticipantsHandler)
			r.Delete("/participants/{participantID}", h.DeleteParticipantHandler)
			//траты
			r.Get("/purchases", h.GetPurchasesHandler)
			r.Post("/purchases", h.CreatePurchaseHandler)
			r.Put("/purchases/{purchaseID}", h.UpdatePurchaseHandler)
			r.Delete("/purchases/{purchaseID}", h.DeletePurchaseHandler)
			//платы
			r.Get("/payments", h.GetPaymentsHandler)
			r.Post("/payments", h.CreatePaymentHandler)
			r.Post("/payments/{paymentID}/confirm", h.ConfirmPaymentHandler)
			//Сводка
			r.Get("/settlements", h.GetSettlementsHandler)
		})
	})

	log.Println("Сервер запущен на порту:", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatal(err)
	}
}
