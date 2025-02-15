package loadbalancer

type ServerInstance struct {
<<<<<<< HEAD
	ID     string
	Host   string
	Port   string
	Weight int
	Active bool
=======
	ID          string
	Host        string
	Port        int
	Active      bool
	MaxConns    int
	connections chan struct{}
}

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
>>>>>>> 06a8271 (add max connections per backend)
}
