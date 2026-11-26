package mailbox

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository 邮箱持久化。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 写入消息。
func (r *Repository) Create(ctx context.Context, m *Message) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// ListByPlayer 按玩家列出，最新在前。
func (r *Repository) ListByPlayer(ctx context.Context, playerID string, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []Message
	err := r.db.WithContext(ctx).
		Where("player_id = ?", playerID).
		Order("id desc").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// FindByID 按主键查询。
func (r *Repository) FindByID(ctx context.Context, id uint64) (*Message, error) {
	var m Message
	err := r.db.WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// MarkRead 标记已读。
func (r *Repository) MarkRead(ctx context.Context, id uint64, playerID string) error {
	res := r.db.WithContext(ctx).
		Model(&Message{}).
		Where("id = ? AND player_id = ?", id, playerID).
		Update("read", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
