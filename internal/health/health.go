package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

// AdminHooks 可选管理端点（需 ADMIN_TOKEN）。
type AdminHooks struct {
	AdminToken string
	OnCfgReload func() error
}

// Server 提供 K8s / BTA 健康检查 HTTP。
type Server struct {
	ready atomic.Bool
	admin AdminHooks
}

// New 创建健康检查服务；初始未就绪。
func New() *Server {
	return &Server{}
}

// SetAdmin 配置管理路由；未设置 token 时不暴露 admin。
func (s *Server) SetAdmin(h AdminHooks) {
	s.admin = h
}

// SetReady 标记进程已可对外服务。
func (s *Server) SetReady(v bool) {
	s.ready.Store(v)
}

// ListenAndServe 在 addr 上提供 /healthz。
func (s *Server) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		if !s.ready.Load() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		if !s.ready.Load() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	if token := strings.TrimSpace(s.admin.AdminToken); token != "" && s.admin.OnCfgReload != nil {
		mux.HandleFunc("/admin/cfg/reload", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if !s.ready.Load() {
				http.Error(w, "not ready", http.StatusServiceUnavailable)
				return
			}
			if strings.TrimSpace(r.Header.Get("X-Admin-Token")) != token {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if err := s.admin.OnCfgReload(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "reloaded"})
		})
	}
	return http.ListenAndServe(addr, mux)
}

// AddrFromPort 拼健康检查监听地址。
func AddrFromPort(port int) string {
	return fmt.Sprintf(":%d", port)
}