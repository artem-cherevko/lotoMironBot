package bot

import (
	"context"
	"log"
	"lotoMironBot/internal/config"

	bot2 "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Bot struct {
	cfg *config.Config
	ctx context.Context
}

func NewBot(cfg *config.Config, ctx context.Context) *Bot {
	return &Bot{cfg: cfg, ctx: ctx}
}

func (b *Bot) Start() {
	bot, err := bot2.New(b.cfg.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.DeleteWebhook(b.ctx, &bot2.DeleteWebhookParams{DropPendingUpdates: true})

	bot.RegisterHandler(bot2.HandlerTypeMessageText, "/start", bot2.MatchTypeExact, func(ctx context.Context, b *bot2.Bot, update *models.Update) {
		_, err := b.SendMessage(ctx, &bot2.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Hello",
		})
		if err != nil {
			return
		}
	})

	log.Println("Bot started")

	bot.Start(b.ctx)
}
