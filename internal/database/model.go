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

type UserRole string

const (
	RolePlayer UserRole = "player"
	RoleAdmin  UserRole = "admin"
	RoleOwner  UserRole = "owner"
)

type User struct {
	TelegramID int64    `gorm:"primaryKey;autoIncrement:false"`
	Username   string   `gorm:"size:64"`
	FirstName  string   `gorm:"size:128"`
	LastName   string   `gorm:"size:128"`
	Role       UserRole `gorm:"type:varchar(16);not null;default:'player';index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

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
	ChatID  int64 `gorm:"not null;index"`

	Collection Collections `gorm:"type:varchar(32);not null;default:'standard'"`
	Status     GameStatus  `gorm:"type:varchar(16);not null;default:'pending';index"`

	ParticipantLimit  int    `gorm:"not null"`
	TicketsPerPlayer  int    `gorm:"not null;default:1"`
	WinnerPlayerID    *int64 `gorm:"index"`
	WinnerTicketCount int    `gorm:"not null;default:0"`
	LastDrawAt        *time.Time

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

type GameSettings struct {
	Key       string `gorm:"primaryKey;size:64"`
	Value     string `gorm:"type:text;not null"`
	UpdatedAt time.Time
}

// GameParticipant — игрок, зарегистрировавшийся в конкретной игре.
type GameParticipant struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	GameID               uint  `gorm:"not null;uniqueIndex:idx_game_participant"`
	PlayerID             int64 `gorm:"not null;uniqueIndex:idx_game_participant"`
	Failures             int   `gorm:"not null;default:0"`
	Disqualified         bool  `gorm:"not null;default:false"`
	LastAttemptDrawIndex int   `gorm:"not null;default:0"`

	CreatedAt time.Time
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
