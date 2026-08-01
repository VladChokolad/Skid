package activity

import (
	"sync"
	"time"
)

var (
	cache sync.Map
	// Интервал между обновлениями в БД (5 минут)
	updateInterval = 5 * time.Minute
)

// ShouldUpdate проверяет, нужно ли обновить last_activity для анонима.
// Возвращает true, если с последнего обновления прошло больше updateInterval.
// Также обновляет время в кэше, чтобы не делать частых вызовов.
func ShouldUpdate(anonID int) bool {
	now := time.Now()
	if last, ok := cache.Load(anonID); ok {
		if now.Sub(last.(time.Time)) < updateInterval {
			return false
		}
	}
	cache.Store(anonID, now)
	return true
}

// SetUpdateInterval позволяет изменить интервал (для тестов).
func SetUpdateInterval(d time.Duration) {
	updateInterval = d
}
