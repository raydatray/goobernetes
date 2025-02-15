package loadbalancer

type RoundRobinLoadBalancer struct {
	BaseLoadBalancer
	current int
}

var _ LoadBalancer = (*RoundRobinLoadBalancer)(nil) //Compile time interface check

func NewRoundRobinLoadBalancer() LoadBalancer {
	return &RoundRobinLoadBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(),
		current:          0,
	}
}

func (rr *RoundRobinLoadBalancer) NextServer() (*ServerInstance, error) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	if len(rr.servers) == 0 {
		return nil, ErrNoServerAvailable
	}

	startIndex := rr.current
	for i := 0; i < len(rr.servers); i++ {
		currentIndex := (startIndex + i) % len(rr.servers)
		server := rr.servers[currentIndex]

		if server.Active && server.AcquireConnection() {
			rr.current = (currentIndex + 1) % len(rr.servers)
			return server, nil
		}
	}

	return nil, ErrNoServerAvailable
}
