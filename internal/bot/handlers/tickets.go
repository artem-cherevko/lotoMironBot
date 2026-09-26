package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) GetTickets(ctx context.Context, b *bot.Bot, update *models.Update) {
	tickets, err := h.lotoService.GetAllTickets(ctx)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Не удалось получить список билетов :(",
		})
		return
	}

	if len(tickets) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "🎫 Билетов пока нет.",
		})
		return
	}

	var text strings.Builder

	for _, t := range tickets {
		numbers := make([]string, 0, len(t.Numbers))

		for _, n := range t.Numbers {
			numbers = append(numbers, fmt.Sprintf("%02d", n))
		}

		fmt.Fprintf(
			&text,
			"🎫 <b>Билет №%d</b>\n"+
				"🔢 %s\n"+
				"🔖 Коллекция: %s\n\n",
			t.ID,
			strings.Join(numbers, " • "),
			t.Collection,
		)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text.String(),
		ParseMode: models.ParseModeHTML,
	})
}
