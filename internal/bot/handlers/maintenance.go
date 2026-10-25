package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) Maintenance(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}

	chatID := update.Message.Chat.ID
	adminID, err := strconv.ParseInt(strings.TrimSpace(h.maintenanceAdminID), 10, 64)
	if err != nil || update.Message.From.ID != adminID {
		sendText(ctx, b, chatID, "Недостаточно прав.")
		return
	}

	args := strings.Fields(update.Message.Text)
	if len(args) != 2 {
		sendText(ctx, b, chatID, "Использование: /maintenance [off/yellow/red]")
		return
	}

	modeByArgument := map[string]string{
		"off":    "normal",
		"yellow": "degraded",
		"red":    "maintenance",
	}
	mode, ok := modeByArgument[strings.ToLower(args[1])]
	if !ok {
		sendText(ctx, b, chatID, "Допустимые значения: off, yellow, red.")
		return
	}

	if err := saveMaintenanceConfig("/app/maintenance.json", mode); err != nil {
		log.Printf("can't save maintenance mode: %v", err)
		sendText(ctx, b, chatID, fmt.Sprintf("Не удалось сохранить режим обслуживания: %v", err))
		return
	}

	sendText(ctx, b, chatID, "Режим обслуживания изменён: "+mode+".")
}

func saveMaintenanceConfig(path, mode string) error {
	data, err := json.MarshalIndent(struct {
		Mode string `json:"mode"`
	}{Mode: mode}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return fmt.Errorf("open maintenance config: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("write maintenance config: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close maintenance config: %w", err)
	}
	return nil
}
