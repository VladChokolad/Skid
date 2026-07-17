package main

import (
	"fmt"

	"github.com/VladChokolad/Skid/Backend/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()      // загружает .env в окружение
	cfg := config.Load() // читает окружение в структуру

	// теперь можешь использовать cfg везде
	fmt.Println(cfg.DBHost)     // localhost
	fmt.Println(cfg.ServerPort) // 8080
	fmt.Println(cfg.JWTSecret)  // 69696969696969
	fmt.Println(cfg.DBPort)     // 5432
	fmt.Println(cfg.DBUser)     // postgres
	fmt.Println(cfg.DBPassword) // 312Mikro50022Pip
	fmt.Println(cfg.DBName)     // skid
}
