package storage

import (
	"database/sql"
	"fmt"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

// Принимает: структуру AnonymousUser с именем и телефоном
// Делает: создаёт запись анонимного пользователя в таблице anonymous_users
// Возвращает: id созданной записи или ошибку
func (s *Storage) CreateAnonymousUser(anonymousUser objects.AnonymousUser) (int, error) {
	var id int
	err := s.db.QueryRow(
		"INSERT INTO anonymous_users (name, phone) VALUES ($1, $2) RETURNING id",
		anonymousUser.Name, anonymousUser.Phone,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Принимает: id анонимного пользователя
// Делает: ищет запись в таблице anonymous_users по id
// Возвращает: заполненную структуру AnonymousUser или ошибку если не найден
func (s *Storage) GetAnonymousByID(id int) (objects.AnonymousUser, error) {
	var anonymousUser objects.AnonymousUser
	err := s.db.QueryRow(
		"SELECT name, phone, created_at FROM anonymous_users WHERE id = $1",
		id,
	).Scan(
		&anonymousUser.Name,
		&anonymousUser.Phone,
		&anonymousUser.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return objects.AnonymousUser{}, fmt.Errorf("пользователь не найден")
	}
	if err != nil {
		return objects.AnonymousUser{}, err
	}
	anonymousUser.ID = id
	return anonymousUser, nil
}
