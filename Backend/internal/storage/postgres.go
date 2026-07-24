package storage

import (
	"database/sql"
	"fmt"

	"github.com/VladChokolad/Skid/Backend/internal/config"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func Connect(cfg config.Config) (*Storage, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия БД: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("БД недоступна: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}
