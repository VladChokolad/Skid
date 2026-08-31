package storage

import (
	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

// Принимает: структуру Debt с id покупки, id участника и долей
// Делает: создаёт запись долга в таблице debts
// Возвращает: id созданной записи или ошибку
func (s *Storage) CreateDebt(debt objects.Debt) (int, error) {
	var id int
	err := s.db.QueryRow(`
		INSERT INTO debts (purchase_id, participant_id, split_value) 
		VALUES ($1, $2, $3) 
		RETURNING id
		`, debt.PurchaseID, debt.ParticipantID, debt.SplitValue).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Принимает: id покупки
// Делает: ищет все долги связанные с этой покупкой
// Возвращает: слайс структур Debt или ошибку
func (s *Storage) GetDebtsByPurchaseID(purchaseID int) ([]objects.Debt, error) {
	rows, err := s.db.Query(`
		SELECT id, participant_id, split_value
		FROM debts
		WHERE purchase_id = $1
		`, purchaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debts []objects.Debt
	for rows.Next() {
		var debt objects.Debt
		err := rows.Scan(
			&debt.ID,
			&debt.ParticipantID,
			&debt.SplitValue,
		)
		if err != nil {
			return nil, err
		}
		debt.PurchaseID = purchaseID
		debts = append(debts, debt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return debts, nil
}

// Принимает: id участника
// Делает: ищет все долги этого участника по всем покупкам
// Возвращает: слайс структур Debt или ошибку
func (s *Storage) GetDebtsByParticipantID(participantID int) ([]objects.Debt, error) {
	rows, err := s.db.Query(`
		SELECT id, purchase_id, split_value
		FROM debts
		WHERE participant_id = $1
		`, participantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debts []objects.Debt
	for rows.Next() {
		var debt objects.Debt
		err := rows.Scan(
			&debt.ID,
			&debt.PurchaseID,
			&debt.SplitValue,
		)
		if err != nil {
			return nil, err
		}
		debt.ParticipantID = participantID
		debts = append(debts, debt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return debts, nil
}

// GetDebtsByEventID возвращает все долги для участников указанной тусовки.
func (s *Storage) GetDebtsByPartyID(partyID int) ([]objects.Debt, error) {
	query := `
        SELECT d.id, d.participant_id, d.purchase_id, d.split_value
        FROM debts d
        JOIN purchases p ON d.purchase_id = p.id
        WHERE p.party_id = $1
    `
	rows, err := s.db.Query(query, partyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debts []objects.Debt
	for rows.Next() {
		var debt objects.Debt
		err := rows.Scan(
			&debt.ID,
			&debt.ParticipantID,
			&debt.PurchaseID,
			&debt.SplitValue,
		)
		if err != nil {
			return nil, err
		}
		debts = append(debts, debt)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return debts, nil
}

// Принимает: id покупки
// Делает: удаляет все долги связанные с этой покупкой (если их нет — не ошибка)
// Возвращает: ошибку только при сбое запроса к БД
func (s *Storage) DeleteDebtsByPurchaseID(purchaseID int) error {
	_, err := s.db.Exec(`
	DELETE FROM debts
	WHERE purchase_id = $1
	`, purchaseID)
	return err
}

// Принимает: id участника
// Делает: удаляет все долги этого участника по всем покупкам (если их нет — не ошибка)
// Возвращает: ошибку только при сбое запроса к БД
func (s *Storage) DeleteDebtsByParticipantID(participantID int) error {
	_, err := s.db.Exec(`
	DELETE FROM debts
	WHERE participant_id = $1
	`, participantID)
	return err
}
