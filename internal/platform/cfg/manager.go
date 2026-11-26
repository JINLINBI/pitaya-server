package cfgmgr

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	gencfg "github.com/JINLINBI/pitaya-server/gen/cfg"
)

// Manager 原子持有 Luban Tables，支持热重载。
type Manager struct {
	dataDir string
	val     atomic.Pointer[gencfg.Tables]
}

// NewManager 加载并返回配置管理器。
func NewManager(dataDir string) (*Manager, error) {
	m := &Manager{dataDir: dataDir}
	if err := m.Reload(context.Background()); err != nil {
		return nil, err
	}
	return m, nil
}

// Get 返回当前配置快照。
func (m *Manager) Get() *gencfg.Tables {
	return m.val.Load()
}

// Reload 从磁盘重新加载；失败时保留旧表。
func (m *Manager) Reload(ctx context.Context) error {
	_ = ctx
	tables, err := gencfg.NewTables(m.dataDir)
	if err != nil {
		log.Printf("cfg reload failed (keeping previous): %v", err)
		return fmt.Errorf("cfg reload: %w", err)
	}
	m.val.Store(tables)
	log.Printf("cfg loaded from %s (min_username_len=%d max_username_len=%d)",
		m.dataDir,
		tables.TbGlobalConfig.MinUsernameLen,
		tables.TbGlobalConfig.MaxUsernameLen,
	)
	return nil
}