package loadbalancer

type LoadBalancer interface {
	AddServer(server Server) error
	RemoveServer(serverID string) error
	NextServer() (Server, error)
	SetServerStatus(serverID string, active bool) error
	GetServers() []Server
	// UpdateServerMetrics(serverID string) error
	// HealthCheck() error
}
