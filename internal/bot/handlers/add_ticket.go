package handlers

import (
	"context"
	"log"
	"lotoMironBot/internal/bot/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) AddTicket(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.f.Transition(update.Message.From.ID, fsm.StateSelectCollection)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выбери коллекцию билетов:",
		ReplyMarkup: CollectionsKeyboard(),
	})

	if err != nil {
		log.Println(err)
		return
	}
}
