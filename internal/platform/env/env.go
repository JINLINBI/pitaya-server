package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DevJWTSecret 未配置 AUTH_JWT_SECRET 时的本地默认值（集群禁止）。
const DevJWTSecret = "dev-change-me"

// Settings 从环境变量读取的运行时配置。
type Settings struct {
	DatabaseDSN     string
	DatabaseDriver  string
	AuthJWTSecret   string
	LubanDataDir    string
	AdminToken      string
	AllowDemoLogin  bool
}

// Load 读取环境变量；未设置项使用开发默认值。
func Load() Settings {
	s := Settings{
		DatabaseDSN:    strings.TrimSpace(os.Getenv("DATABASE_DSN")),
		DatabaseDriver: strings.TrimSpace(os.Getenv("DATABASE_DRIVER")),
		AuthJWTSecret:  strings.TrimSpace(os.Getenv("AUTH_JWT_SECRET")),
		LubanDataDir:   strings.TrimSpace(os.Getenv("LUBAN_DATA_DIR")),
		AdminToken:     strings.TrimSpace(os.Getenv("ADMIN_TOKEN")),
	}
	if s.DatabaseDSN == "" {
		s.DatabaseDSN = "pitaya.db"
	}
	if s.DatabaseDriver == "" {
		if strings.Contains(s.DatabaseDSN, "://") || strings.HasPrefix(s.DatabaseDSN, "postgres") {
			s.DatabaseDriver = inferDriver(s.DatabaseDSN)
		} else {
			s.DatabaseDriver = "sqlite"
		}
	}
	if s.LubanDataDir == "" {
		s.LubanDataDir = "data/cfg"
	}
	if s.AuthJWTSecret == "" {
		s.AuthJWTSecret = DevJWTSecret
	}
	s.AllowDemoLogin = truthy(os.Getenv("ALLOW_DEMO_LOGIN"))
	return s
}

// Validate 校验运行配置；cluster 为 true 时拒绝默认 JWT 密钥。
func (s Settings) Validate(cluster bool) error {
	if cluster && (s.AuthJWTSecret == "" || s.AuthJWTSecret == DevJWTSecret) {
		return errors.New("cluster mode requires AUTH_JWT_SECRET to be set explicitly")
	}
	return nil
}

func inferDriver(dsn string) string {
	lower := strings.ToLower(dsn)
	switch {
	case strings.HasPrefix(lower, "postgres"):
		return "postgres"
	case strings.HasPrefix(lower, "mysql"):
		return "mysql"
	default:
		return "sqlite"
	}
}

func truthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// NeedDB 该进程是否需要数据库（connector/room 默认不连库）。
func NeedDB(serverType string) bool {
	t := strings.ToLower(strings.TrimSpace(serverType))
	switch t {
	case "monolith", "standalone", "account":
		return true
	default:
		return false
	}
}

// String 脱敏摘要（日志用）。
func (s Settings) String() string {
	return fmt.Sprintf("db_driver=%s luban_dir=%s demo_login=%v",
		s.DatabaseDriver, s.LubanDataDir, s.AllowDemoLogin)
}

// ParseBoolEnv 解析布尔环境变量（测试用导出）。
func ParseBoolEnv(key string) bool {
	return truthy(os.Getenv(key))
}

// PortFromEnv 读取整型环境变量。
func PortFromEnv(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}