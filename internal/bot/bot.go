package bot

import (
	"context"
	"log"
	"lotoMironBot/internal/bot/handlers"
	"lotoMironBot/internal/config"

	bot2 "github.com/go-telegram/bot"
)

type Bot struct {
	cfg     *config.Config
	ctx     context.Context
	handler *handlers.Handler
}

func NewBot(cfg *config.Config, ctx context.Context, handler *handlers.Handler) *Bot {
	return &Bot{cfg: cfg, ctx: ctx, handler: handler}
}

func (b *Bot) Start() {
	opts := []bot2.Option{
		bot2.WithDefaultHandler(b.handler.Default),
	}
	bot, err := bot2.New(b.cfg.BotToken, opts...)
	if err != nil {
		log.Fatal(err)
	}

	bot.DeleteWebhook(b.ctx, &bot2.DeleteWebhookParams{DropPendingUpdates: true})

	bot.RegisterHandler(bot2.HandlerTypeMessageText, "/start", bot2.MatchTypeExact, b.handler.Start)
	bot.RegisterHandler(bot2.HandlerTypeMessageText, "/add_ticket", bot2.MatchTypePrefix, b.handler.AddTicket)

	log.Println("Bot started")

	bot.Start(b.ctx)
}
