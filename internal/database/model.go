package database

import (
	"time"

	"github.com/lib/pq"
)

type Collections string

const (
	Standard  Collections = "standard"
	NewYear   Collections = "new-year"
	Halloween Collections = "halloween"
)

type GameStatus string

const (
	GamePending  GameStatus = "pending"
	GameStarted  GameStatus = "started"
	GameCanceled GameStatus = "canceled"
	GameFinished GameStatus = "finished"
)

// Ticket — постоянный пул подготовленных билетов.
type Ticket struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	FileID string `gorm:"not null;uniqueIndex"`

	Numbers pq.Int32Array `gorm:"type:int[];not null"`

	Collection Collections `gorm:"type:varchar(32);not null;default:'standard'"`
}

// Game — конкретная игровая сессия.
type Game struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	// Telegram ID администратора, который создал/ведёт игру.
	AdminID int64 `gorm:"not null;index"`

	Collection Collections `gorm:"type:varchar(32);not null;default:'standard'"`
	Status     GameStatus  `gorm:"type:varchar(16);not null;default:'pending';index"`

	// Перемешанные числа 1-100.
	// Например: [73, 18, 46, 92, ...]
	Deck pq.Int32Array `gorm:"type:int[];not null"`

	// Индекс следующего бочонка.
	DrawIndex int `gorm:"not null;default:0"`

	CreatedAt  time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
	CanceledAt *time.Time
}

// GameTicket — конкретная выдача билета конкретному игроку
// в рамках конкретной игры.
type GameTicket struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	GameID   uint  `gorm:"not null;uniqueIndex:idx_game_ticket"`
	PlayerID int64 `gorm:"not null;index"`
	TicketID uint  `gorm:"not null;uniqueIndex:idx_game_ticket"`

	// Snapshot чисел билета на момент начала/выдачи игры.
	Numbers pq.Int32Array `gorm:"type:int[];not null"`

	// Числа, которые игрок подтвердил кнопкой «ЕСТЬ».
	MarkedNumbers pq.Int32Array `gorm:"type:int[];not null;default:'{}'"`

	// Все 6 чисел отмечены.
	Completed bool `gorm:"not null;default:false"`
}
