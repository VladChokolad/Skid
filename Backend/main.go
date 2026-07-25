package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/VladChokolad/Skid/Backend/internal/config"
	"github.com/VladChokolad/Skid/Backend/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()      // загружает .env в окружение
	cfg := config.Load() // читает окружение в структуру

	// теперь можешь использовать cfg везде
	fmt.Println(cfg.DBHost)     // УДАЛИТЬ В ФИНАЛЬНОЙ ВЕРСИИ
	fmt.Println(cfg.ServerPort) // УДАЛИТЬ В ФИНАЛЬНОЙ ВЕРСИИ
	fmt.Println(cfg.JWTSecret)  // УДАЛИТЬ В ФИНАЛЬНОЙ ВЕРСИИ
	fmt.Println(cfg.DBPort)     // УДАЛИТЬ В ФИНАЛЬНОЙ ВЕРСИИ
	fmt.Println(cfg.DBUser)     // УДАЛИТЬ В ФИНАЛЬНОЙ ВЕРСИИ
	fmt.Println(cfg.DBPassword) // УДАЛИТЬ В ФИНАЛЬНОЙ ВЕРСИИ
	fmt.Println(cfg.DBName)     // УДАЛИТЬ В ФИНАЛЬНОЙ ВЕРСИИ
	db, err := storage.Connect(cfg)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	fmt.Println("Подключились к базе данных!")
	fmt.Println("Сервер запустится на порту:", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		log.Fatal(err)
	}
}
