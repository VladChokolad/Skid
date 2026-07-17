package objects

import "time"

//НАЧАЛО СТРУКТУР БАЗЫ ДАННЫХ
type User struct { // Зарегистрированный пользователь
	ID           int
	Name         string
	Email        string
	PasswordHash string    // Хэш пароля — никогда не храним пароль в открытом виде
	Phone        *string   // Опционально — для будущих уведомлений
	ProfileImage *string   // URL аватара пользователя
	CreatedAt    time.Time // Дата регистрации
}

type AnonymousUser struct { // Анонимный пользователь — участвует без регистрации
	ID           int
	Name         string
	Phone        *string // Опционально — для матчинга с Placeholder
	Access_token string
	CreatedAt    time.Time // Дата первого входа
}

type Party struct { // Вечеринка
	ID          int
	Name        string
	Description string
	PartyImage  *string // URL обложки вечеринки
	OwnerID     int     // ID User который является владельцем
	InviteCode  string  // Уникальный код по которому можно присоединиться — используется в пригласительной ссылке
	IsActive    bool    // Активна ли вечеринка. Неактивнные вечеринки не могут быть изменены, но их можно просматривать
	CreatedAt   time.Time
	UpdatedAt   time.Time // Обновляется при любом изменении вечеринки
}

type Participant struct { // Участник вечеринки
	ID                int
	PartyID           int  // Вечеринка к которой принадлежит участник
	UserOrAnonymousID *int // ID User или AnonymousUser — nil если Placeholder
	Name              string
	IsAdmin           bool // Имеет расширенные права управления вечеринкой
	IsAnonymous       bool // Указывает что UserOrAnonymousID ссылается на AnonymousUser
	IsPlaceholder     bool // Пустышка созданная Owner — ждёт реального человека
	CreatedAt         time.Time
}

type Purchase struct { // Покупка
	ID             int
	PartyID        int // Вечеринка в которой сделана покупка
	BuyerID        int // Participant который заплатил
	Name           string
	Description    *string
	PurchaseIconID *int       // Иконка покупки — для визуального отображения в списке
	Price          float64    // Полная сумма покупки в рублях
	SplitType      int        // 0 - поровну всем, 1 - поровну выбранным, 2 - индивидуальные суммы, 3 - индивидуальные доли
	DateofPurchase *time.Time // Когда совершена покупка — может отличаться от CreatedAt
	CreatedAt      time.Time  // Когда запись добавлена в систему
}

type Debt struct { // Долг участника по конкретной покупке
	ID            int
	PurchaseID    int     // Покупка к которой относится долг
	ParticipantID int     // Участник который должен
	SplitAmount   float64 // Доля от общей суммы покупки которую должен заплатить участник
}

type Payment struct { // Перевод между участниками
	ID                int
	PartyID           int     // Вечеринка в которой совершён перевод
	FromParticipantID int     // Кто перевёл деньги
	ToParticipantID   int     // Кому перевели деньги
	Amount            float64 // Сумма перевода в рублях
	Note              string  // Комментарий — например «перевёл через СБП»
	IsConfirmed       bool    // Получатель подтвердил что получил деньги
	CreatedAt         time.Time
}
type PurchaseIcon struct { // Иконка покупки — для визуального отображения в списке
	ID   int
	Name string // Название иконки — например «Пицца», «Такси», «Кино» и т.д.
	Icon string // URL иконки
}

//КОНЕЦ СТРУКТУР БАЗЫ ДАННЫХ

type Settlement struct {
	ID                int
	PartyID           int
	FromParticipantID int
	ToParticipantID   int
	Amount            float64
	CreatedAt         time.Time
}
