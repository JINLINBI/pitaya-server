package mailbox

import (
	"context"
	"errors"

	domain "github.com/JINLINBI/pitaya-server/internal/domain/mailbox"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

// SendReq 投递邮件（示例：当前玩家给自己发一条，便于 CLI 验证）。
type SendReq struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// SendResp 投递结果。
type SendResp struct {
	Code int    `json:"code"`
	ID   uint64 `json:"id"`
	Msg  string `json:"msg"`
}

// ListResp 列表。
type ListResp struct {
	Code     int         `json:"code"`
	Messages []MsgView   `json:"messages"`
	Msg      string      `json:"msg"`
}

// MsgView 对外视图。
type MsgView struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Read  bool   `json:"read"`
}

// MarkReadReq 标记已读。
type MarkReadReq struct {
	ID uint64 `json:"id"`
}

// MarkReadResp 标记结果。
type MarkReadResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Handler 邮箱 Pitaya 组件（薄适配层）。
type Handler struct {
	component.Base
	app     pitaya.Pitaya
	mailbox *domain.Service
}

// New 创建 Handler。
func New(app pitaya.Pitaya, d Deps) *Handler {
	return &Handler{app: app, mailbox: d.Mailbox}
}

// Send 向当前会话玩家投递邮件。
func (h *Handler) Send(ctx context.Context, req *SendReq) (*SendResp, error) {
	if h.mailbox == nil {
		return nil, pitaya.Error(errors.New("mailbox unavailable"), "CN-MAILBOX", nil)
	}
	uid, err := h.uid(ctx)
	if err != nil {
		return &SendResp{Code: 401, Msg: "not logged in"}, nil
	}
	m, err := h.mailbox.Send(ctx, domain.SendInput{
		PlayerID: uid,
		Title:    req.Title,
		Body:     req.Body,
	})
	if err != nil {
		return &SendResp{Code: 400, Msg: err.Error()}, nil
	}
	return &SendResp{Code: 200, ID: m.ID, Msg: "ok"}, nil
}

// List 列出当前玩家邮箱。
func (h *Handler) List(ctx context.Context) (*ListResp, error) {
	if h.mailbox == nil {
		return nil, pitaya.Error(errors.New("mailbox unavailable"), "CN-MAILBOX", nil)
	}
	uid, err := h.uid(ctx)
	if err != nil {
		return &ListResp{Code: 401, Msg: "not logged in"}, nil
	}
	rows, err := h.mailbox.List(ctx, uid, 50)
	if err != nil {
		return &ListResp{Code: 500, Msg: err.Error()}, nil
	}
	out := make([]MsgView, 0, len(rows))
	for _, r := range rows {
		out = append(out, MsgView{ID: r.ID, Title: r.Title, Body: r.Body, Read: r.Read})
	}
	return &ListResp{Code: 200, Messages: out, Msg: "ok"}, nil
}

// MarkRead 标记已读。
func (h *Handler) MarkRead(ctx context.Context, req *MarkReadReq) (*MarkReadResp, error) {
	if h.mailbox == nil {
		return nil, pitaya.Error(errors.New("mailbox unavailable"), "CN-MAILBOX", nil)
	}
	uid, err := h.uid(ctx)
	if err != nil {
		return &MarkReadResp{Code: 401, Msg: "not logged in"}, nil
	}
	if err := h.mailbox.MarkRead(ctx, uid, req.ID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return &MarkReadResp{Code: 404, Msg: "not found"}, nil
		}
		return &MarkReadResp{Code: 400, Msg: err.Error()}, nil
	}
	return &MarkReadResp{Code: 200, Msg: "ok"}, nil
}

func (h *Handler) uid(ctx context.Context) (string, error) {
	s := h.app.GetSessionFromCtx(ctx)
	if s == nil || s.UID() == "" {
		return "", errors.New("not bound")
	}
	return s.UID(), nil
}
