package main

import (
	"context"
	"lotoMironBot/internal/bot"
	"lotoMironBot/internal/config"
	"lotoMironBot/internal/database"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Connect to db
	_, err = database.Connect(cfg.DBDsn)
	if err != nil {
		panic(err)
	}

	// Init and start bot
	newBot := bot.NewBot(cfg, ctx)
	newBot.Start()
}
