package services

import (
	"context"
	"lotoMironBot/internal/database"
	"lotoMironBot/internal/repository"

	"github.com/lib/pq"
)

type LotoService struct {
	repo *repository.LotoRepository
}

func NewLotoService(repo *repository.LotoRepository) *LotoService {
	return &LotoService{
		repo: repo,
	}
}

func (s *LotoService) AddTicket(ctx context.Context, fileID, collection string, numbers []int32) (*database.Ticket, error) {
	ticket := &database.Ticket{
		FileID:     fileID,
		Collection: database.Collections(collection),
		Numbers:    pq.Int32Array(numbers),
	}

	addedTicket, err := s.repo.AddTicket(ctx, ticket)
	if err != nil {
		return nil, err
	}

	return addedTicket, nil
}

func (s *LotoService) GetAllTickets(ctx context.Context) ([]*database.Ticket, error) {
	return s.repo.GetAllTickets(ctx)
}
