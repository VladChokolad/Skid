package objects

import "time"

//НАЧАЛО СТРУКТУР БАЗЫ ДАННЫХ
type User struct { // Зарегистрированный пользователь
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`            // Хэш пароля — никогда не отдаём наружу
	Phone        *string   `json:"phone"`        // Опционально — для будущих уведомлений
	ProfileImage *string   `json:"profileImage"` // URL аватара пользователя
	CreatedAt    time.Time `json:"createdAt"`    // Дата регистрации
}

type AnonymousUser struct { // Анонимный пользователь — участвует без регистрации
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Phone        *string   `json:"phone"`        // Опционально
	CreatedAt    time.Time `json:"createdAt"`    // Дата первого входа
	LastActivity time.Time `json:"lastActivity"` //Дата последней активности
}

type Party struct { // Вечеринка
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	PartyImage  *string   `json:"partyImage"` // URL обложки вечеринки
	OwnerID     int       `json:"ownerID"`    // ID User который является владельцем
	InviteCode  string    `json:"inviteCode"` // Уникальный код по которому можно присоединиться — используется в пригласительной ссылке
	IsActive    bool      `json:"isActive"`   // Активна ли вечеринка. Неактивнные вечеринки не могут быть изменены, но их можно просматривать
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"` // Обновляется при любом изменении вечеринки
}

type Participant struct { // Участник вечеринки
	ID                int       `json:"id"`
	PartyID           int       `json:"partyID"`           // Вечеринка к которой принадлежит участник
	UserOrAnonymousID *int      `json:"userOrAnonymousID"` // ID User или AnonymousUser — nil если Placeholder
	Name              string    `json:"name"`
	IsAdmin           bool      `json:"isAdmin"`       // Имеет расширенные права управления вечеринкой
	IsAnonymous       bool      `json:"isAnonymous"`   // Указывает что UserOrAnonymousID ссылается на AnonymousUser
	IsPlaceholder     bool      `json:"isPlaceholder"` // Пустышка созданная Owner — ждёт реального человека
	CreatedAt         time.Time `json:"createdAt"`
}

type Purchase struct { // Покупка
	ID             int        `json:"id"`
	PartyID        int        `json:"partyID"` // Вечеринка в которой сделана покупка
	BuyerID        int        `json:"buyerID"` // Participant который заплатил
	Name           string     `json:"name"`
	Description    *string    `json:"description"`
	PurchaseIconID *int       `json:"purchaseIconId"` // Иконка покупки — для визуального отображения в списке
	Price          float64    `json:"price"`          // Полная сумма покупки в рублях
	SplitType      int        `json:"splitType"`      // 0 - поровну всем, 1 - поровну выбранным, 2 - индивидуальные суммы, 3 - индивидуальные доли
	DateofPurchase *time.Time `json:"dateOfPurchase"` // Когда совершена покупка — может отличаться от CreatedAt
	CreatedAt      time.Time  `json:"createdAt"`      // Когда запись добавлена в систему
}

type Debt struct { // Долг участника по конкретной покупке
	ID            int     `json:"id"`
	PurchaseID    int     `json:"purchaseID"`    // Покупка к которой относится долг
	ParticipantID int     `json:"participantID"` // Участник который должен
	SplitValue    float64 `json:"splitValue"`    // Доля от общей суммы покупки которую должен заплатить участник
}

type Payment struct { // Перевод между участниками
	ID                int       `json:"id"`
	PartyID           int       `json:"partyID"`           // Вечеринка в которой совершён перевод
	FromParticipantID int       `json:"fromParticipantID"` // Кто перевёл деньги
	ToParticipantID   int       `json:"toParticipantID"`   // Кому перевели деньги
	Amount            float64   `json:"amount"`            // Сумма перевода в рублях
	Note              string    `json:"note"`              // Комментарий — например «перевёл через СБП»
	IsConfirmed       bool      `json:"isConfirmed"`       // Получатель подтвердил что получил деньги
	CreatedAt         time.Time `json:"createdAt"`
}
type PurchaseIcon struct { // Иконка покупки — для визуального отображения в списке
	ID   int    `json:"id"`
	Name string `json:"name"` // Название иконки — например «Пицца», «Такси», «Кино» и т.д.
	Icon string `json:"icon"` // URL иконки
}
