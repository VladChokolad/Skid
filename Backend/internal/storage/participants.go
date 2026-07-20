package storage

import (
	"database/sql"
	"fmt"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

func (s *Storage) CreateParticipant(participant objects.Participant) (int, error) {
	// Принимает: структуру Participant
	// Делает: создаёт запись в таблице participants
	// Возвращает: id созданного участника или ошибку
	var id int
	err := s.db.QueryRow(
		`INSERT INTO participants (party_id, user_or_anonymous_id, name, is_admin, is_anonymous, is_placeholder)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		participant.PartyID,
		participant.UserOrAnonymousID,
		participant.Name,
		participant.IsAdmin,
		participant.IsAnonymous,
		participant.IsPlaceholder,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func (s *Storage) GetParticipantByID(id int) (objects.Participant, error) {
	// Принимает: id участника
	// Делает: ищет участника по id
	// Возвращает: структуру Participant или ошибку
	var participant objects.Participant
	err := s.db.QueryRow(
		`SELECT party_id, user_or_anonymous_id, name, is_admin, is_anonymous, is_placeholder, created_at
		 FROM participants
		 WHERE id = $1`, id,
	).Scan(
		&participant.PartyID,
		&participant.UserOrAnonymousID,
		&participant.Name,
		&participant.IsAdmin,
		&participant.IsAnonymous,
		&participant.IsPlaceholder,
		&participant.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return objects.Participant{}, fmt.Errorf("участник не найден")
	}
	if err != nil {
		return objects.Participant{}, err
	}
	participant.ID = id
	return participant, nil
}
func (s *Storage) GetParticipantsByPartyID(partyID int) ([]objects.Participant, error) {
	// Принимает: id вечеринки
	// Делает: ищет всех участников вечеринки
	// Возвращает: список участников или ошибку
	rows, err := s.db.Query(
		`SELECT id, user_or_anonymous_id, name, is_admin, is_anonymous, is_placeholder, created_at
		 FROM participants
		 WHERE party_id = $1`, partyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []objects.Participant
	for rows.Next() {
		var participant objects.Participant
		err := rows.Scan(
			&participant.ID,
			&participant.UserOrAnonymousID,
			&participant.Name,
			&participant.IsAdmin,
			&participant.IsAnonymous,
			&participant.IsPlaceholder,
			&participant.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		participant.PartyID = partyID
		participants = append(participants, participant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return participants, nil
}
func (s *Storage) GetParticipantByUserID(userID, partyID int) (objects.Participant, error) {
	// Принимает: id пользователя и id вечеринки
	// Делает: ищет участника-пользователя (не анонима) в конкретной вечеринке
	// Возвращает: структуру Participant или ошибку
	var participant objects.Participant
	err := s.db.QueryRow(
		`SELECT id, name, is_admin, is_anonymous, is_placeholder, created_at
		 FROM participants
		 WHERE user_or_anonymous_id = $1 AND party_id = $2 AND is_anonymous = FALSE`,
		userID, partyID,
	).Scan(
		&participant.ID,
		&participant.Name,
		&participant.IsAdmin,
		&participant.IsAnonymous,
		&participant.IsPlaceholder,
		&participant.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return objects.Participant{}, fmt.Errorf("участник не найден")
	}
	if err != nil {
		return objects.Participant{}, err
	}
	participant.PartyID = partyID
	participant.UserOrAnonymousID = &userID
	return participant, nil
}
func (s *Storage) GetParticipantByAnonID(anonID, partyID int) (objects.Participant, error) {
	// Принимает: id анонимного пользователя и id вечеринки
	// Делает: ищет участника-анонима в конкретной вечеринке
	// Возвращает: структуру Participant или ошибку
	var participant objects.Participant
	err := s.db.QueryRow(
		`SELECT id, name, is_admin, is_anonymous, is_placeholder, created_at
		 FROM participants
		 WHERE user_or_anonymous_id = $1 AND party_id = $2 AND is_anonymous = TRUE`,
		anonID, partyID,
	).Scan(
		&participant.ID,
		&participant.Name,
		&participant.IsAdmin,
		&participant.IsAnonymous,
		&participant.IsPlaceholder,
		&participant.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return objects.Participant{}, fmt.Errorf("участник не найден")
	}
	if err != nil {
		return objects.Participant{}, err
	}
	participant.PartyID = partyID
	participant.UserOrAnonymousID = &anonID
	return participant, nil
}
func (s *Storage) UpdateParticipant(participant objects.Participant) error {
	// Принимает: структуру Participant с обновлёнными данными
	// Делает: обновляет запись в таблице participants
	// Возвращает: только ошибку
	result, err := s.db.Exec(
		`UPDATE participants
		 SET user_or_anonymous_id = $1, name = $2, is_admin = $3, is_anonymous = $4, is_placeholder = $5
		 WHERE id = $6`,
		participant.UserOrAnonymousID,
		participant.Name,
		participant.IsAdmin,
		participant.IsAnonymous,
		participant.IsPlaceholder,
		participant.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("участник не найден")
	}
	return nil
}
func (s *Storage) DeleteParticipant(id int) error {
	// Принимает: id участника
	// Делает: удаляет запись из таблицы participants
	// Возвращает: только ошибку
	result, err := s.db.Exec("DELETE FROM participants WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("участник не найден")
	}
	return nil
}
