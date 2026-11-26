package account

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/JINLINBI/pitaya-server/internal/domain/account"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

// RegisterReq 注册。
type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResp 注册响应。
type RegisterResp struct {
	Code     int    `json:"code"`
	PlayerID string `json:"player_id"`
	Msg      string `json:"msg"`
}

// LoginReq 登录。
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResp 登录响应。
type LoginResp struct {
	Code     int    `json:"code"`
	Token    string `json:"token"`
	PlayerID string `json:"player_id"`
	ExpireAt int64  `json:"expire_at"`
	Msg      string `json:"msg"`
}

// Handler 账户 Pitaya 组件。
type Handler struct {
	component.Base
	accounts *domain.Service
}

// New 创建 Handler。
func New(d Deps) *Handler {
	return &Handler{accounts: d.Accounts}
}

// Register 注册账号。
func (h *Handler) Register(ctx context.Context, req *RegisterReq) (*RegisterResp, error) {
	if h.accounts == nil {
		return nil, pitaya.Error(errors.New("account service unavailable"), "CN-ACCOUNT", nil)
	}
	res, err := h.accounts.Register(ctx, req.Username, req.Password)
	if err != nil {
		return mapRegisterErr(err), nil
	}
	return &RegisterResp{Code: 200, PlayerID: res.PlayerID, Msg: "ok"}, nil
}

// Login 登录并返回 JWT。
func (h *Handler) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	if h.accounts == nil {
		return nil, pitaya.Error(errors.New("account service unavailable"), "CN-ACCOUNT", nil)
	}
	res, err := h.accounts.Login(ctx, req.Username, req.Password)
	if err != nil {
		return mapLoginErr(err), nil
	}
	return &LoginResp{
		Code:     200,
		Token:    res.Token,
		PlayerID: res.PlayerID,
		ExpireAt: res.ExpireAt.Unix(),
		Msg:      "ok",
	}, nil
}

func mapRegisterErr(err error) *RegisterResp {
	switch {
	case errors.Is(err, domain.ErrUsernameTaken):
		return &RegisterResp{Code: 409, Msg: "username taken"}
	default:
		return &RegisterResp{Code: 400, Msg: err.Error()}
	}
}

func mapLoginErr(err error) *LoginResp {
	switch {
	case errors.Is(err, domain.ErrInvalidCreds):
		return &LoginResp{Code: 401, Msg: "invalid credentials"}
	case errors.Is(err, domain.ErrAccountDisabled):
		return &LoginResp{Code: 403, Msg: "account disabled"}
	default:
		return &LoginResp{Code: 400, Msg: err.Error()}
	}
}

// Name 调试用。
func (h *Handler) String() string { return fmt.Sprintf("account") }