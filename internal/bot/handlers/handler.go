package handlers

import (
	"lotoMironBot/internal/services"

	"github.com/go-telegram/fsm"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	uService *services.UserService
	r        *redis.Client
	f        *fsm.FSM
}

func NewHandler(uService *services.UserService, f *fsm.FSM, r *redis.Client) *Handler {
	return &Handler{
		uService: uService,
		f:        f,
		r:        r,
	}
}
