package handlers

import (
	"context"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.Chat.Type != models.ChatTypePrivate {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: "Чтобы получать билеты и результаты, сначала откройте личный чат с ботом и отправьте /start."})
		return
	}
	ownerID, _ := strconv.ParseInt(h.adminID, 10, 64)
	if err := h.uService.Register(ctx, *update.Message.From, ownerID); err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: "Не удалось сохранить профиль. Попробуйте позже."})
		return
	}
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
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
}
