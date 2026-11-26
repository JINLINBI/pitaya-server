package room

import (
	"context"
	"fmt"
	"strings"

	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

const groupName = "lobby"

// UserMessage 聊天消息。
type UserMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// NewUser 新用户通知。
type NewUser struct {
	Content string `json:"content"`
}

// AllMembers 房间成员列表。
type AllMembers struct {
	Members []string `json:"members"`
}

// JoinResponse 进房结果。
type JoinResponse struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
}

// Room 聊天房间组件。
type Room struct {
	component.Base
	app            pitaya.Pitaya
	frontendType   string // GroupBroadcast 目标前端类型
}

// New 创建 Room；frontendType 为 connector（集群）或本进程 server type（单体）。
func New(app pitaya.Pitaya, frontendType string) *Room {
	if frontendType == "" {
		frontendType = "connector"
	}
	return &Room{app: app, frontendType: frontendType}
}

// Init 创建大厅 group。
func (r *Room) Init() {
	_ = r.app.GroupCreate(context.Background(), groupName)
}

// Join 加入大厅并推送成员列表。
func (r *Room) Join(ctx context.Context) (*JoinResponse, error) {
	s := r.app.GetSessionFromCtx(ctx)
	uid := s.UID()
	if uid == "" {
		return &JoinResponse{Code: 401, Result: "not logged in"}, nil
	}
	if err := r.app.GroupAddMember(ctx, groupName, uid); err != nil {
		return nil, err
	}
	s.OnClose(func() {
		_ = r.app.GroupRemoveMember(context.Background(), groupName, uid)
	})
	members, err := r.app.GroupMembers(ctx, groupName)
	if err != nil {
		return nil, err
	}
	_ = s.Push("onMembers", &AllMembers{Members: members})
	_ = r.app.GroupBroadcast(ctx, r.frontendType, groupName, "onNewUser", &NewUser{
		Content: fmt.Sprintf("New user: %s", uid),
	})
	return &JoinResponse{Code: 200, Result: "success"}, nil
}

// Message 广播聊天消息（展示名取自会话，忽略客户端 name 防冒充）。
func (r *Room) Message(ctx context.Context, msg *UserMessage) {
	s := r.app.GetSessionFromCtx(ctx)
	name := s.UID()
	if v := s.Get("username"); v != nil {
		if uname, ok := v.(string); ok && strings.TrimSpace(uname) != "" {
			name = strings.TrimSpace(uname)
		}
	}
	content := strings.TrimSpace(msg.Content)
	if content == "" {
		return
	}
	_ = r.app.GroupBroadcast(ctx, r.frontendType, groupName, "onMessage", &UserMessage{
		Name:    name,
		Content: content,
	})
}
