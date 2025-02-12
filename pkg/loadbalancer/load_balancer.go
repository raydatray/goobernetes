package loadbalancer

type LoadBalancer interface {
	AddServer(server *ServerInstance) error
	RemoveServer(serverID string) error
	NextServer() (*ServerInstance, error)
	SetServerStatus(serverID string, active bool) error
	GetServers() []*ServerInstance
	// UpdateServerMetrics(serverID string) error
	// HealthCheck() error
}
