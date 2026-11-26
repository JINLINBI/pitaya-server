package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultTokenTTL = 24 * time.Hour

// JWT 签发与校验 access token。
type JWT struct {
	secret []byte
	ttl    time.Duration
}

// NewJWT 创建 JWT 助手。
func NewJWT(secret string, ttl time.Duration) *JWT {
	if ttl <= 0 {
		ttl = defaultTokenTTL
	}
	return &JWT{secret: []byte(secret), ttl: ttl}
}

// Claims 自定义载荷。
type Claims struct {
	PlayerID string `json:"player_id"`
	Username string `json:"username,omitempty"`
	jwt.RegisteredClaims
}

// Issue 为玩家签发 token；username 写入 claims 供 connector 绑定会话展示名。
func (j *JWT) Issue(playerID, username string) (token string, expireAt time.Time, err error) {
	expireAt = time.Now().Add(j.ttl)
	claims := Claims{
		PlayerID: playerID,
		Username: strings.TrimSpace(username),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   playerID,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tok.SignedString(j.secret)
	return token, expireAt, err
}

// Verify 校验 token 并返回 player_id。
func (j *JWT) Verify(token string) (string, error) {
	c, err := j.VerifyClaims(token)
	if err != nil {
		return "", err
	}
	return c.PlayerID, nil
}

// VerifyClaims 校验 token 并返回完整 claims。
func (j *JWT) VerifyClaims(token string) (*Claims, error) {
	if token == "" {
		return nil, errors.New("token required")
	}
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.PlayerID == "" {
		return nil, errors.New("missing player_id")
	}
	return claims, nil
}