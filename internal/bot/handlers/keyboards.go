package handlers

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

func CollectionsKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Стандартная коллекция", Style: "primary", CallbackData: "standard-collection"},
			},
			{
				{Text: "Новогодняя коллекция", Style: "primary", CallbackData: "new-year-collection", IconCustomEmojiID: "5449857802593901902"},
			},
			{
				{Text: "Хэллоуинская коллекция", Style: "primary", CallbackData: "halloween-collection", IconCustomEmojiID: "5370610867094166617"},
			},
		},
	}
}

func FinishCreateCollectionKb() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: " Финиш загрузки фото", IconCustomEmojiID: "5411520005386806155"},
			},
		},
	}
}

func DrawBarrelKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{
			{Text: "🎱 Вытянуть бочонок", Style: "primary", CallbackData: "game:draw"},
		}},
	}
}

func ClaimNumberKeyboard(gameID uint, number int32, drawIndex int) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: fmt.Sprintf("У меня есть %d", number), Style: "success", CallbackData: fmt.Sprintf("game:claim:%d:%d:%d", gameID, number, drawIndex)}},
		},
	}
}

func NextBarrelKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{
		{Text: "Следующий бочонок", Style: "primary", CallbackData: "game:draw"},
	}}}
}

func PrivateMenuKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{ResizeKeyboard: true, Keyboard: [][]models.KeyboardButton{{{Text: "🎟 Мои билеты"}}, {{Text: "📖 Правила игры"}}}}
}

func AdminPanelKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{{Text: "🎮 Создать игру", CallbackData: "admin:create_game"}},
		{{Text: "🎮 Текущая игра", CallbackData: "admin:current"}},
		{{Text: "👥 Игроки", CallbackData: "admin:players"}},
		{{Text: "🏆 Победители", CallbackData: "admin:winners"}},
	}}
}

func OwnerPanelKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{{Text: "👥 Администраторы", CallbackData: "owner:admins"}},
		{{Text: "📖 Правила игры", CallbackData: "owner:rules"}},
		{{Text: "🎟 Коллекция для игр", CallbackData: "owner:collection"}},
		{{Text: "🎮 Управление играми", CallbackData: "admin:current"}},
	}}
}

func DefaultCollectionKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{{Text: "Стандартная", CallbackData: "owner:set_collection:standard"}},
		{{Text: "Новогодняя", CallbackData: "owner:set_collection:new-year"}},
		{{Text: "Хэллоуинская", CallbackData: "owner:set_collection:halloween"}},
	}}
}

func RulesKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{{Text: "✏️ Изменить правила", CallbackData: "owner:edit_rules"}}}}
}
