package connector

import (
	"context"
	"strings"

	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

// LoginReq 兼容 demo：仅 ALLOW_DEMO_LOGIN 时可用。
type LoginReq struct {
	Username string `json:"username"`
}

// LoginResp 登录响应。
type LoginResp struct {
	Code int    `json:"code"`
	UID  string `json:"uid"`
	Msg  string `json:"msg"`
}

// BindReq 使用 JWT 绑定会话。
type BindReq struct {
	Token string `json:"token"`
}

// BindResp 绑定结果。
type BindResp struct {
	Code int    `json:"code"`
	UID  string `json:"uid"`
	Msg  string `json:"msg"`
}

// Connector 前端连接器：会话绑定。
type Connector struct {
	component.Base
	app  pitaya.Pitaya
	deps Deps
}

// New 创建 Connector。
func New(app pitaya.Pitaya, d Deps) *Connector {
	return &Connector{app: app, deps: d}
}

// Login 无密码 demo 登录（ALLOW_DEMO_LOGIN=1）。
func (c *Connector) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	if !c.deps.Settings.AllowDemoLogin {
		return &LoginResp{Code: 403, Msg: "demo login disabled"}, nil
	}
	uid := strings.TrimSpace(req.Username)
	if uid == "" {
		return &LoginResp{Code: 400, Msg: "username required"}, nil
	}
	s := c.app.GetSessionFromCtx(ctx)
	if err := s.Bind(ctx, uid); err != nil {
		return nil, pitaya.Error(err, "CN-LOGIN", map[string]string{"failed": "bind"})
	}
	_ = s.Set("username", uid)
	return &LoginResp{Code: 200, UID: uid, Msg: "ok"}, nil
}

// Bind 校验 JWT 并绑定 Pitaya 会话。
func (c *Connector) Bind(ctx context.Context, req *BindReq) (*BindResp, error) {
	token := strings.TrimSpace(req.Token)
	claims, err := c.deps.JWT.VerifyClaims(token)
	if err != nil {
		return &BindResp{Code: 401, Msg: "invalid token"}, nil
	}
	playerID := claims.PlayerID
	s := c.app.GetSessionFromCtx(ctx)
	if err := s.Bind(ctx, playerID); err != nil {
		return nil, pitaya.Error(err, "CN-BIND", map[string]string{"failed": "bind"})
	}
	_ = s.Set("player_id", playerID)
	if claims.Username != "" {
		_ = s.Set("username", claims.Username)
	}
	return &BindResp{Code: 200, UID: playerID, Msg: "ok"}, nil
}