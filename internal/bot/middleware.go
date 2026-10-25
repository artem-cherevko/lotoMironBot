package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type MaintenanceConfig struct {
	Mode string `json:"mode"`
}

const maintenanceConfigPath = "maintenance.json"

func loadMaintenanceConfig(path string) (*MaintenanceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read maintenance config: %w", err)
	}

	var cfg MaintenanceConfig

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse maintenance config: %w", err)
	}

	switch cfg.Mode {
	case "normal", "degraded", "maintenance":
	default:
		return nil, fmt.Errorf("invalid maintenance mode: %q", cfg.Mode)
	}

	return &cfg, nil
}

func (b *Bot) MaintenanceMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, bot2 *bot.Bot, update *models.Update) {
		if update.Message != nil {
			text := strings.TrimSpace(update.Message.Text)

			if strings.HasPrefix(text, "/maintenance") {
				next(ctx, bot2, update)
				return
			}
		}

		conf, err := loadMaintenanceConfig(maintenanceConfigPath)
		if err != nil {
			log.Printf("Can't load maintenance config: %v", err)
			next(ctx, bot2, update)
			return
		}

		if update.Message == nil {
			next(ctx, bot2, update)
			return
		}

		switch conf.Mode {
		case "degraded":
			yFile, err := os.ReadFile("/app/assets/maintenance/yellow.png")
			if err != nil {
				return
			}
			bot2.SendPhoto(ctx, &bot.SendPhotoParams{
				ChatID: update.Message.Chat.ID,
				Photo: &models.InputFileUpload{
					Filename: "YellowMaintenance",
					Data:     bytes.NewReader(yFile),
				},
			})
			return
		case "maintenance":
			rFile, err := os.ReadFile("/app/assets/maintenance/red.png")
			if err != nil {
				return
			}
			bot2.SendPhoto(ctx, &bot.SendPhotoParams{
				ChatID: update.Message.Chat.ID,
				Photo: &models.InputFileUpload{
					Filename: "RedMaintenance",
					Data:     bytes.NewReader(rFile),
				},
			})
			return
		default:
			next(ctx, bot2, update)
		}
	}
}
