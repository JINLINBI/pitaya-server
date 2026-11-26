package serverwire

import (
	"testing"

	"github.com/topfreegames/pitaya/v3/pkg/config"
)

func TestNewGroupServiceMemory(t *testing.T) {
	cfg := config.NewDefaultPitayaConfig()
	gs, err := NewGroupService(false, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if gs == nil {
		t.Fatal("want memory group service")
	}
}

func TestNewGroupServiceClusterRequiresEtcd(t *testing.T) {
	cfg := config.NewDefaultPitayaConfig()
	cfg.Groups.Etcd.Endpoints = nil
	if _, err := NewGroupService(true, cfg); err == nil {
		t.Fatal("expected error without etcd endpoints")
	}
}