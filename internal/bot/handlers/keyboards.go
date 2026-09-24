package handlers

import "github.com/go-telegram/bot/models"

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
