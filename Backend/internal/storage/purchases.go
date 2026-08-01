package storage

import (
	"fmt"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

func (s *Storage) CreatePurchase(purchase objects.Purchase) (int, error) {
	//Принимает:  Делает:  Возвращает:
	var id int
	err := s.db.QueryRow(`
		INSERT INTO purchases (party_id, buyer_id, name, description, 
		purchase_icon_id, price, split_type, date_of_purchase) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id
		`,
		purchase.PartyID, purchase.BuyerID, purchase.Name, purchase.Description, purchase.PurchaseIconID, purchase.Price, purchase.SplitType, purchase.DateofPurchase).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func (s *Storage) GetPurchaseByID(id int) (objects.Purchase, error) {
	//Принимает:  Делает:  Возвращает:
	var Purchase objects.Purchase
	err := s.db.QueryRow(`
		SELECT party_id, buyer_id, name, description, purchase_icon_id,
		price, split_type, date_of_purchase, created_at 
		From purchases 
		WHERE id = $1
		`, id).Scan(
		&Purchase.PartyID,
		&Purchase.BuyerID,
		&Purchase.Name,
		&Purchase.Description,
		&Purchase.PurchaseIconID,
		&Purchase.Price,
		&Purchase.SplitType,
		&Purchase.DateofPurchase,
		&Purchase.CreatedAt)
	if err != nil {
		return objects.Purchase{}, err
	}
	Purchase.ID = id
	return Purchase, nil
}
func (s *Storage) GetPurchasesByPartyID(partyID int) ([]objects.Purchase, error) {
	//Принимает:  Делает:  Возвращает:
	rows, err := s.db.Query(`
		SELECT id, buyer_id, name, description, purchase_icon_id,
		price, split_type, date_of_purchase, created_at 
		From purchases 
		WHERE party_id = $1
		`, partyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var Purchases []objects.Purchase
	for rows.Next() {
		var Purchase objects.Purchase
		err := rows.Scan(
			&Purchase.ID,
			&Purchase.BuyerID,
			&Purchase.Name,
			&Purchase.Description,
			&Purchase.PurchaseIconID,
			&Purchase.Price,
			&Purchase.SplitType,
			&Purchase.DateofPurchase,
			&Purchase.CreatedAt)
		if err != nil {
			return nil, err
		}
		Purchase.PartyID = partyID
		Purchases = append(Purchases, Purchase)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return Purchases, nil
}
func (s *Storage) UpdatePurchase(Purchase objects.Purchase) error {
	//Принимает:  Делает:  Возвращает:
	result, err := s.db.Exec(`
		UPDATE purchases 
		SET buyer_id = $1, name = $2, description = $3, purchase_icon_id = $4, 
		price = $5, split_type = $6, date_of_purchase = $7 
		WHERE id = $8
		`,
		Purchase.BuyerID, Purchase.Name, Purchase.Description, Purchase.PurchaseIconID, Purchase.Price, Purchase.SplitType, Purchase.DateofPurchase, Purchase.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("Покупка не найдена")
	}
	return nil
}
func (s *Storage) DeletePurchase(id int) error {
	//Принимает:  Делает:  Возвращает:
	result, err := s.db.Exec(`
		DELETE FROM purchases 
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
		return fmt.Errorf("Покупка не найдена")
	}
	return nil
}
