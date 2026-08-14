package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
	"github.com/lib/pq"
)

func generateInviteCode() string { //генерит инвайт-код
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:8]
}
func isUniqueViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}
	return pqErr.Code == "23505"
}
func (s *Storage) CreateParty(party objects.Party) (int, error) {
	// Принимает: структуру Party с данными о тусовке
	// Делает: генерирует invite_code, создаёт запись в таблице parties
	// Возвращает: id созданной тусовки или ошибку
	if party.Name == "" {
		party.Name = "Без названия"
	}

	for attempts := 0; attempts < 5; attempts++ {
		party.InviteCode = generateInviteCode()

		var id int
		err := s.db.QueryRow(
			`INSERT INTO parties (name, description, party_image, owner_id, invite_code)
			 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			party.Name, party.Description, party.PartyImage, party.OwnerID, party.InviteCode,
		).Scan(&id)

		if err == nil {
			return id, nil
		}
		if isUniqueViolation(err) {
			continue
		}
		return 0, err
	}
	return 0, fmt.Errorf("не удалось создать тусовку — попробуйте ещё раз")
}
func (s *Storage) GetPartyByID(id int) (objects.Party, error) {
	//Принимает:  Делает:  Возвращает:
	var Party objects.Party
	err := s.db.QueryRow(`
		SELECT name, description, party_image, owner_id, invite_code, is_active, created_at, updated_at 
		From parties 
		WHERE id = $1
		`, id).Scan(
		&Party.Name,
		&Party.Description,
		&Party.PartyImage,
		&Party.OwnerID,
		&Party.InviteCode,
		&Party.IsActive,
		&Party.CreatedAt,
		&Party.UpdatedAt)
	if err == sql.ErrNoRows {
		return objects.Party{}, fmt.Errorf("Вечеринка не найдена")
	}
	if err != nil {
		return objects.Party{}, err
	}
	Party.ID = id
	return Party, nil
}
func (s *Storage) GetPartyByInviteCode(code string) (objects.Party, error) {
	//Принимает:  Делает:  Возвращает:
	var Party objects.Party
	err := s.db.QueryRow(`
		SELECT id, name, description, party_image, owner_id, is_active, created_at, updated_at 
		From parties 
		WHERE invite_code = $1
		`, code).Scan(
		&Party.ID,
		&Party.Name,
		&Party.Description,
		&Party.PartyImage,
		&Party.OwnerID,
		&Party.IsActive,
		&Party.CreatedAt,
		&Party.UpdatedAt)
	if err == sql.ErrNoRows {
		return objects.Party{}, fmt.Errorf("Вечеринка не найдена")
	}
	if err != nil {
		return objects.Party{}, err
	}
	Party.InviteCode = code
	return Party, nil
}
func (s *Storage) GetPartiesByUserID(userID int) ([]objects.Party, error) {
	//Принимает:id пользователя/анонима  Делает:ищет все вечеринки ассоциированные с пользователем  Возвращает: список вечеринок
	// Шаг 1 — выполняем запрос
	// Query возвращает много строк — в отличие от QueryRow
	rows, err := s.db.Query(`
		SELECT parties.id, parties.name, parties.description,
	              parties.party_image, parties.owner_id, parties.invite_code,
	              parties.is_active, parties.created_at, parties.updated_at
	       FROM parties
	       INNER JOIN participants ON parties.id = participants.party_id
	       WHERE participants.user_or_anonymous_id = $1
		AND participants.is_anonymous = FALSE
		`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // закрыть курсор когда функция завершится

	// Шаг 2 — создаём пустой слайс куда будем складывать тусовки
	var parties []objects.Party

	// Шаг 3 — перебираем строки одну за другой
	// rows.Next() возвращает true пока есть строки
	// когда строки закончились — возвращает false и цикл останавливается
	for rows.Next() {
		// создаём пустую структуру для текущей строки
		var Party objects.Party

		// читаем все колонки текущей строки в поля структуры
		// порядок полей в Scan должен совпадать с порядком в SELECT
		err := rows.Scan(
			&Party.ID,
			&Party.Name,
			&Party.Description,
			&Party.PartyImage,
			&Party.OwnerID,
			&Party.InviteCode,
			&Party.IsActive,
			&Party.CreatedAt,
			&Party.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// добавляем заполненную структуру в слайс
		parties = append(parties, Party)
	}

	// Шаг 4 — проверяем не было ли ошибки во время перебора
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Шаг 5 — возвращаем слайс тусовок
	// если тусовок нет — вернётся пустой слайс [], nil
	return parties, nil
}
func (s *Storage) GetPartiesByAnonID(anonID int) ([]objects.Party, error) {
	rows, err := s.db.Query(`
		SELECT parties.id, parties.name, parties.description,
	              parties.party_image, parties.owner_id, parties.invite_code,
	              parties.is_active, parties.created_at, parties.updated_at
	       FROM parties
	       INNER JOIN participants ON parties.id = participants.party_id
	       WHERE participants.user_or_anonymous_id = $1
	       AND participants.is_anonymous = TRUE
		`, anonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parties []objects.Party
	for rows.Next() {
		var party objects.Party
		err := rows.Scan(
			&party.ID,
			&party.Name,
			&party.Description,
			&party.PartyImage,
			&party.OwnerID,
			&party.InviteCode,
			&party.IsActive,
			&party.CreatedAt,
			&party.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return parties, nil
}
func (s *Storage) UpdateParty(party objects.Party) error {
	//Принимает:Информацию о вечеринке  Делает:Обновляет бд  Возвращает: только ошибку
	result, err := s.db.Exec(`
		UPDATE parties 
		SET name = $1, description = $2, party_image = $3, owner_id = $4, is_active = $5 
		WHERE id = $6
		`,
		party.Name, party.Description, party.PartyImage, party.OwnerID, party.IsActive, party.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("Вечеринка не найдена")
	}
	return nil
}
func (s *Storage) DeleteParty(id int) error {
	//Принимает:  Делает:  Возвращает:
	result, err := s.db.Exec(`
	DELETE FROM parties 
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
		return fmt.Errorf("Вечеринка не найдена")
	}
	return nil
}
