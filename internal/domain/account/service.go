package account

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	gencfg "github.com/JINLINBI/pitaya-server/gen/cfg"
	"github.com/JINLINBI/pitaya-server/internal/platform/auth"
	"golang.org/x/crypto/bcrypt"
)

// Service 账户用例。
type Service struct {
	repo *Repository
	jwt  *auth.JWT
	cfg  func() *gencfg.Tables
}

// NewService 创建账户服务。
func NewService(repo *Repository, jwt *auth.JWT, cfg func() *gencfg.Tables) *Service {
	return &Service{repo: repo, jwt: jwt, cfg: cfg}
}

// RegisterResult 注册结果。
type RegisterResult struct {
	PlayerID string
}

// LoginResult 登录结果。
type LoginResult struct {
	PlayerID string
	Token    string
	ExpireAt time.Time
}

var (
	ErrUsernameTaken   = errors.New("username taken")
	ErrInvalidCreds    = errors.New("invalid credentials")
	ErrAccountDisabled = errors.New("account disabled")
)

// Register 注册新账户。
func (s *Service) Register(ctx context.Context, username, password string) (*RegisterResult, error) {
	username = strings.TrimSpace(username)
	if err := s.validateUsername(username); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	exists, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return nil, ErrUsernameTaken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	ent := &Entity{Username: username, PasswordHash: string(hash), Status: 1}
	if err := s.repo.Create(ctx, ent); err != nil {
		return nil, err
	}
	return &RegisterResult{PlayerID: ent.PlayerID()}, nil
}

// Login 校验密码并签发 JWT。
func (s *Service) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	username = strings.TrimSpace(username)
	ent, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if ent == nil {
		return nil, ErrInvalidCreds
	}
	if ent.Status != 1 {
		return nil, ErrAccountDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(ent.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCreds
	}
	playerID := ent.PlayerID()
	token, exp, err := s.jwt.Issue(playerID, username)
	if err != nil {
		return nil, err
	}
	return &LoginResult{PlayerID: playerID, Token: token, ExpireAt: exp}, nil
}

func (s *Service) validateUsername(username string) error {
	n := utf8.RuneCountInString(username)
	t := s.cfg()
	if t == nil || t.TbGlobalConfig == nil {
		if n < 3 || n > 32 {
			return fmt.Errorf("username length must be 3-32")
		}
		return nil
	}
	minL := int(t.TbGlobalConfig.MinUsernameLen)
	maxL := int(t.TbGlobalConfig.MaxUsernameLen)
	if minL <= 0 {
		minL = 3
	}
	if maxL <= 0 {
		maxL = 32
	}
	if n < minL || n > maxL {
		return fmt.Errorf("username length must be %d-%d", minL, maxL)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password too short (min 6)")
	}
	return nil
}