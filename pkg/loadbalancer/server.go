package loadbalancer

type Server interface {
	GetID() string
	GetHost() string
	GetPort() int
	GetActive() bool
	GetMaxConns() int
	GetConnections() chan struct{}
	AcquireConnection() bool
	ReleaseConnection()
}

type ServerInstance struct {
	ID          string
	Host        string
	Port        int
	Active      bool
	MaxConns    int
	connections chan struct{}
}

var _ Server = (*ServerInstance)(nil)

func NewServerInstance(id string, host string, port int, maxConns int) *ServerInstance {
	return &ServerInstance{
		ID:          id,
		Host:        host,
		Port:        port,
		Active:      true,
		MaxConns:    maxConns,
		connections: make(chan struct{}, maxConns),
	}
}

func (s *ServerInstance) GetID() string {
	return s.ID
}

func (s *ServerInstance) GetHost() string {
	return s.Host
}

func (s *ServerInstance) GetPort() int {
	return s.Port
}

func (s *ServerInstance) GetActive() bool {
	return s.Active
}

func (s *ServerInstance) GetMaxConns() int {
	return s.MaxConns
}

func (s *ServerInstance) GetConnections() chan struct{} {
	return s.connections
}

func (s *ServerInstance) AcquireConnection() bool {
	success := false
	select {
	case s.connections <- struct{}{}:
		success = true
	default:
		success = false
	}
	return success
}

func (s *ServerInstance) ReleaseConnection() {
	select {
	case <-s.connections:
	default:
	}
}
