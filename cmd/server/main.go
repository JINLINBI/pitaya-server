package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/JINLINBI/pitaya-server/internal/health"
	"github.com/JINLINBI/pitaya-server/internal/services/connector"
	"github.com/JINLINBI/pitaya-server/internal/services/room"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/acceptor"
	"github.com/topfreegames/pitaya/v3/pkg/component"
	"github.com/topfreegames/pitaya/v3/pkg/config"
	"github.com/topfreegames/pitaya/v3/pkg/groups"
)

func main() {
	svType := flag.String("type", "connector", "server type: connector | room | monolith")
	port := flag.Int("port", 3250, "TCP acceptor port (frontend only)")
	nodeID := flag.Int("node_id", 0, "node id (for BTA workload args; informational)")
	healthPort := flag.Int("health_port", 8008, "HTTP health check port")
	flag.Parse()

	_ = nodeID // reserved for future; BTA passes --node_id

	hs := health.New()
	go func() {
		addr := health.AddrFromPort(*healthPort)
		log.Printf("health listening on %s", addr)
		if err := hs.ListenAndServe(addr); err != nil {
			log.Printf("health server stopped: %v", err)
		}
	}()

	mode := strings.ToLower(strings.TrimSpace(*svType))
	switch mode {
	case "monolith", "standalone":
		runMonolith(*port, hs)
	case "connector":
		runCluster(true, "connector", *port, hs)
	case "room":
		runCluster(false, "room", *port, hs)
	default:
		log.Fatalf("unknown --type %q (want connector|room|monolith)", *svType)
	}
}

func applyClusterEnv(cfg *config.PitayaConfig) {
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

func runMonolith(port int, hs *health.Server) {
	cfg := config.NewDefaultPitayaConfig()
	builder := pitaya.NewDefaultBuilder(true, "connector", pitaya.Standalone, map[string]string{}, *cfg)
	builder.AddAcceptor(acceptor.NewTCPAcceptor(fmt.Sprintf(":%d", port)))
	builder.Groups = groups.NewMemoryGroupService(builder.Config.Groups.Memory)
	app := builder.Build()
	defer app.Shutdown()

	app.Register(connector.New(app),
		component.WithName("connector"),
		component.WithNameFunc(strings.ToLower),
	)
	app.Register(room.New(app, "connector"),
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)

	hs.SetReady(true)
	log.Printf("monolith listening tcp :%d", port)
	app.Start()
}

func runCluster(isFrontend bool, svType string, port int, hs *health.Server) {
	cfg := config.NewDefaultPitayaConfig()
	applyClusterEnv(cfg)

	builder := pitaya.NewDefaultBuilder(isFrontend, svType, pitaya.Cluster, map[string]string{}, *cfg)
	if isFrontend {
		builder.AddAcceptor(acceptor.NewTCPAcceptor(fmt.Sprintf(":%d", port)))
	}
	builder.Groups = groups.NewMemoryGroupService(builder.Config.Groups.Memory)
	app := builder.Build()
	defer app.Shutdown()

	if isFrontend {
		app.Register(connector.New(app),
			component.WithName("connector"),
			component.WithNameFunc(strings.ToLower),
		)
		log.Printf("connector listening tcp :%d (etcd=%v nats=%s)",
			port, cfg.Cluster.SD.Etcd.Endpoints, cfg.Cluster.RPC.Client.Nats.Connect)
	} else {
		r := room.New(app, "connector")
		app.Register(r,
			component.WithName("room"),
			component.WithNameFunc(strings.ToLower),
		)
		app.RegisterRemote(r,
			component.WithName("room"),
			component.WithNameFunc(strings.ToLower),
		)
		log.Printf("room backend started (etcd=%v nats=%s)",
			cfg.Cluster.SD.Etcd.Endpoints, cfg.Cluster.RPC.Client.Nats.Connect)
	}

	hs.SetReady(true)
	app.Start()
}
