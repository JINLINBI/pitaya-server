package env

import (
	"strings"
	"testing"
)

func TestNeedDB(t *testing.T) {
	cases := []struct {
		typ  string
		want bool
	}{
		{"monolith", true},
		{"standalone", true},
		{"account", true},
		{"connector", false},
		{"room", false},
	}
	for _, c := range cases {
		if got := NeedDB(c.typ); got != c.want {
			t.Fatalf("%s: want %v got %v", c.typ, c.want, got)
		}
	}
}

func TestValidateJWTSecret(t *testing.T) {
	dev := Settings{AuthJWTSecret: DevJWTSecret, EnvName: "development"}
	if err := dev.Validate(true); err == nil {
		t.Fatal("cluster should reject dev secret")
	}
	if err := dev.Validate(false); err != nil {
		t.Fatalf("monolith dev secret: %v", err)
	}
	ok := Settings{AuthJWTSecret: "prod-secret-long", EnvName: "development"}
	t.Setenv("PITAYA_ETCD_ENDPOINTS", "etcd:2379")
	t.Setenv("PITAYA_NATS_URL", "nats://nats:4222")
	if err := ok.Validate(true); err != nil {
		t.Fatalf("cluster custom secret: %v", err)
	}
}

func TestValidateClusterRequiresEtcdNats(t *testing.T) {
	t.Setenv("PITAYA_ETCD_ENDPOINTS", "")
	t.Setenv("PITAYA_NATS_URL", "")
	s := Settings{AuthJWTSecret: "compose-dev-secret", EnvName: "development"}
	err := s.Validate(true)
	if err == nil {
		t.Fatal("expected etcd/nats errors")
	}
	msg := err.Error()
	if !strings.Contains(msg, "PITAYA_ETCD_ENDPOINTS") {
		t.Fatalf("want etcd error, got: %s", msg)
	}
	if !strings.Contains(msg, "PITAYA_NATS_URL") {
		t.Fatalf("want nats error, got: %s", msg)
	}
}

func TestValidateProductionRejectsDemoLogin(t *testing.T) {
	t.Setenv("PITAYA_ETCD_ENDPOINTS", "etcd:2379")
	t.Setenv("PITAYA_NATS_URL", "nats://nats:4222")
	s := Settings{
		AuthJWTSecret:  "production-secret-ok",
		EnvName:        "production",
		AllowDemoLogin: true,
	}
	err := s.Validate(true)
	if err == nil || !strings.Contains(err.Error(), "ALLOW_DEMO_LOGIN") {
		t.Fatalf("want demo-login error, got: %v", err)
	}
}

func TestValidateProductionAllowsEmptyAdminToken(t *testing.T) {
	t.Setenv("PITAYA_ETCD_ENDPOINTS", "etcd:2379")
	t.Setenv("PITAYA_NATS_URL", "nats://nats:4222")
	s := Settings{
		AuthJWTSecret:  "production-secret-ok",
		EnvName:        "production",
		AllowDemoLogin: false,
		AdminToken:     "",
	}
	if err := s.Validate(true); err != nil {
		t.Fatalf("empty ADMIN_TOKEN should be ok in production: %v", err)
	}
}

func TestValidateShortJWTSecret(t *testing.T) {
	t.Setenv("PITAYA_ETCD_ENDPOINTS", "etcd:2379")
	t.Setenv("PITAYA_NATS_URL", "nats://nats:4222")
	s := Settings{AuthJWTSecret: "short", EnvName: "development"}
	err := s.Validate(true)
	if err == nil || !strings.Contains(err.Error(), "at least") {
		t.Fatalf("want length error, got: %v", err)
	}
}
