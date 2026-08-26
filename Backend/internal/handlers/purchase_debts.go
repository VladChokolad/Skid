package handlers

import (
	"fmt"
	"math"
	"sort"

	"github.com/VladChokolad/Skid/Backend/internal/objects"
)

// buildDebts считает индивидуальные доли участников для покупки в соответствии
// со splitType. purchaseID может быть 0 на момент вызова (до вставки покупки
// в БД) — вызывающий код обязан проставить его в каждый Debt перед сохранением.
func buildDebts(purchaseID int, price float64, splitType int, allParticipants []objects.Participant, debtorIDs []int, debtorAmounts map[int]float64) ([]objects.Debt, error) {
	validIDs := make(map[int]bool, len(allParticipants))
	allIDs := make([]int, 0, len(allParticipants))
	for _, p := range allParticipants {
		validIDs[p.ID] = true
		allIDs = append(allIDs, p.ID)
	}

	switch splitType {
	case 0: // поровну всем участникам тусовки
		return splitEqually(purchaseID, price, allIDs)

	case 1: // поровну выбранным
		if len(debtorIDs) == 0 {
			return nil, fmt.Errorf(`для разбивки "поровну выбранным" укажите debtors`)
		}
		for _, id := range debtorIDs {
			if !validIDs[id] {
				return nil, fmt.Errorf("участник %d не найден в этой тусовке", id)
			}
		}
		return splitEqually(purchaseID, price, debtorIDs)

	case 2: // индивидуальные суммы — debtorAmounts[id] это сумма в рублях, сумма должна сойтись с price
		if len(debtorAmounts) == 0 {
			return nil, fmt.Errorf("для индивидуальных сумм укажите debtorAmounts")
		}
		for id := range debtorAmounts {
			if !validIDs[id] {
				return nil, fmt.Errorf("участник %d не найден в этой тусовке", id)
			}
		}
		return splitByExactAmounts(purchaseID, price, debtorAmounts)

	case 3: // индивидуальные доли — debtorAmounts[id] это вес (не обязательно нормированный к 1 или 100)
		if len(debtorAmounts) == 0 {
			return nil, fmt.Errorf("для индивидуальных долей укажите debtorAmounts")
		}
		for id := range debtorAmounts {
			if !validIDs[id] {
				return nil, fmt.Errorf("участник %d не найден в этой тусовке", id)
			}
		}
		return splitByShares(purchaseID, price, debtorAmounts)

	default:
		return nil, fmt.Errorf("неизвестный тип разбивки: %d", splitType)
	}
}

// splitEqually делит price поровну между ids, раздавая копейки округления
// первым участникам по возрастанию id — так сумма долей всегда точно равна price.
func splitEqually(purchaseID int, price float64, ids []int) ([]objects.Debt, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("нет участников для разбивки")
	}
	sorted := append([]int(nil), ids...)
	sort.Ints(sorted)

	totalCents := int64(math.Round(price * 100))
	n := int64(len(sorted))
	base := totalCents / n
	remainder := totalCents % n

	debts := make([]objects.Debt, 0, len(sorted))
	for i, id := range sorted {
		cents := base
		if int64(i) < remainder {
			cents++
		}
		debts = append(debts, objects.Debt{
			PurchaseID:    purchaseID,
			ParticipantID: id,
			SplitValue:    float64(cents) / 100,
		})
	}
	return debts, nil
}

// splitByExactAmounts требует, чтобы суммы участников точно (с точностью до копейки) сходились с ценой.
func splitByExactAmounts(purchaseID int, price float64, amounts map[int]float64) ([]objects.Debt, error) {
	var sum float64
	debts := make([]objects.Debt, 0, len(amounts))
	for id, amount := range amounts {
		if amount <= 0 {
			return nil, fmt.Errorf("сумма участника %d должна быть больше нуля", id)
		}
		sum += amount
		debts = append(debts, objects.Debt{PurchaseID: purchaseID, ParticipantID: id, SplitValue: round2(amount)})
	}
	if math.Abs(sum-price) > 0.01 {
		return nil, fmt.Errorf("сумма долей (%.2f) не совпадает с ценой покупки (%.2f)", sum, price)
	}
	return debts, nil
}

// splitByShares делит price пропорционально весам (не обязательно нормированным к 1 или 100).
// Последнему участнику (по id) достаётся остаток копеек, чтобы сумма точно сходилась с price.
func splitByShares(purchaseID int, price float64, shares map[int]float64) ([]objects.Debt, error) {
	var totalWeight float64
	ids := make([]int, 0, len(shares))
	for id, w := range shares {
		if w <= 0 {
			return nil, fmt.Errorf("доля участника %d должна быть больше нуля", id)
		}
		totalWeight += w
		ids = append(ids, id)
	}
	sort.Ints(ids)

	totalCents := int64(math.Round(price * 100))
	var allocated int64
	debts := make([]objects.Debt, 0, len(ids))
	for i, id := range ids {
		var cents int64
		if i == len(ids)-1 {
			cents = totalCents - allocated
		} else {
			cents = int64(math.Round(shares[id] / totalWeight * float64(totalCents)))
			allocated += cents
		}
		debts = append(debts, objects.Debt{PurchaseID: purchaseID, ParticipantID: id, SplitValue: float64(cents) / 100})
	}
	return debts, nil
}

func round2(x float64) float64 {
	return math.Round(x*100) / 100
}
