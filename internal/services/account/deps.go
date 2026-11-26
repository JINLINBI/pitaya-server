package account

import domain "github.com/JINLINBI/pitaya-server/internal/domain/account"

// Deps 账户 Pitaya 组件依赖（不依赖 bootstrap）。
type Deps struct {
	Accounts *domain.Service
}