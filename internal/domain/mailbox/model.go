package mailbox

import "time"

// Message 玩家邮箱消息（示例领域模型，无 Pitaya 依赖）。
type Message struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	PlayerID  string    `gorm:"index;size:64;not null"`
	Title     string    `gorm:"size:128;not null"`
	Body      string    `gorm:"size:2048;not null"`
	Read      bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName GORM 表名。
func (Message) TableName() string { return "mailbox_messages" }
