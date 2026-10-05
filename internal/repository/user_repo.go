package repository

import (
	"context"
	"lotoMironBot/internal/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Register(ctx context.Context, user *database.User, owner bool) error {
	updates := map[string]any{
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}
	if owner {
		updates["role"] = database.RoleOwner
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "telegram_id"}},
		DoUpdates: clause.Assignments(updates),
	}).Create(user).Error
}

func (r *UserRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*database.User, error) {
	var user database.User
	if err := r.db.WithContext(ctx).First(&user, "telegram_id = ?", telegramID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) SetRole(ctx context.Context, telegramID int64, role database.UserRole) error {
	return r.db.WithContext(ctx).Model(&database.User{}).
		Where("telegram_id = ?", telegramID).
		Update("role", role).Error
}
