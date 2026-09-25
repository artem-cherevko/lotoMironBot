package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	fsm2 "lotoMironBot/internal/bot/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) selectCollection(ctx context.Context, b *bot.Bot, update *models.Update) error {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Хорошо, отправьте теперь фото билетов для данной коллекции",
		ReplyMarkup: FinishCreateCollectionKb(),
	})

	h.f.Set(update.CallbackQuery.From.ID, "collection", update.CallbackQuery.Data)

	h.f.Transition(chatID, fsm2.StateProvidePhotos, b, chatID)

	return err
}

func (h *Handler) providePhotos(ctx context.Context, b *bot.Bot, update *models.Update) error {
	chatID := update.Message.Chat.ID
	photoID := update.Message.Photo[len(update.Message.Photo)-1].FileID

	collection, exist := h.f.Get(update.Message.From.ID, "collection")
	if !exist {
		return errors.New("Не возможно получить наименования коллекции")
	}

	key := fmt.Sprintf("ticket:%s:%d", collection, chatID)

	photos, err := h.r.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return err
	}

	if slices.Contains(photos, photoID) {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "<b>Обнаружен(и) дубликат <tg-emoji emoji-id=\"5395695537687123235\">🚨</tg-emoji></b>\nСохраняем все без дублирования.",
			ParseMode: models.ParseModeHTML,
		})
		if err != nil {
			log.Println(err)
		}
		return nil
	}

	err = h.r.RPush(ctx, key, photoID).Err()
	h.r.Expire(ctx, key, 30*time.Minute)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}
