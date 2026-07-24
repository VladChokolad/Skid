package calculator

import (
	"fmt"
	"math"
	"sort"
)

type LightPayment struct {
	FromParticipantID int // Кто перевёл деньги
	ToParticipantID   int // Кому перевели деньги
	Amount            float64
}

type LightPurchase struct {
	BuyerID   int     // Participant который заплатил
	Price     float64 // Полная сумма покупки в рублях
	SplitType int
}

type LightDebt struct {
	PurchaseID    int // Покупка к которой относится долг
	ParticipantID int // Участник который должен
	SplitValue    float64
}

type LightParticipant struct {
	ID   int
	Name string
}

type Balance struct {
	ParticipantID int
	Name          string
	Paid          float64 // сколько заплатил
	Owed          float64 // сколько должен был
	Balance       float64 // разница (положительная — ему должны)
}

type Settlement struct {
	FromParticipantID int
	ToParticipantID   int
	Amount            float64
}

type OutportFromCalc struct {
	Balances    []Balance
	Settlements []Settlement
}

type ImportToCalc struct {
	LightPayments     []LightPayment
	LightPurchases    []LightPurchase
	LightParticipants []LightParticipant
	LightDebts        []LightDebt
}

func Calculation(in ImportToCalc) (OutportFromCalc, error) {
	if len(in.LightParticipants) == 0 {
		return OutportFromCalc{}, fmt.Errorf("нет участников")
	}
	// Шаг 1 — инициализация
	names := make(map[int]string)
	paid := make(map[int]float64)
	owed := make(map[int]float64)

	// Шаг 2 — заполнить из участников
	for _, p := range in.LightParticipants {
		names[p.ID] = p.Name
		paid[p.ID] = 0
		owed[p.ID] = 0
	}

	// Шаг 3 — покупки увеличивают paid покупателя
	for _, p := range in.LightPurchases {
		paid[p.BuyerID] += p.Price
	}

	// Шаг 4 — долги увеличивают owed должника
	for _, d := range in.LightDebts {
		owed[d.ParticipantID] += d.SplitValue
	}

	// Шаг 5 — платежи корректируют балансы
	balances := make(map[int]float64)
	for id := range names {
		balances[id] = paid[id] - owed[id]
	}
	for _, p := range in.LightPayments {
		balances[p.FromParticipantID] += p.Amount
		balances[p.ToParticipantID] -= p.Amount
	}

	// Шаг 6 — собрать Balance для каждого участника
	var balanceList []Balance
	for _, p := range in.LightParticipants {
		balanceList = append(balanceList, Balance{
			ParticipantID: p.ID,
			Name:          names[p.ID],
			Paid:          round2(paid[p.ID]),
			Owed:          round2(owed[p.ID]),
			Balance:       round2(balances[p.ID]),
		})
	}

	// Шаг 7 — минимизировать переводы
	settlements := minimizeTransfers(balances)

	// Шаг 8 — вернуть результат
	return OutportFromCalc{
		Balances:    balanceList,
		Settlements: settlements,
	}, nil
}

func round2(x float64) float64 {
	if x < 0 {
		return float64(int(x*100-0.5)) / 100
	}
	return float64(int(x*100+0.5)) / 100
}

type entry struct {
	id     int
	amount float64
}

func minimizeTransfers(balances map[int]float64) []Settlement {
	var debtors, creditors []entry

	// Разделение на должников и кредиторов
	for id, balance := range balances {
		rounded := round2(balance)
		if rounded < 0 {
			debtors = append(debtors, entry{id, -rounded}) // положительное число долга
		} else if rounded > 0 {
			creditors = append(creditors, entry{id, rounded})
		}
		// balance == 0 — участник в расчёте, пропускаем
	}

	// Сортировка для детерминированного результата
	sort.Slice(debtors, func(i, j int) bool { return debtors[i].id < debtors[j].id })
	sort.Slice(creditors, func(i, j int) bool { return creditors[i].id < creditors[j].id })

	var settlements []Settlement
	i, j := 0, 0

	for i < len(debtors) && j < len(creditors) {
		// Переводим минимум
		pay := math.Min(debtors[i].amount, creditors[j].amount)

		settlements = append(settlements, Settlement{
			FromParticipantID: debtors[i].id,
			ToParticipantID:   creditors[j].id,
			Amount:            round2(pay),
		})

		debtors[i].amount -= pay
		creditors[j].amount -= pay

		// Переход к следующему если закрыт
		if round2(debtors[i].amount) == 0 {
			i++
		}
		if round2(creditors[j].amount) == 0 {
			j++
		}
	}

	return settlements
}
