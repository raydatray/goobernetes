package loadbalancer

import "github.com/raydatray/goobernetes/pkg/server"

type LoadBalancer interface {
	AddServer(server *server.ServerInstance) error
	RemoveServer(serverID string) error
	NextServer() (*server.ServerInstance, error)
	SetServerStatus(serverID string, active bool) error
	GetServers() []*server.ServerInstance
	// UpdateServerMetrics(serverID string) error
	// HealthCheck() error
}
