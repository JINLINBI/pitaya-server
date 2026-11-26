package account

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository 账户持久化。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 写入新账户。
func (r *Repository) Create(ctx context.Context, e *Entity) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// FindByUsername 按用户名查询。
func (r *Repository) FindByUsername(ctx context.Context, username string) (*Entity, error) {
	var e Entity
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// FindByID 按主键查询。
func (r *Repository) FindByID(ctx context.Context, id uint64) (*Entity, error) {
	var e Entity
	err := r.db.WithContext(ctx).First(&e, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}