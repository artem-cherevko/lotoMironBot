package services

import "lotoMironBot/internal/repository"

type UserService struct {
	repo *repository.UserRepo
}

func NewUserService(repo *repository.UserRepo) *UserService {
	return &UserService{
		repo: repo,
	}
}
