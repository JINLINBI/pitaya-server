package connector

import (
	"context"
	"strings"

	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

// LoginReq 登录请求。
type LoginReq struct {
	Username string `json:"username"`
}

// LoginResp 登录响应。
type LoginResp struct {
	Code int    `json:"code"`
	UID  string `json:"uid"`
	Msg  string `json:"msg"`
}

// Connector 前端连接器：负责会话绑定（登录）。
type Connector struct {
	component.Base
	app pitaya.Pitaya
}

// New 创建 Connector。
func New(app pitaya.Pitaya) *Connector {
	return &Connector{app: app}
}

// Login 用用户名绑定会话 UID（无密码，demo 用途）。
func (c *Connector) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
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
