package account

import (
	"fmt"
	"time"
)

// Entity 账户（player_id 与 Pitaya UID 一致，为十进制字符串）。
type Entity struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"uniqueIndex;size:64;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Status       int8      `gorm:"not null;default:1"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

// PlayerID 返回 Pitaya 绑定用的 UID。
func (e Entity) PlayerID() string {
	return fmt.Sprintf("%d", e.ID)
}

// TableName GORM 表名。
func (Entity) TableName() string { return "accounts" }