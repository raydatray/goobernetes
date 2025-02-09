package tests

import (
	"errors"
	"testing"

	"github.com/raydatray/goobernetes/pkg/loadbalancer"
	"github.com/raydatray/goobernetes/pkg/server"
)

func setupLoadBalancer() loadbalancer.LoadBalancer {
	lb := loadbalancer.NewRoundRobinLoadBalancer()
	servers := []*server.ServerInstance{
		{ID: "server1", Host: "192.168.1.10", Port: "8080", Active: true},
		{ID: "server2", Host: "192.168.1.11", Port: "8080", Active: true},
		{ID: "server3", Host: "192.168.1.12", Port: "8080", Active: true},
	}
	for _, server := range servers {
		_ = lb.AddServer(server)
	}
	return lb
}

func TestRoundRobinNormalFlow(t *testing.T) {
	lb := setupLoadBalancer()
	expectedOrder := []string{"server1", "server2", "server3", "server1", "server2", "server3"}
	for i, expected := range expectedOrder {
		server, err := lb.NextServer()
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
		if server.ID != expected {
			t.Errorf("request %d: expected %s, got %s", i+1, expected, server.ID)
		}
	}
}

func TestRoundRobinWithOneServerDown(t *testing.T) {
	lb := setupLoadBalancer()
	_ = lb.SetServerStatus("server2", false)
	expectedOrder := []string{"server1", "server3", "server1", "server3", "server1", "server3"}
	for i, expected := range expectedOrder {
		server, err := lb.NextServer()
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
		if server.ID != expected {
			t.Errorf("request %d: expected %s, got %s", i+1, expected, server.ID)
		}
	}
}

func TestRoundRobinNoServersAvailable(t *testing.T) {
	lb := setupLoadBalancer()
	for _, server := range lb.GetServers() {
		_ = lb.SetServerStatus(server.ID, false)
	}
	_, err := lb.NextServer()
	if !errors.Is(err, loadbalancer.ErrNoServerAvailable) {
		t.Errorf("expected ErrNoServerAvailable, got %v", err)
	}
}
