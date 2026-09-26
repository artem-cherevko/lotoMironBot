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
