package loadbalancer

type LoadBalancer interface {
	AddServer(server Server) error
	RemoveServer(serverID string) error
	NextServer() (Server, error)
	SetServerStatus(serverID string, active bool) error
	GetServers() []Server
	UpdateServerMaxConn(serverID string, maxConn int) error
	// UpdateServerMetrics(serverID string) error
	// HealthCheck() error
}
