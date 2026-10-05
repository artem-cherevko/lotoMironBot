package handlers

import (
	"context"
	"strings"

	fsm2 "lotoMironBot/internal/bot/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) Default(ctx context.Context, b *bot.Bot, update *models.Update) {
	var userID int64

	switch {
	case update.Message != nil:
		userID = update.Message.From.ID

	case update.CallbackQuery != nil:
		userID = update.CallbackQuery.From.ID

	default:
		return
	}

	state := h.f.Current(userID)
	if update.Message != nil && state == fsm2.StateOwnerRules {
		owner, err := h.uService.IsOwner(ctx, userID)
		if err != nil || !owner {
			h.f.Transition(userID, fsm2.StateDefault)
			sendText(ctx, b, update.Message.Chat.ID, "Недостаточно прав.")
			return
		}
		if update.Message.Chat.Type != models.ChatTypePrivate {
			sendText(ctx, b, update.Message.Chat.ID, "Отправьте новый текст правил в личном чате с ботом.")
			return
		}
		if err := h.lotoService.SetRules(ctx, update.Message.Text); err != nil {
			sendText(ctx, b, update.Message.Chat.ID, "Не удалось сохранить правила. Отправьте текст ещё раз.")
			return
		}
		h.f.Transition(userID, fsm2.StateDefault)
		sendText(ctx, b, update.Message.Chat.ID, "📖 Правила сохранены и уже доступны игрокам.")
		return
	}
	if update.Message != nil && update.Message.Chat.Type == models.ChatTypePrivate {
		switch strings.TrimSpace(update.Message.Text) {
		case "🎟 Мои билеты":
			h.PrivateMyTickets(ctx, b, update)
			return
		case "📖 Правила игры":
			h.ShowRules(ctx, b, update.Message.Chat.ID)
			return
		}
	}

	switch state {
	case fsm2.StateSelectCollection:
		h.selectCollection(ctx, b, update)

	case fsm2.StateProvidePhotos:
		if update.Message.Text == "Финиш загрузки фото" {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Точно?",
				ReplyMarkup: &models.ReplyKeyboardMarkup{
					ResizeKeyboard: true,
					ForceReply:     true,
					Keyboard: [][]models.KeyboardButton{
						{
							{Text: "Да", Style: "success"},
							{Text: "Нет", Style: "danger"},
						},
					},
				},
			})
			h.f.Transition(update.Message.From.ID, fsm2.StateFinish)
			return
		}

		if update.Message.Photo == nil {
			return
		}

		h.providePhotos(ctx, b, update)
	case fsm2.StateFinish:
		switch update.Message.Text {
		case "Да":
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Загрузка завершена ✨",
				ReplyMarkup: &models.ReplyKeyboardRemove{
					RemoveKeyboard: true,
				},
			})
			h.f.Transition(update.Message.From.ID, fsm2.StateDefault)
		case "Нет":
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Хорошо, отправляйте фото",
				ReplyMarkup: FinishCreateCollectionKb(),
			})
			h.f.Transition(update.Message.From.ID, fsm2.StateProvidePhotos)
		}
	case fsm2.StateDefault:
		return
	}
}
