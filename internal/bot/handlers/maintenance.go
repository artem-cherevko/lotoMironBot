package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	tmp, err := os.CreateTemp(filepath.Dir(path), ".maintenance-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
