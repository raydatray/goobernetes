package loadbalancer

type ServerInstance struct {
	ID          string
	Host        string
	Port        int
	Active      bool
	MaxConns    int
	connections chan struct{}
	Weight      int
}

type Option func(*ServerInstance)

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
