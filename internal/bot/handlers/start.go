package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `🎱 <b>Русское лото</b>

Добро пожаловать в игру!

🎟 Получи свой билет у администратора
🎲 Дождись начала розыгрыша
🪵 Следи за выпадающими бочонками
🏆 Закрой все 6 чисел и забери победу!

Твои билеты будут сохранены здесь, поэтому ты всегда сможешь посмотреть их перед игрой.

🍀 <i>Удачи! Пусть именно твой билет окажется счастливым.</i>`,
		ParseMode: "HTML",
	})
	if err != nil {
		return
	}
}
