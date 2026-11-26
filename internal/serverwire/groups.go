package serverwire

import (
	"fmt"
	"os"
	"strings"

	"github.com/topfreegames/pitaya/v3/pkg/config"
	"github.com/topfreegames/pitaya/v3/pkg/groups"
)

// ApplyClusterEnv 用环境变量覆盖 Pitaya 集群与 Groups 的 etcd / NATS 配置。
func ApplyClusterEnv(cfg *config.PitayaConfig) {
	if v := strings.TrimSpace(os.Getenv("PITAYA_ETCD_ENDPOINTS")); v != "" {
		parts := splitCSV(v)
		cfg.Cluster.SD.Etcd.Endpoints = parts
		cfg.Groups.Etcd.Endpoints = parts
	}
	if v := strings.TrimSpace(os.Getenv("PITAYA_NATS_URL")); v != "" {
		cfg.Cluster.RPC.Client.Nats.Connect = v
		cfg.Cluster.RPC.Server.Nats.Connect = v
	}
}

// NewGroupService 单体用内存群组；集群用 etcd 共享 lobby 成员。
func NewGroupService(cluster bool, cfg *config.PitayaConfig) (groups.GroupService, error) {
	if !cluster {
		return groups.NewMemoryGroupService(cfg.Groups.Memory), nil
	}
	if len(cfg.Groups.Etcd.Endpoints) == 0 {
		return nil, fmt.Errorf("cluster mode requires PITAYA_ETCD_ENDPOINTS for group service")
	}
	return groups.NewEtcdGroupService(cfg.Groups.Etcd, nil)
}

func splitCSV(s string) []string {
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}