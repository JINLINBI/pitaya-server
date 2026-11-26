package cfgmgr

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestManagerReload(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join("..", "..", "..", "data", "cfg", "tbglobalconfig.json")
	raw, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tbglobalconfig.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	g := m.Get()
	if g.TbGlobalConfig.MaxUsernameLen != 32 {
		t.Fatalf("max len %d", g.TbGlobalConfig.MaxUsernameLen)
	}
	if err := m.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
}