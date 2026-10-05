package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	fsm2 "lotoMironBot/internal/bot/fsm"
	"lotoMironBot/internal/database"
)

func (h *Handler) AdminPanel(ctx context.Context, b *bot.Bot, u *models.Update) {
	if !h.isAdmin(ctx, u.Message.From.ID) {
		sendText(ctx, b, u.Message.Chat.ID, "Эта панель доступна только администратору.")
		return
	}
	sendTextWithKeyboard(ctx, b, u.Message.Chat.ID, "🛠 ADMIN PANEL", AdminPanelKeyboard())
}

func (h *Handler) OwnerPanel(ctx context.Context, b *bot.Bot, u *models.Update) {
	owner, err := h.uService.IsOwner(ctx, u.Message.From.ID)
	if err != nil || !owner {
		sendText(ctx, b, u.Message.Chat.ID, "Эта панель доступна только владельцу бота.")
		return
	}
	sendTextWithKeyboard(ctx, b, u.Message.Chat.ID, "⚙️ OWNER PANEL", OwnerPanelKeyboard())
}

func (h *Handler) ShowRules(ctx context.Context, b *bot.Bot, chatID int64) {
	rules, err := h.lotoService.GetRules(ctx)
	if err != nil {
		sendText(ctx, b, chatID, "Не удалось получить правила.")
		return
	}
	sendText(ctx, b, chatID, "📖 ТЕКУЩИЕ ПРАВИЛА\n\n"+rules)
}

func (h *Handler) PrivateMyTickets(ctx context.Context, b *bot.Bot, u *models.Update) {
	tickets, err := h.lotoService.GetPlayerTicketsPrivate(ctx, u.Message.From.ID)
	if err != nil || len(tickets) == 0 {
		sendText(ctx, b, u.Message.Chat.ID, "🎟 Сейчас у вас нет билетов в активной игре.")
		return
	}
	for _, ticket := range tickets {
		_, _ = b.SendPhoto(ctx, &bot.SendPhotoParams{ChatID: u.Message.Chat.ID, Photo: &models.InputFileString{Data: ticket.FileID}})
	}
}

func (h *Handler) AdminCallback(ctx context.Context, b *bot.Bot, u *models.Update) {
	q := u.CallbackQuery
	isOwner, err := h.uService.IsOwner(ctx, q.From.ID)
	admin, adminErr := h.uService.IsAdmin(ctx, q.From.ID)
	if (err != nil || !isOwner) && (adminErr != nil || !admin) {
		answerCallback(ctx, b, q.ID, "Недостаточно прав.", true)
		return
	}
	if q.Message.Message == nil {
		answerCallback(ctx, b, q.ID, "Команда недоступна в этом сообщении.", true)
		return
	}
	chat := q.Message.Message.Chat
	switch q.Data {
	case "admin:create_game":
		if !isGroupChat(chat) {
			answerCallback(ctx, b, q.ID, "Создать игру можно в группе.", true)
			return
		}
		collection, e := h.lotoService.GetDefaultCollection(ctx)
		if e != nil {
			answerCallback(ctx, b, q.ID, "Не удалось получить стандартную коллекцию игры.", true)
			return
		}
		if _, e := h.lotoService.CreateGame(ctx, q.From.ID, chat.ID, 0, collection); e != nil {
			answerCallback(ctx, b, q.ID, gameErrorText(e), true)
			return
		}
		answerCallback(ctx, b, q.ID, "🎮 Игра создана.", false)
		sendText(ctx, b, chat.ID, gameCreatedText())
	case "admin:current", "admin:players", "admin:winners":
		if !isGroupChat(chat) {
			answerCallback(ctx, b, q.ID, "Откройте панель в нужной группе.", true)
			return
		}
		stats, e := h.lotoService.GetPlayerGameStats(ctx, chat.ID)
		if e != nil {
			answerCallback(ctx, b, q.ID, gameErrorText(e), true)
			return
		}
		users := make(map[int64]string)
		for _, s := range stats {
			user, e := h.uService.Get(ctx, s.PlayerID)
			if e == nil {
				name := strings.TrimSpace(user.FirstName + " " + user.LastName)
				if user.Username != "" {
					name = "@" + user.Username
				}
				users[s.PlayerID] = name
			}
		}
		var out strings.Builder
		if q.Data == "admin:winners" {
			out.WriteString("🏆 ПОБЕДИТЕЛИ\n\n")
		} else {
			out.WriteString("👥 ИГРОКИ\n\n")
		}
		printed := 0
		for _, s := range stats {
			if q.Data == "admin:winners" && !s.Winner {
				continue
			}
			name := users[s.PlayerID]
			if name == "" {
				name = fmt.Sprintf("ID %d", s.PlayerID)
			}
			if q.Data == "admin:winners" {
				printed++
				fmt.Fprintf(&out, "%d. %s — %d билетов\n", printed, name, s.Closed)
				continue
			}
			status := "В игре"
			if s.Disqualified {
				status = "Дисквалифицирован"
			}
			fmt.Fprintf(&out, "👤 %s (ID %d)\n🎟 Билетов: %d\n✅ Закрыто: %d\n⏳ Осталось: %d\n❌ Ошибок: %d\nСтатус: %s", name, s.PlayerID, s.Tickets, s.Closed, s.Tickets-s.Closed, s.Failures, status)
			if s.Winner {
				out.WriteString("\n🏆 ПОБЕДИТЕЛЬ")
			}
			out.WriteString("\n\n")
		}
		if printed == 0 && q.Data == "admin:winners" {
			out.WriteString("Победителей пока нет.")
		}
		answerCallback(ctx, b, q.ID, "Готово.", false)
		sendText(ctx, b, chat.ID, out.String())
	default:
		answerCallback(ctx, b, q.ID, "Неизвестное действие.", true)
	}
}

func (h *Handler) OwnerCallback(ctx context.Context, b *bot.Bot, u *models.Update) {
	q := u.CallbackQuery
	owner, err := h.uService.IsOwner(ctx, q.From.ID)
	if err != nil || !owner {
		answerCallback(ctx, b, q.ID, "Только owner может выполнять это действие.", true)
		return
	}
	switch q.Data {
	case "owner:collection":
		collection, err := h.lotoService.GetDefaultCollection(ctx)
		if err != nil {
			answerCallback(ctx, b, q.ID, "Не удалось получить текущую коллекцию.", true)
			return
		}
		answerCallback(ctx, b, q.ID, "Выберите коллекцию для новых игр.", false)
		if q.Message.Message != nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: q.Message.Message.Chat.ID, Text: fmt.Sprintf("Текущая стандартная коллекция: %s", collection), ReplyMarkup: DefaultCollectionKeyboard()})
		}
	case "owner:set_collection:standard", "owner:set_collection:new-year", "owner:set_collection:halloween":
		collection := q.Data[len("owner:set_collection:"):]
		if err := h.lotoService.SetDefaultCollection(ctx, database.Collections(collection)); err != nil {
			answerCallback(ctx, b, q.ID, "Не удалось сохранить коллекцию.", true)
			return
		}
		answerCallback(ctx, b, q.ID, "Коллекция для новых игр обновлена.", false)
		if q.Message.Message != nil {
			sendText(ctx, b, q.Message.Message.Chat.ID, fmt.Sprintf("✅ Стандартная коллекция для новых игр: %s", collection))
		}
	case "owner:admins":
		admins, err := h.uService.ListAdmins(ctx)
		if err != nil {
			answerCallback(ctx, b, q.ID, "Не удалось получить список.", true)
			return
		}
		var out strings.Builder
		out.WriteString("👥 АДМИНИСТРАТОРЫ\n\n")
		if len(admins) == 0 {
			out.WriteString("Администраторов пока нет.\n")
		}
		for _, a := range admins {
			name := a.Username
			if name != "" {
				name = "@" + name
			} else {
				name = strings.TrimSpace(a.FirstName + " " + a.LastName)
			}
			fmt.Fprintf(&out, "• %s — %d\n", name, a.TelegramID)
		}
		out.WriteString("\nДобавить: /addadmin <telegram_id>\nУдалить: /removeadmin <telegram_id>")
		answerCallback(ctx, b, q.ID, "Список администраторов.", false)
		if q.Message.Message != nil {
			sendText(ctx, b, q.Message.Message.Chat.ID, out.String())
		}
	case "owner:rules":
		rules, err := h.lotoService.GetRules(ctx)
		if err != nil {
			answerCallback(ctx, b, q.ID, "Не удалось получить правила.", true)
			return
		}
		answerCallback(ctx, b, q.ID, "Текущие правила.", false)
		if q.Message.Message != nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: q.Message.Message.Chat.ID, Text: "📖 ТЕКУЩИЕ ПРАВИЛА\n\n" + rules, ReplyMarkup: RulesKeyboard()})
		}
	case "owner:edit_rules":
		if q.Message.Message == nil || q.Message.Message.Chat.Type != models.ChatTypePrivate {
			answerCallback(ctx, b, q.ID, "Откройте /owner в личном чате с ботом, чтобы изменить правила.", true)
			return
		}
		h.f.Transition(q.From.ID, fsm2.StateOwnerRules)
		answerCallback(ctx, b, q.ID, "Отправьте новый текст правил в личном чате.", false)
		if q.Message.Message != nil {
			sendText(ctx, b, q.Message.Message.Chat.ID, "✏️ Отправьте новый текст правил одним сообщением.")
		}
	default:
		answerCallback(ctx, b, q.ID, "Неизвестное действие.", true)
	}
}

func (h *Handler) AddAdmin(ctx context.Context, b *bot.Bot, u *models.Update) {
	h.changeAdmin(ctx, b, u, true)
}
func (h *Handler) RemoveAdmin(ctx context.Context, b *bot.Bot, u *models.Update) {
	h.changeAdmin(ctx, b, u, false)
}

func (h *Handler) changeAdmin(ctx context.Context, b *bot.Bot, u *models.Update, add bool) {
	owner, err := h.uService.IsOwner(ctx, u.Message.From.ID)
	if err != nil || !owner {
		sendText(ctx, b, u.Message.Chat.ID, "Только владелец бота может управлять администраторами.")
		return
	}
	parts := strings.Fields(u.Message.Text)
	var id int64
	if u.Message.ReplyToMessage != nil && u.Message.ReplyToMessage.From != nil && len(parts) == 1 {
		id = u.Message.ReplyToMessage.From.ID
	} else if len(parts) == 2 {
		id, err = strconv.ParseInt(parts[1], 10, 64)
	}
	if id == 0 {
		sendText(ctx, b, u.Message.Chat.ID, "Использование: ответьте на сообщение пользователя командой или укажите /addadmin <telegram_id>.")
		return
	}
	user, err := h.uService.Get(ctx, id)
	if err != nil {
		sendText(ctx, b, u.Message.Chat.ID, "Пользователь не зарегистрирован. Он должен сначала отправить боту /start.")
		return
	}
	if user.Role == "owner" {
		sendText(ctx, b, u.Message.Chat.ID, "Нельзя изменить роль владельца бота.")
		return
	}
	if add && user.Role == "admin" {
		sendText(ctx, b, u.Message.Chat.ID, "Этот пользователь уже администратор.")
		return
	}
	if !add && user.Role != "admin" {
		sendText(ctx, b, u.Message.Chat.ID, "Этот пользователь не является администратором.")
		return
	}
	role := user.Role
	if add {
		role = "admin"
	} else {
		role = "player"
	}
	if err := h.uService.SetRole(ctx, id, role); err != nil {
		sendText(ctx, b, u.Message.Chat.ID, "Не удалось обновить роль.")
		return
	}
	if add {
		sendText(ctx, b, u.Message.Chat.ID, "✅ Администратор добавлен.")
	} else {
		sendText(ctx, b, u.Message.Chat.ID, "✅ Администратор удалён.")
	}
}
