package services

import (
	"context"
	"lotoMironBot/internal/database"
	"lotoMironBot/internal/repository"

	"github.com/go-telegram/bot/models"
)

type UserService struct {
	repo *repository.UserRepo
}

func NewUserService(repo *repository.UserRepo) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) Register(ctx context.Context, telegramUser models.User, bootstrapOwnerID int64) error {
	return s.repo.Register(ctx, &database.User{
		TelegramID: telegramUser.ID,
		Username:   telegramUser.Username,
		FirstName:  telegramUser.FirstName,
		LastName:   telegramUser.LastName,
		Role:       database.RolePlayer,
	}, bootstrapOwnerID != 0 && telegramUser.ID == bootstrapOwnerID)
}

func (s *UserService) Get(ctx context.Context, telegramID int64) (*database.User, error) {
	return s.repo.GetByTelegramID(ctx, telegramID)
}

func (s *UserService) SetRole(ctx context.Context, telegramID int64, role database.UserRole) error {
	return s.repo.SetRole(ctx, telegramID, role)
}

func (s *UserService) IsAdmin(ctx context.Context, telegramID int64) (bool, error) {
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return false, err
	}
	return user.Role == database.RoleAdmin || user.Role == database.RoleOwner, nil
}

func (s *UserService) IsOwner(ctx context.Context, telegramID int64) (bool, error) {
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return false, err
	}
	return user.Role == database.RoleOwner, nil
}
