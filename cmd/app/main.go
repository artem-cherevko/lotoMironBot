package main

import (
	"context"
	"lotoMironBot/internal/bot"
	fsm2 "lotoMironBot/internal/bot/fsm"
	"lotoMironBot/internal/bot/handlers"
	"lotoMironBot/internal/config"
	"lotoMironBot/internal/database"
	"lotoMironBot/internal/redis"
	"lotoMironBot/internal/repository"
	"lotoMironBot/internal/services"
	"lotoMironBot/internal/worker"

	"github.com/go-telegram/fsm"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Connect to db
	db, err := database.Connect(cfg.DBDsn)
	if err != nil {
		panic(err)
	}

	// Connect Redis
	rdb, err := redis.Connect(cfg.RedisAddr)
	if err != nil {
		panic(err)
	}

	// User
	uRepo := repository.NewUserRepo(db)
	uService := services.NewUserService(uRepo)

	// Loto
	lotoRepo := repository.NewLotoRepo(db)
	lotoService := services.NewLotoService(lotoRepo)

	f := fsm.New(fsm2.StateDefault, nil)

	// HANDLER
	handler := handlers.NewHandler(uService, f, rdb, lotoService)

	// Init and start bot
	newBot := bot.NewBot(cfg, ctx, handler)
	b, err := newBot.Build()
	if err != nil {
		panic(err)
	}

	w := worker.NewWorker(b, rdb, cfg.PPKey, lotoService)

	go w.Run(ctx)

	newBot.Start(b)
}
