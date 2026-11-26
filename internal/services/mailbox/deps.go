package mailbox

import domain "github.com/JINLINBI/pitaya-server/internal/domain/mailbox"

// Deps 邮箱 Pitaya 组件依赖（不依赖 bootstrap）。
type Deps struct {
	Mailbox *domain.Service
}
