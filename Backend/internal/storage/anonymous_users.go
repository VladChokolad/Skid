package storage

import (
	"database/sql"
	"fmt"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

// Принимает: структуру AnonymousUser с именем и телефоном
// Делает: создаёт запись анонимного пользователя в таблице anonymous_users
// Возвращает: id созданной записи или ошибку
func (s *Storage) CreateAnonymousUser() (int, error) {
	var id int
	err := s.db.QueryRow(`
		INSERT INTO anonymous_users (name, last_activity) 
		VALUES ('Аноним', NOW()) 
		RETURNING id
		`).Scan(&id)
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
	err := s.db.QueryRow(`
		SELECT name, phone, created_at 
		FROM anonymous_users 
		WHERE id = $1
		`, id).Scan(
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

func (s *Storage) UpdateAnonymousActivity(id int) error {
	_, err := s.db.Exec(
		"UPDATE anonymous_users SET last_activity = NOW() WHERE id = $1",
		id,
	)
	return err
}

func (s *Storage) UpdateAnonymous(anonymousUser objects.AnonymousUser) error {
	//Принимает: Cтруктуру anonymousUser. Делает: обнавляет информацию в бд  Возвращает: только ошибку
	result, err := s.db.Exec(`
		UPDATE anonymous_users 
		SET name = $1, phone = $2
		WHERE id = $3
		`, anonymousUser.Name, anonymousUser.Phone, anonymousUser.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("аноним не найден")
	}
	return nil

}
func (s *Storage) DeleteAnonymous(id int) error {
	//Принимает: id анонима  Делает: удаляет анонима  Возвращает: только ошибку
	result, err := s.db.Exec(`
		DELETE FROM anonymous_users 
		WHERE id = $1
		`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("аноним не найден")
	}
	return nil
}
