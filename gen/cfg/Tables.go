package cfg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Tables Luban 聚合根。
type Tables struct {
	TbGlobalConfig *GlobalConfig
}

// NewTables 从 data 目录加载 JSON 表（与 Luban data_json 布局一致）。
func NewTables(dataDir string) (*Tables, error) {
	loader := jsonLoader(dataDir)
	return newTablesFromLoader(loader)
}

func newTablesFromLoader(loader func(string) ([]byte, error)) (*Tables, error) {
	raw, err := loader("tbglobalconfig")
	if err != nil {
		return nil, fmt.Errorf("load tbglobalconfig: %w", err)
	}
	var rows []GlobalConfig
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse tbglobalconfig: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("tbglobalconfig: empty")
	}
	return &Tables{TbGlobalConfig: &rows[0]}, nil
}

func jsonLoader(dataDir string) func(string) ([]byte, error) {
	return func(table string) ([]byte, error) {
		name := table + ".json"
		path := filepath.Join(dataDir, name)
		return os.ReadFile(path)
	}
}