package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DevJWTSecret 未配置 AUTH_JWT_SECRET 时的本地默认值（集群 / 生产禁止）。
const DevJWTSecret = "dev-change-me"

// MinJWTSecretLen 集群与生产环境 JWT 密钥最小长度。
const MinJWTSecretLen = 16

// Settings 从环境变量读取的运行时配置。
type Settings struct {
	DatabaseDSN    string
	DatabaseDriver string
	AuthJWTSecret  string
	LubanDataDir   string
	AdminToken     string
	AllowDemoLogin bool
	// EnvName 运行环境名：development（默认）| production。
	EnvName string
}

// Load 读取环境变量；未设置项使用开发默认值。
func Load() Settings {
	s := Settings{
		DatabaseDSN:    strings.TrimSpace(os.Getenv("DATABASE_DSN")),
		DatabaseDriver: strings.TrimSpace(os.Getenv("DATABASE_DRIVER")),
		AuthJWTSecret:  strings.TrimSpace(os.Getenv("AUTH_JWT_SECRET")),
		LubanDataDir:   strings.TrimSpace(os.Getenv("LUBAN_DATA_DIR")),
		AdminToken:     strings.TrimSpace(os.Getenv("ADMIN_TOKEN")),
		EnvName:        strings.ToLower(strings.TrimSpace(os.Getenv("PITAYA_ENV"))),
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
	if s.EnvName == "" {
		s.EnvName = "development"
	}
	s.AllowDemoLogin = truthy(os.Getenv("ALLOW_DEMO_LOGIN"))
	return s
}

// IsProduction 是否生产环境。
func (s Settings) IsProduction() bool {
	return s.EnvName == "production" || s.EnvName == "prod"
}

// Validate 校验运行配置。
// cluster=true 时强制显式 JWT、etcd、nats；production 额外拒绝弱密钥与 demo 登录。
func (s Settings) Validate(cluster bool) error {
	var errs []string

	if cluster || s.IsProduction() {
		if s.AuthJWTSecret == "" || s.AuthJWTSecret == DevJWTSecret {
			errs = append(errs, fmt.Sprintf(
				"AUTH_JWT_SECRET must be set to a non-default value (do not use %q)", DevJWTSecret))
		} else if len(s.AuthJWTSecret) < MinJWTSecretLen {
			errs = append(errs, fmt.Sprintf(
				"AUTH_JWT_SECRET must be at least %d characters (got %d)", MinJWTSecretLen, len(s.AuthJWTSecret)))
		}
	}

	if cluster {
		if strings.TrimSpace(os.Getenv("PITAYA_ETCD_ENDPOINTS")) == "" {
			errs = append(errs, "PITAYA_ETCD_ENDPOINTS is required in cluster mode (e.g. etcd:2379)")
		}
		if strings.TrimSpace(os.Getenv("PITAYA_NATS_URL")) == "" {
			errs = append(errs, "PITAYA_NATS_URL is required in cluster mode (e.g. nats://nats:4222)")
		}
	}

	if s.IsProduction() && s.AllowDemoLogin {
		errs = append(errs, "ALLOW_DEMO_LOGIN must not be enabled when PITAYA_ENV=production")
	}
	// 注意：生产环境缺省不设 ADMIN_TOKEN 更安全（不暴露 /admin/cfg/reload）。

	if len(errs) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString("invalid runtime config:")
	for _, e := range errs {
		b.WriteString("\n  - ")
		b.WriteString(e)
	}
	return errors.New(b.String())
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
	return fmt.Sprintf("env=%s db_driver=%s luban_dir=%s demo_login=%v",
		s.EnvName, s.DatabaseDriver, s.LubanDataDir, s.AllowDemoLogin)
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
