package repository

import (
	"context"
	"lotoMironBot/internal/database"

	"gorm.io/gorm"
)

type LotoRepository struct {
	db *gorm.DB
}

func NewLotoRepo(db *gorm.DB) *LotoRepository {
	return &LotoRepository{
		db: db,
	}
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

func (r *LotoRepository) GetGameTickets(ctx context.Context, gameID uint) ([]*database.GameTicket, error) {
	var tickets []*database.GameTicket
	if err := r.db.WithContext(ctx).Find(&tickets, gameID).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *LotoRepository) CreateGameTicket(ctx context.Context, game *database.GameTicket) (*database.GameTicket, error) {
	if err := r.db.WithContext(ctx).Create(game).Error; err != nil {
		return nil, err
	}
	return game, nil
}
