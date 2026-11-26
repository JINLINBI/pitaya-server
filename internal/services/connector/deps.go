package connector

import (
	"github.com/JINLINBI/pitaya-server/internal/platform/auth"
	"github.com/JINLINBI/pitaya-server/internal/platform/env"
)

// Deps 前端 connector 依赖。
type Deps struct {
	JWT      *auth.JWT
	Settings env.Settings
}