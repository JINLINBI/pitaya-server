package health

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Server 提供 K8s / BTA 健康检查 HTTP。
type Server struct {
	ready atomic.Bool
}

// New 创建健康检查服务；初始未就绪。
func New() *Server {
	return &Server{}
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
	return http.ListenAndServe(addr, mux)
}

// AddrFromPort 拼健康检查监听地址。
func AddrFromPort(port int) string {
	return fmt.Sprintf(":%d", port)
}
