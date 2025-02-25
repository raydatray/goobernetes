package loadbalancer

type RoundRobinLoadBalancer struct {
	BaseLoadBalancer
	current int
}

var _ LoadBalancer = (*RoundRobinLoadBalancer)(nil) // Compile time interface check

func NewRoundRobinLoadBalancer() LoadBalancer {
	return &RoundRobinLoadBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(),
		current:          0,
	}
}

func (rr *RoundRobinLoadBalancer) NextServer() (Server, error) {
	rr.Lock()
	defer rr.Unlock()

	if len(rr.servers) == 0 {
		return nil, ErrNoServerAvailable
	}

	startIndex := rr.current
	for i := 0; i < len(rr.servers); i++ {
		currentIndex := (startIndex + i) % len(rr.servers)
		server, _ := rr.servers[currentIndex].(*ServerInstance)
		if server.Active && server.AcquireConnection() {
			rr.current = (currentIndex + 1) % len(rr.servers)
			return server, nil
		}
	}

	return nil, ErrNoServerAvailable
}
