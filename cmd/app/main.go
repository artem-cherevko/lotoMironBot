package main

import (
	"context"
	"lotoMironBot/internal/bot"
	"lotoMironBot/internal/config"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	newBot := bot.NewBot(cfg, ctx)
	newBot.Start()
}
