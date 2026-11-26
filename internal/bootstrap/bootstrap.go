package bootstrap

import (
	"fmt"

	"github.com/JINLINBI/pitaya-server/internal/domain/account"
	"github.com/JINLINBI/pitaya-server/internal/platform/auth"
	cfgmgr "github.com/JINLINBI/pitaya-server/internal/platform/cfg"
	"github.com/JINLINBI/pitaya-server/internal/platform/db"
	"github.com/JINLINBI/pitaya-server/internal/platform/env"
	"gorm.io/gorm"
)

// Platform 各服务共享的平台依赖。
type Platform struct {
	Settings env.Settings
	Cfg      *cfgmgr.Manager
	JWT      *auth.JWT
	DB       *gorm.DB
	Accounts *account.Service
}

// Init 按进程类型初始化平台组件。
func Init(serverType string, cluster bool) (*Platform, error) {
	st := env.Load()
	if err := st.Validate(cluster); err != nil {
		return nil, err
	}
	needDB := env.NeedDB(serverType)

	cfg, err := cfgmgr.NewManager(st.LubanDataDir)
	if err != nil {
		return nil, fmt.Errorf("cfg: %w", err)
	}
	jwt := auth.NewJWT(st.AuthJWTSecret, 0)

	p := &Platform{
		Settings: st,
		Cfg:      cfg,
		JWT:      jwt,
	}

	if needDB {
		gdb, err := db.Open(st)
		if err != nil {
			return nil, fmt.Errorf("db: %w", err)
		}
		p.DB = gdb
		repo := account.NewRepository(gdb)
		p.Accounts = account.NewService(repo, jwt, cfg.Get)
	}
	return p, nil
}