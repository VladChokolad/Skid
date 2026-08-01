package storage

import (
	"database/sql"
	"fmt"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

func (s *Storage) CreatePayment(p objects.Payment) (int, error) {
	// Принимает: структуру Payment
	// Делает: создаёт запись в таблице payments
	// Возвращает: id созданного перевода или ошибку
	var id int
	err := s.db.QueryRow(`
		INSERT INTO payments (party_id, from_participant_id, to_participant_id, amount, note, is_confirmed)
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id
		`,
		p.PartyID,
		p.FromParticipantID,
		p.ToParticipantID,
		p.Amount,
		p.Note,
		p.IsConfirmed,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func (s *Storage) GetPaymentByID(id int) (objects.Payment, error) {
	// Принимает: id перевода
	// Делает: ищет перевод по id
	// Возвращает: структуру Payment или ошибку
	var payment objects.Payment
	err := s.db.QueryRow(`
		SELECT party_id, from_participant_id, to_participant_id, amount, note, is_confirmed, created_at
		FROM payments
		WHERE id = $1
		`, id).Scan(
		&payment.PartyID,
		&payment.FromParticipantID,
		&payment.ToParticipantID,
		&payment.Amount,
		&payment.Note,
		&payment.IsConfirmed,
		&payment.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return objects.Payment{}, fmt.Errorf("перевод не найден")
	}
	if err != nil {
		return objects.Payment{}, err
	}
	payment.ID = id
	return payment, nil
}
func (s *Storage) GetPaymentsByPartyID(partyID int) ([]objects.Payment, error) {
	// Принимает: id вечеринки
	// Делает: ищет все переводы внутри вечеринки
	// Возвращает: список переводов или ошибку
	rows, err := s.db.Query(`
		SELECT id, from_participant_id, to_participant_id, amount, note, is_confirmed, created_at
		FROM payments
		WHERE party_id = $1
		`, partyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []objects.Payment
	for rows.Next() {
		var payment objects.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.FromParticipantID,
			&payment.ToParticipantID,
			&payment.Amount,
			&payment.Note,
			&payment.IsConfirmed,
			&payment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		payment.PartyID = partyID
		payments = append(payments, payment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payments, nil
}
func (s *Storage) ConfirmPayment(id int) error {
	// Принимает: id перевода
	// Делает: помечает перевод как подтверждённый получателем
	// Возвращает: только ошибку
	result, err := s.db.Exec(`
		UPDATE payments 
		SET is_confirmed = TRUE 
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
		return fmt.Errorf("перевод не найден")
	}
	return nil
}
func (s *Storage) DeletePayment(id int) error {
	// Принимает: id перевода
	// Делает: удаляет запись из таблицы payments
	// Возвращает: только ошибку
	result, err := s.db.Exec(`
		DELETE FROM payments 
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
		return fmt.Errorf("перевод не найден")
	}
	return nil
}
