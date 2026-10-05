package repository

import (
	"context"
	"lotoMironBot/internal/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LotoRepository struct {
	db *gorm.DB
}

func NewLotoRepo(db *gorm.DB) *LotoRepository {
	return &LotoRepository{
		db: db,
	}
}

func (r *LotoRepository) WithinTransaction(ctx context.Context, fn func(*LotoRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&LotoRepository{db: tx})
	})
}

func (r *LotoRepository) LockAdmin(ctx context.Context, adminID int64) error {
	return r.db.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(?)", adminID).Error
}

func (r *LotoRepository) LockChat(ctx context.Context, chatID int64) error {
	return r.db.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(?)", chatID).Error
}

func (r *LotoRepository) AddTicket(ctx context.Context, ticket *database.Ticket) (*database.Ticket, error) {
	if err := r.db.WithContext(ctx).Create(ticket).Error; err != nil {
		return nil, err
	}

	return ticket, nil
}

func (r *LotoRepository) GetAllTickets(ctx context.Context) ([]*database.Ticket, error) {
	var tickets []*database.Ticket
	if err := r.db.WithContext(ctx).Find(&tickets).Error; err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *LotoRepository) CreateGame(ctx context.Context, game *database.Game) (*database.Game, error) {
	if err := r.db.WithContext(ctx).Create(game).Error; err != nil {
		return nil, err
	}
	return game, nil
}

func (r *LotoRepository) GetGameByID(ctx context.Context, gameID uint) (*database.Game, error) {
	var game database.Game
	if err := r.db.WithContext(ctx).First(&game, gameID).Error; err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *LotoRepository) FindActiveGameByAdmin(ctx context.Context, adminID int64) (*database.Game, error) {
	var game database.Game
	err := r.db.WithContext(ctx).
		Where("admin_id = ? AND status IN ?", adminID, []database.GameStatus{database.GamePending, database.GameStarted}).
		Order("id DESC").
		First(&game).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *LotoRepository) FindActiveGameByChat(ctx context.Context, chatID int64) (*database.Game, error) {
	var game database.Game
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND status IN ?", chatID, []database.GameStatus{database.GamePending, database.GameStarted}).
		Order("id DESC").
		First(&game).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *LotoRepository) LockActiveGameByChat(ctx context.Context, chatID int64) (*database.Game, error) {
	var game database.Game
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("chat_id = ? AND status IN ?", chatID, []database.GameStatus{database.GamePending, database.GameStarted}).
		Order("id DESC").
		First(&game).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *LotoRepository) SaveGame(ctx context.Context, game *database.Game) error {
	return r.db.WithContext(ctx).Save(game).Error
}

func (r *LotoRepository) GetGameParticipant(ctx context.Context, gameID uint, playerID int64) (*database.GameParticipant, error) {
	var participant database.GameParticipant
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("game_id = ? AND player_id = ?", gameID, playerID).
		First(&participant).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

func (r *LotoRepository) SaveGameParticipant(ctx context.Context, participant *database.GameParticipant) error {
	return r.db.WithContext(ctx).Save(participant).Error
}

func (r *LotoRepository) CountEligibleParticipants(ctx context.Context, gameID uint, adminID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.GameParticipant{}).
		Where("game_id = ? AND player_id <> ? AND disqualified = ?", gameID, adminID, false).
		Count(&count).Error
	return count, err
}

func (r *LotoRepository) CreateGameParticipant(ctx context.Context, participant *database.GameParticipant) error {
	return r.db.WithContext(ctx).Create(participant).Error
}

func (r *LotoRepository) IsGameParticipant(ctx context.Context, gameID uint, playerID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.GameParticipant{}).
		Where("game_id = ? AND player_id = ?", gameID, playerID).
		Count(&count).Error
	return count > 0, err
}

func (r *LotoRepository) CountGameParticipants(ctx context.Context, gameID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.GameParticipant{}).Where("game_id = ?", gameID).Count(&count).Error
	return count, err
}

func (r *LotoRepository) ListGameParticipants(ctx context.Context, gameID uint) ([]*database.GameParticipant, error) {
	var participants []*database.GameParticipant
	err := r.db.WithContext(ctx).Where("game_id = ?", gameID).Order("id ASC").Find(&participants).Error
	return participants, err
}

func (r *LotoRepository) GetGameTickets(ctx context.Context, gameID uint) ([]*database.GameTicket, error) {
	var tickets []*database.GameTicket
	if err := r.db.WithContext(ctx).Where("game_id = ?", gameID).Order("id ASC").Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *LotoRepository) ListPlayerGameTickets(ctx context.Context, gameID uint, playerID int64) ([]*database.GameTicket, error) {
	var tickets []*database.GameTicket
	err := r.db.WithContext(ctx).
		Where("game_id = ? AND player_id = ?", gameID, playerID).
		Order("id ASC").
		Find(&tickets).Error
	return tickets, err
}

func (r *LotoRepository) GetTicketByID(ctx context.Context, ticketID uint) (*database.Ticket, error) {
	var ticket database.Ticket
	if err := r.db.WithContext(ctx).First(&ticket, ticketID).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *LotoRepository) ListLatestPlayerGameTickets(ctx context.Context, playerID int64) ([]*database.GameTicket, error) {
	var tickets []*database.GameTicket
	err := r.db.WithContext(ctx).Where("player_id = ? AND game_id = (SELECT id FROM games WHERE status IN ? AND id IN (SELECT game_id FROM game_participants WHERE player_id = ?) ORDER BY id DESC LIMIT 1)", playerID, []database.GameStatus{database.GamePending, database.GameStarted}, playerID).Order("id ASC").Find(&tickets).Error
	return tickets, err
}

func (r *LotoRepository) CountPlayerGameTickets(ctx context.Context, gameID uint, playerID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.GameTicket{}).
		Where("game_id = ? AND player_id = ?", gameID, playerID).
		Count(&count).Error
	return count, err
}

func (r *LotoRepository) SaveSettings(ctx context.Context, key, value string) error {
	setting := database.GameSettings{Key: key, Value: value}
	return r.db.WithContext(ctx).Save(&setting).Error
}

func (r *LotoRepository) GetSettings(ctx context.Context, key string) (string, error) {
	var setting database.GameSettings
	err := r.db.WithContext(ctx).First(&setting, "key = ?", key).Error
	return setting.Value, err
}

func (r *LotoRepository) ListAdmins(ctx context.Context) ([]*database.User, error) {
	var users []*database.User
	err := r.db.WithContext(ctx).Where("role = ?", database.RoleAdmin).Order("telegram_id ASC").Find(&users).Error
	return users, err
}

func (r *LotoRepository) ListActiveGameStats(ctx context.Context, chatID int64) (*database.Game, []*database.GameParticipant, []*database.GameTicket, error) {
	game, err := r.FindActiveGameByChat(ctx, chatID)
	if err != nil {
		return nil, nil, nil, err
	}
	participants, err := r.ListGameParticipants(ctx, game.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	tickets, err := r.GetGameTickets(ctx, game.ID)
	return game, participants, tickets, err
}

func (r *LotoRepository) ListLatestGameStats(ctx context.Context, chatID int64) (*database.Game, []*database.GameParticipant, []*database.GameTicket, error) {
	var game database.Game
	if err := r.db.WithContext(ctx).Where("chat_id = ?", chatID).Order("id DESC").First(&game).Error; err != nil {
		return nil, nil, nil, err
	}
	participants, err := r.ListGameParticipants(ctx, game.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	tickets, err := r.GetGameTickets(ctx, game.ID)
	return &game, participants, tickets, err
}

func (r *LotoRepository) ListWinners(ctx context.Context, gameID uint) ([]*database.Game, error) {
	var games []*database.Game
	err := r.db.WithContext(ctx).Where("id = ? AND winner_player_id IS NOT NULL", gameID).Find(&games).Error
	return games, err
}

func (r *LotoRepository) ListAvailableTickets(ctx context.Context, gameID uint, collection database.Collections, limit int) ([]*database.Ticket, error) {
	var tickets []*database.Ticket
	err := r.db.WithContext(ctx).
		Where("collection = ?", collection).
		Where("NOT EXISTS (SELECT 1 FROM game_tickets WHERE game_tickets.game_id = ? AND game_tickets.ticket_id = tickets.id)", gameID).
		Order("id ASC").
		Limit(limit).
		Find(&tickets).Error
	return tickets, err
}

func (r *LotoRepository) CreateGameTicket(ctx context.Context, ticket *database.GameTicket) (*database.GameTicket, error) {
	if err := r.db.WithContext(ctx).Create(ticket).Error; err != nil {
		return nil, err
	}
	return ticket, nil
}

func (r *LotoRepository) SaveGameTicket(ctx context.Context, ticket *database.GameTicket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}
