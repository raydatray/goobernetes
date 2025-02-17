package loadbalancer

import "fmt"

type WeightedRoundRobinLoadBalancer struct {
	BaseLoadBalancer
	current   int
	delivered int
}

var _ LoadBalancer = (*WeightedRoundRobinLoadBalancer)(nil) // Compile time interface check

func NewWeightedRoundRobinLoadBalancer() LoadBalancer {
	return &WeightedRoundRobinLoadBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(),
		current:          0,
		delivered:        0,
	}
}

func (wrr *WeightedRoundRobinLoadBalancer) NextServer() (*ServerInstance, error) {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	if len(wrr.servers) == 0 {
		return nil, ErrNoServerAvailable
	}

	for i := 0; i < len(wrr.servers); i++ {
		server := wrr.servers[wrr.current]
		fmt.Printf("%#v\n", server)
		if server.Active {
			if wrr.delivered < server.Weight {
				wrr.delivered++
				return server, nil
			}
			wrr.delivered = 0
		}
		wrr.current = (wrr.current + 1) % len(wrr.servers)
	}

	return nil, ErrNoServerAvailable
}
