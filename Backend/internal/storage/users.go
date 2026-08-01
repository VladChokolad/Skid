package storage

import (
	"database/sql"
	"fmt"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

func (s *Storage) CreateUser(User objects.User) (int, error) {
	//Принимает: Данные о пользователе.  Делает: создаёт строку пользователя в бд.  Возвращает: id только что созданного user.
	var id int
	err := s.db.QueryRow(`
		INSERT INTO users (name, email, password_hash) 
		VALUES ($1, $2, $3) 
		RETURNING id
		`,
		User.Name, User.Email, User.PasswordHash).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}
func (s *Storage) GetUserByID(id int) (objects.User, error) {
	//Принимает: Id пользователя  Делает: Читает из бд  Возвращает: всю информацию о пользователе.
	var User objects.User
	err := s.db.QueryRow(`
		SELECT name, email, password_hash, phone, profile_image,created_at 
		FROM users 
		WHERE id = $1
		`, id).Scan(
		&User.Name,
		&User.Email,
		&User.PasswordHash,
		&User.Phone,
		&User.ProfileImage,
		&User.CreatedAt)
	if err == sql.ErrNoRows {
		return objects.User{}, fmt.Errorf("пользователь не найден")
	}
	if err != nil {
		return objects.User{}, err
	}
	User.ID = id
	return User, nil

}
func (s *Storage) GetUserByEmail(email string) (objects.User, error) {
	//Принимает: email пользователя Делает: Читает из бд.  Возвращает: всю информацию о пользователе.
	var User objects.User
	err := s.db.QueryRow(`
		SELECT id, name, password_hash, phone, profile_image,created_at 
		FROM users 
		WHERE email = $1
		`, email).Scan(
		&User.ID,
		&User.Name,
		&User.PasswordHash,
		&User.Phone,
		&User.ProfileImage,
		&User.CreatedAt)
	if err == sql.ErrNoRows {
		return objects.User{}, fmt.Errorf("пользователь не найден")
	}
	if err != nil {
		return objects.User{}, err
	}
	User.Email = email
	return User, nil
}
func (s *Storage) UpdateUser(user objects.User) error {
	//Принимает: Cтруктуру User. Делает: обнавляет информацию в бд  Возвращает: только ошибку
	result, err := s.db.Exec(`
		UPDATE users 
		SET name = $1, phone = $2, profile_image = $3 
		WHERE id = $4
		`,
		user.Name, user.Phone, user.ProfileImage, user.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("пользователь не найден")
	}
	return nil

}
func (s *Storage) DeleteUser(id int) error {
	//Принимает: id пользователя  Делает: удаляет пользователя  Возвращает: только ошибку
	result, err := s.db.Exec(`
	DELETE FROM users 
	WHERE id = $1
	`, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("пользователь не найден")
	}
	return nil
}
