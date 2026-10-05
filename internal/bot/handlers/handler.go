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
	adminID     string
}

func NewHandler(uService *services.UserService, f *fsm.FSM, r *redis.Client, lotoService *services.LotoService, adminID string) *Handler {
	return &Handler{
		uService:    uService,
		f:           f,
		r:           r,
		lotoService: lotoService,
		adminID:     adminID,
	}
}
