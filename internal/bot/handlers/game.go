package handlers

import (
	"context"
	"errors"
	"fmt"
	"lotoMironBot/internal/database"
	"lotoMironBot/internal/services"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) CreateGame(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !isGroupChat(update.Message.Chat) {
		sendGroupOnly(ctx, b, update.Message.Chat.ID)
		return
	}
	if !h.isAdmin(ctx, update.Message.From.ID) {
		sendNotAdmin(ctx, b, update.Message.Chat.ID)
		return
	}
	collection, err := h.lotoService.GetDefaultCollection(ctx)
	if err != nil {
		sendText(ctx, b, update.Message.Chat.ID, "Не удалось получить стандартную коллекцию игры.")
		return
	}
	_, err = h.lotoService.CreateGame(ctx, update.Message.From.ID, update.Message.Chat.ID, 0, collection)
	if err != nil {
		sendText(ctx, b, update.Message.Chat.ID, gameErrorText(err))
		return
	}
	sendText(ctx, b, update.Message.Chat.ID, gameCreatedText())
}

func gameCreatedText() string {
	return "🎰 ИГРА СОЗДАНА!\n\n🍀 Приобретайте свой счастливый билетик и испытайте удачу!\n\n🔥 Следите за бочонками и не упустите свой шанс на победу!"
}

func (h *Handler) GiveTickets(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.requireGroupAdmin(ctx, b, update) {
		return
	}
	if update.Message.ReplyToMessage == nil || update.Message.ReplyToMessage.From == nil {
		sendText(ctx, b, update.Message.Chat.ID, "Ответьте командой /gticket <1-5> на сообщение игрока.")
		return
	}
	parts := strings.Fields(update.Message.Text)
	if len(parts) != 2 {
		sendText(ctx, b, update.Message.Chat.ID, "Использование: ответьте на сообщение игрока командой /gticket <1-5>.")
		return
	}
	count, err := strconv.Atoi(parts[1])
	if err != nil || count < 1 || count > 5 {
		sendText(ctx, b, update.Message.Chat.ID, "Количество билетов должно быть от 1 до 5.")
		return
	}
	player := update.Message.ReplyToMessage.From
	if _, err := h.uService.Get(ctx, player.ID); err != nil {
		sendText(ctx, b, update.Message.Chat.ID, "Игрок должен сначала открыть личный чат с ботом и отправить /start, затем повторите /gticket.")
		return
	}
	assignments, err := h.lotoService.GiveTickets(ctx, update.Message.Chat.ID, player.ID, count)
	if err != nil {
		sendText(ctx, b, update.Message.Chat.ID, gameErrorText(err))
		return
	}
	sent := 0
	for _, assignment := range assignments {
		_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: assignment.PlayerID,
			Photo:  &models.InputFileString{Data: assignment.Ticket.FileID},
		})
		if err == nil {
			sent++
		}
	}
	if sent != len(assignments) {
		sendText(ctx, b, update.Message.Chat.ID, fmt.Sprintf("Назначено билетов: %d из %d отправлены в личные сообщения. Повторите /gticket %d ответом на это сообщение игрока для повторной отправки.", len(assignments), sent, count))
		return
	}
	sendText(ctx, b, update.Message.Chat.ID, fmt.Sprintf("Игроку выдано билетов: %d.", len(assignments)))
}

func (h *Handler) MyTickets(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !isGroupChat(update.Message.Chat) {
		sendGroupOnly(ctx, b, update.Message.Chat.ID)
		return
	}
	tickets, err := h.lotoService.GetPlayerTickets(ctx, update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		sendText(ctx, b, update.Message.Chat.ID, gameErrorText(err))
		return
	}
	if len(tickets) == 0 {
		sendText(ctx, b, update.Message.Chat.ID, "У вас нет билетов в этой игре.")
		return
	}
	for _, ticket := range tickets {
		_, _ = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:          update.Message.Chat.ID,
			Photo:           &models.InputFileString{Data: ticket.FileID},
			ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID, AllowSendingWithoutReply: true},
		})
	}
}

func (h *Handler) StartGame(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.requireGroupAdmin(ctx, b, update) {
		return
	}
	if err := h.lotoService.StartGame(ctx, update.Message.Chat.ID); err != nil {
		sendText(ctx, b, update.Message.Chat.ID, gameErrorText(err))
		return
	}
	sendTextWithKeyboard(ctx, b, update.Message.Chat.ID, "🔥 ВСЁ, ПОЕХАЛИ! 🔥\n\n🎟 Билетики куплены.\n🎱 Бочонки готовы.\n🍀 Удача уже выбирает своего победителя…\n\n👀 Не отвлекайтесь ни на секунду — ваше число может выпасть прямо сейчас!\n\n🎰 Игра началась.\nПогнали за победой! 🏆🔥", DrawBarrelKeyboard())
}

func (h *Handler) EndGame(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.requireGroupAdmin(ctx, b, update) {
		return
	}
	if err := h.lotoService.EndGame(ctx, update.Message.Chat.ID); err != nil {
		sendText(ctx, b, update.Message.Chat.ID, gameErrorText(err))
		return
	}
	sendText(ctx, b, update.Message.Chat.ID, "Игра завершена администратором.")
}

func (h *Handler) DrawBarrel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
		return
	}
	message := update.CallbackQuery.Message.Message
	chat := message.Chat
	query := update.CallbackQuery
	if !isGroupChat(chat) {
		answerCallback(ctx, b, query.ID, "Игра доступна только в группах.", true)
		return
	}
	if !h.isAdmin(ctx, query.From.ID) {
		answerCallback(ctx, b, query.ID, "Только администратор может вытянуть бочонок.", true)
		return
	}
	result, err := h.lotoService.Draw(ctx, chat.ID)
	if err != nil {
		if errors.Is(err, services.ErrClaimWindowOpen) {
			answerCallback(ctx, b, query.ID, "Следующий бочонок пока недоступен.", true)
			return
		}
		answerCallback(ctx, b, query.ID, gameErrorText(err), true)
		return
	}
	if result.Finished {
		_, _ = b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{ChatID: chat.ID, MessageID: message.ID})
		sendText(ctx, b, chat.ID, "Игра завершена: все 100 бочонков вытянуты, победителя нет.")
		answerCallback(ctx, b, query.ID, "Игра завершена.", false)
		return
	}
	_, _ = b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{ChatID: chat.ID, MessageID: message.ID})
	text := fmt.Sprintf("🎱 Выпало число: %d\n%s", result.Number, barrelName(result.Number))
	newMessage, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chat.ID, Text: text,
		ReplyMarkup: ClaimNumberKeyboard(result.GameID, result.Number, result.DrawIndex),
	})
	if sendErr != nil {
		answerCallback(ctx, b, query.ID, "Не удалось отправить результат розыгрыша.", true)
		return
	}
	if newMessage != nil {
		go h.closeClaimWindow(chat.ID, newMessage.ID, result.GameID, result.DrawIndex, b)
	}
	answerCallback(ctx, b, query.ID, "Бочонок вытянут.", false)
}

func (h *Handler) closeClaimWindow(chatID int64, messageID int, gameID uint, drawIndex int, b *bot.Bot) {
	time.Sleep(15 * time.Second)
	ctx := context.Background()
	if !h.lotoService.IsCurrentDraw(ctx, chatID, gameID, drawIndex) {
		return
	}
	_, _ = b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID: chatID, MessageID: messageID,
		ReplyMarkup: NextBarrelKeyboard(),
	})
}

func (h *Handler) ClaimNumber(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query.Message.Message == nil {
		return
	}
	chat := query.Message.Message.Chat
	if !isGroupChat(chat) {
		answerCallback(ctx, b, query.ID, "Игра доступна только в группах.", true)
		return
	}
	parts := strings.Split(query.Data, ":")
	if len(parts) != 5 {
		answerCallback(ctx, b, query.ID, "Кнопка игры устарела.", true)
		return
	}
	gameID64, gameErr := strconv.ParseUint(parts[2], 10, 32)
	number64, numberErr := strconv.ParseInt(parts[3], 10, 32)
	drawIndex, indexErr := strconv.Atoi(parts[4])
	if gameErr != nil || numberErr != nil || indexErr != nil {
		answerCallback(ctx, b, query.ID, "Кнопка игры устарела.", true)
		return
	}
	result, err := h.lotoService.ClaimDrawNumber(ctx, chat.ID, uint(gameID64), query.From.ID, int32(number64), drawIndex)
	if err != nil {
		answerCallback(ctx, b, query.ID, claimErrorText(err), true)
		return
	}
	playerName := query.From.FirstName
	if query.From.Username != "" {
		playerName = "@" + query.From.Username
	}
	dmText := fmt.Sprintf("❌ Число %d отсутствует в ваших активных билетах. Ошибок: %d из 9.", number64, result.Failures)
	if result.Success {
		dmText = fmt.Sprintf("✅ Успешно! Число %d отмечено в ваших билетах.", number64)
	}
	if result.Disqualified {
		dmText = "Вы дисквалифицированы за девять неверных нажатий."
	}
	if _, dmErr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: query.From.ID, Text: dmText}); dmErr != nil {
		answerCallback(ctx, b, query.ID, dmText+" Если это личное сообщение не пришло, отправьте боту /start.", true)
	} else {
		answerCallback(ctx, b, query.ID, "Результат отправлен вам в личные сообщения.", false)
	}

	if result.GameFinished {
		_, _ = b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{ChatID: chat.ID, MessageID: query.Message.Message.ID})
		finalText := fmt.Sprintf("🚫 %s дисквалифицирован(а). Игра завершена: других игроков не осталось.", playerName)
		if result.Winner {
			finalText = fmt.Sprintf("🏆 %s победил(а)!\n✅ Закрыто билетов: %d.\n🎉 Игра завершена.", playerName, result.ClosedTickets)
		}
		sendText(ctx, b, chat.ID, finalText)
		return
	}
	if result.Disqualified {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat.ID, Text: fmt.Sprintf("🚫 %s дисквалифицирован(а) за девять неверных нажатий. Игра продолжается.", playerName)})
	}
}

func (h *Handler) requireGroupAdmin(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	if !isGroupChat(update.Message.Chat) {
		sendGroupOnly(ctx, b, update.Message.Chat.ID)
		return false
	}
	if !h.isAdmin(ctx, update.Message.From.ID) {
		sendNotAdmin(ctx, b, update.Message.Chat.ID)
		return false
	}
	return true
}

func (h *Handler) isAdmin(ctx context.Context, userID int64) bool {
	admin, err := h.uService.IsAdmin(ctx, userID)
	return err == nil && admin
}

func isGroupChat(chat models.Chat) bool {
	return chat.Type == models.ChatTypeGroup || chat.Type == models.ChatTypeSupergroup
}

func sendGroupOnly(ctx context.Context, b *bot.Bot, chatID int64) {
	sendText(ctx, b, chatID, "Игра доступна только в Telegram-группах.")
}

func sendNotAdmin(ctx context.Context, b *bot.Bot, chatID int64) {
	sendText(ctx, b, chatID, "Эта команда доступна только администратору игры.")
}

func sendText(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text})
}

func answerCallback(ctx context.Context, b *bot.Bot, callbackID, text string, alert bool) {
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: callbackID, Text: text, ShowAlert: alert})
}

func sendTextWithKeyboard(ctx context.Context, b *bot.Bot, chatID int64, text string, keyboard models.ReplyMarkup) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text, ReplyMarkup: keyboard})
}

func gameErrorText(err error) string {
	switch {
	case errors.Is(err, services.ErrActiveGameExists):
		return "У администратора уже есть активная игра."
	case errors.Is(err, services.ErrChatGameExists):
		return "В этой группе уже есть активная игра."
	case errors.Is(err, services.ErrGameNotFound):
		return "В этой группе нет активной игры."
	case errors.Is(err, services.ErrRecruitmentClosed):
		return "Регистрация уже закрыта."
	case errors.Is(err, services.ErrParticipantLimit):
		return "Лимит участников уже достигнут."
	case errors.Is(err, services.ErrInvalidGiveCount):
		return "Количество билетов должно быть от 1 до 5."
	case errors.Is(err, services.ErrExistingTicketCount):
		return "У игрока уже назначено больше билетов, чем указано."
	case errors.Is(err, services.ErrGameNotReady):
		return "Игра ещё не готова: выдайте игрокам билеты командой /gticket и повторите /startgame."
	case errors.Is(err, services.ErrNotGameParticipant):
		return "У вас нет билетов в этой игре."
	case errors.Is(err, services.ErrGameNotRunning):
		return "Игра ещё не запущена."
	case errors.Is(err, services.ErrDeckExhausted):
		return "Бочонки закончились."
	case errors.Is(err, services.ErrClaimWindowOpen):
		return "Следующий бочонок пока недоступен."
	default:
		return "Не удалось выполнить игровую операцию."
	}
}

func claimErrorText(err error) string {
	switch {
	case errors.Is(err, services.ErrClaimWindowExpired):
		return "Время ответа на этот бочонок истекло."
	case errors.Is(err, services.ErrNotGameParticipant):
		return "Вы не участвуете в этой игре или уже дисквалифицированы."
	case errors.Is(err, services.ErrAlreadyAttempted):
		return "Вы уже отвечали на этот бочонок."
	case errors.Is(err, services.ErrGameNotFound), errors.Is(err, services.ErrGameNotRunning):
		return "Игра уже завершена или кнопка устарела."
	default:
		return "Не удалось проверить ответ."
	}
}

func (h *Handler) PromoteAdmin(ctx context.Context, b *bot.Bot, update *models.Update) {
	owner, err := h.uService.IsOwner(ctx, update.Message.From.ID)
	if err != nil || !owner {
		sendText(ctx, b, update.Message.Chat.ID, "Назначать администраторов может только владелец бота.")
		return
	}
	parts := strings.Fields(update.Message.Text)
	if len(parts) != 2 {
		sendText(ctx, b, update.Message.Chat.ID, "Использование: /promote <telegram_user_id>. Пользователь должен сначала отправить боту /start.")
		return
	}
	telegramID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		sendText(ctx, b, update.Message.Chat.ID, "Telegram ID должен быть числом.")
		return
	}
	if _, err := h.uService.Get(ctx, telegramID); err != nil {
		sendText(ctx, b, update.Message.Chat.ID, "Пользователь не зарегистрирован. Сначала пусть отправит боту /start.")
		return
	}
	if err := h.uService.SetRole(ctx, telegramID, database.RoleAdmin); err != nil {
		sendText(ctx, b, update.Message.Chat.ID, "Не удалось назначить администратора.")
		return
	}
	sendText(ctx, b, update.Message.Chat.ID, "Пользователь назначен администратором.")
}
