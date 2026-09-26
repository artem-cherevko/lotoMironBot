package handlers

import (
	"lotoMironBot/internal/services"

	"github.com/go-telegram/fsm"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	uService    *services.UserService
	lotoService *services.LotoService
	r           *redis.Client
	f           *fsm.FSM
}

func NewHandler(uService *services.UserService, f *fsm.FSM, r *redis.Client, lotoService *services.LotoService) *Handler {
	return &Handler{
		uService:    uService,
		f:           f,
		r:           r,
		lotoService: lotoService,
	}
}
