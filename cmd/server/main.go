package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/JINLINBI/pitaya-server/internal/bootstrap"
	"github.com/JINLINBI/pitaya-server/internal/health"
	"github.com/JINLINBI/pitaya-server/internal/platform/pitayalog"
	"github.com/JINLINBI/pitaya-server/internal/serverwire"
	"github.com/JINLINBI/pitaya-server/internal/services/account"
	"github.com/JINLINBI/pitaya-server/internal/services/connector"
	"github.com/JINLINBI/pitaya-server/internal/services/mailbox"
	"github.com/JINLINBI/pitaya-server/internal/services/room"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/acceptor"
	"github.com/topfreegames/pitaya/v3/pkg/component"
	"github.com/topfreegames/pitaya/v3/pkg/config"
)

func main() {
	svType := flag.String("type", "connector", "server type: connector | room | account | monolith")
	port := flag.Int("port", 3250, "TCP acceptor port (frontend only)")
	nodeID := flag.Int("node_id", 0, "node id (for BTA workload args; informational)")
	healthPort := flag.Int("health_port", 8008, "HTTP health check port")
	flag.Parse()

	_ = nodeID

	mode := strings.ToLower(strings.TrimSpace(*svType))
	cluster := mode != "monolith" && mode != "standalone"

	pitayalog.Init(!cluster)

	platform, err := bootstrap.Init(mode, cluster)
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
	log.Printf("platform ready (%s)", platform.Settings.String())

	hs := health.New()
	if platform.Cfg != nil {
		hs.SetAdmin(health.AdminHooks{
			AdminToken: platform.Settings.AdminToken,
			OnCfgReload: func() error {
				return platform.Cfg.Reload(context.Background())
			},
		})
	}
	go func() {
		addr := health.AddrFromPort(*healthPort)
		log.Printf("health listening on %s", addr)
		if err := hs.ListenAndServe(addr); err != nil {
			log.Printf("health server stopped: %v", err)
		}
	}()

	switch mode {
	case "monolith", "standalone":
		runMonolith(*port, hs, platform)
	case "connector":
		runClusterFrontend(*port, hs, platform)
	case "room":
		runClusterBackend("room", hs, platform, false)
	case "account":
		runClusterBackend("account", hs, platform, true)
	default:
		log.Fatalf("unknown --type %q (want connector|room|account|monolith)", *svType)
	}
}

func connDeps(p *bootstrap.Platform) connector.Deps {
	return connector.Deps{JWT: p.JWT, Settings: p.Settings}
}

func acctDeps(p *bootstrap.Platform) account.Deps {
	return account.Deps{Accounts: p.Accounts}
}

func mailDeps(p *bootstrap.Platform) mailbox.Deps {
	return mailbox.Deps{Mailbox: p.Mailbox}
}

func attachGroups(builder *pitaya.Builder, cluster bool, cfg *config.PitayaConfig) {
	gs, err := serverwire.NewGroupService(cluster, cfg)
	if err != nil {
		log.Fatalf("groups: %v", err)
	}
	builder.Groups = gs
}

func runMonolith(port int, hs *health.Server, platform *bootstrap.Platform) {
	cfg := config.NewDefaultPitayaConfig()
	builder := pitaya.NewDefaultBuilder(true, "connector", pitaya.Standalone, map[string]string{}, *cfg)
	builder.AddAcceptor(acceptor.NewTCPAcceptor(fmt.Sprintf(":%d", port)))
	attachGroups(builder, false, cfg)
	app := builder.Build()
	defer app.Shutdown()

	registerFrontend(app, platform)
	registerAccount(app, platform)
	registerMailbox(app, platform)
	registerRoom(app, "connector")

	hs.SetReady(true)
	log.Printf("monolith listening tcp :%d", port)
	app.Start()
}

func runClusterFrontend(port int, hs *health.Server, platform *bootstrap.Platform) {
	cfg := config.NewDefaultPitayaConfig()
	serverwire.ApplyClusterEnv(cfg)

	builder := pitaya.NewDefaultBuilder(true, "connector", pitaya.Cluster, map[string]string{}, *cfg)
	builder.AddAcceptor(acceptor.NewTCPAcceptor(fmt.Sprintf(":%d", port)))
	attachGroups(builder, true, cfg)
	app := builder.Build()
	defer app.Shutdown()

	registerFrontend(app, platform)

	hs.SetReady(true)
	log.Printf("connector listening tcp :%d (etcd=%v nats=%s)",
		port, cfg.Cluster.SD.Etcd.Endpoints, cfg.Cluster.RPC.Client.Nats.Connect)
	app.Start()
}

func runClusterBackend(svType string, hs *health.Server, platform *bootstrap.Platform, withAccount bool) {
	cfg := config.NewDefaultPitayaConfig()
	serverwire.ApplyClusterEnv(cfg)

	builder := pitaya.NewDefaultBuilder(false, svType, pitaya.Cluster, map[string]string{}, *cfg)
	attachGroups(builder, true, cfg)
	app := builder.Build()
	defer app.Shutdown()

	if withAccount {
		// 仅 Register handler：客户端经 connector 用三元组路由转发（account.account.*）。
		registerAccount(app, platform)
		registerMailbox(app, platform)
		log.Printf("account backend started (etcd=%v nats=%s)",
			cfg.Cluster.SD.Etcd.Endpoints, cfg.Cluster.RPC.Client.Nats.Connect)
	} else {
		registerRoom(app, "connector")
		log.Printf("room backend started (groups=etcd etcd=%v nats=%s)",
			cfg.Cluster.SD.Etcd.Endpoints, cfg.Cluster.RPC.Client.Nats.Connect)
	}

	hs.SetReady(true)
	app.Start()
}

func registerFrontend(app pitaya.Pitaya, platform *bootstrap.Platform) {
	app.Register(connector.New(app, connDeps(platform)),
		component.WithName("connector"),
		component.WithNameFunc(strings.ToLower),
	)
}

func registerAccount(app pitaya.Pitaya, platform *bootstrap.Platform) {
	h := account.New(acctDeps(platform))
	app.Register(h,
		component.WithName("account"),
		component.WithNameFunc(strings.ToLower),
	)
}

func registerMailbox(app pitaya.Pitaya, platform *bootstrap.Platform) {
	h := mailbox.New(app, mailDeps(platform))
	app.Register(h,
		component.WithName("mailbox"),
		component.WithNameFunc(strings.ToLower),
	)
}

func registerRoom(app pitaya.Pitaya, frontendType string) {
	r := room.New(app, frontendType)
	app.Register(r,
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)
}
