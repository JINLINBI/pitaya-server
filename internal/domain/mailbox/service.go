package mailbox

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrNotFound      = errors.New("message not found")
	ErrEmptyPlayer   = errors.New("player_id required")
	ErrInvalidTitle  = errors.New("title required (1-64 chars)")
	ErrInvalidBody   = errors.New("body required (1-1024 chars)")
)

// Service 邮箱用例（纯领域，不依赖 Pitaya）。
type Service struct {
	repo *Repository
}

// NewService 创建邮箱服务。
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SendInput 发送入参。
type SendInput struct {
	PlayerID string
	Title    string
	Body     string
}

// Send 向玩家邮箱投递一条消息（示例：可由系统或其它玩法调用）。
func (s *Service) Send(ctx context.Context, in SendInput) (*Message, error) {
	playerID := strings.TrimSpace(in.PlayerID)
	title := strings.TrimSpace(in.Title)
	body := strings.TrimSpace(in.Body)
	if playerID == "" {
		return nil, ErrEmptyPlayer
	}
	if n := utf8.RuneCountInString(title); n < 1 || n > 64 {
		return nil, ErrInvalidTitle
	}
	if n := utf8.RuneCountInString(body); n < 1 || n > 1024 {
		return nil, ErrInvalidBody
	}
	m := &Message{PlayerID: playerID, Title: title, Body: body}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// List 列出玩家邮箱。
func (s *Service) List(ctx context.Context, playerID string, limit int) ([]Message, error) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return nil, ErrEmptyPlayer
	}
	return s.repo.ListByPlayer(ctx, playerID, limit)
}

// MarkRead 标记已读（仅本人）。
func (s *Service) MarkRead(ctx context.Context, playerID string, id uint64) error {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return ErrEmptyPlayer
	}
	return s.repo.MarkRead(ctx, id, playerID)
}
