package loadbalancer

import (
	"github.com/raydatray/goobernetes/pkg/server"
)

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

func (rr *RoundRobinLoadBalancer) NextServer() (*server.ServerInstance, error) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	if len(rr.servers) == 0 {
		return nil, ErrNoServerAvailable
	}

	startIndex := rr.current
	for i := 0; i < len(rr.servers); i++ {
		currentIndex := (startIndex + i) % len(rr.servers)
		if rr.servers[currentIndex].Active {
			rr.current = (currentIndex + 1) % len(rr.servers)
			return rr.servers[currentIndex], nil
		}
	}

	return nil, ErrNoServerAvailable
}
