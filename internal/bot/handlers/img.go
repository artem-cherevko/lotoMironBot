package handlers

import (
	"context"
	"log"

	fsm2 "lotoMironBot/internal/bot/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// ImgCommand enters a state where the bot waits for an image and then
// replies with the image file ID.
func (h *Handler) Img(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if update.Message.Chat.Type != models.ChatTypePrivate {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Команда доступна только в личном чате с ботом.",
		})
		return
	}

	h.f.Transition(update.Message.From.ID, fsm2.StateImageID)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Отправьте фото, и я пришлю вам его file ID.",
	})
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) provideImageID(ctx context.Context, b *bot.Bot, update *models.Update) error {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if update.Message.Photo == nil {
		return nil
	}

	photoID := update.Message.Photo[len(update.Message.Photo)-1].FileID

	h.f.Transition(userID, fsm2.StateDefault)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "File ID изображения:\n<code>" + photoID + "</code>",
		ParseMode: models.ParseModeHTML,
	})
	return err
}
