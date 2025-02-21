package loadbalancer

import (
	"errors"
	"fmt"
)

var ErrInvalidWeight = errors.New("Invalid weight (must be between 1-100 inclusive)")

type WeightedServerInstance struct {
	ServerInstance
	Weight int
}

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

func (wrr *WeightedRoundRobinLoadBalancer) AddServer(server Server) error {
	wrr.Lock()
	defer wrr.Unlock()

	srv, ok := server.(*WeightedServerInstance)
	if !ok {
		return ErrBadServerInterface
	}

	for _, s := range wrr.servers {
		if s.(*WeightedServerInstance).ID == srv.ID {
			return ErrServerAlreadyExists
		}
	}

	wrr.servers = append(wrr.servers, srv)
	return nil
}

func (wrr *WeightedRoundRobinLoadBalancer) GetServers() []Server {
	wrr.RLock()
	defer wrr.RUnlock()

	serversCopy := make([]Server, 0, len(wrr.servers))
	for _, s := range wrr.servers {
		serversCopy = append(serversCopy, s.(*WeightedServerInstance))
	}

	return serversCopy
}

func (wrr *WeightedRoundRobinLoadBalancer) SetServerStatus(serverID string, active bool) error {
	wrr.Lock()
	defer wrr.Unlock()

	for _, s := range wrr.servers {
		if s.(*WeightedServerInstance).ID == serverID {
			s.(*WeightedServerInstance).Active = active
			return nil
		}
	}
	return ErrServerNotFound
}

func (wrr *WeightedRoundRobinLoadBalancer) NextServer() (Server, error) {
	wrr.Lock()
	defer wrr.Unlock()

	if len(wrr.servers) == 0 {
		return nil, ErrNoServerAvailable
	}

	for i := 0; i < len(wrr.servers); i++ {
		server := wrr.servers[wrr.current].(*WeightedServerInstance)
		fmt.Printf("%#v\n", server)
		if server.Active {
			if wrr.delivered < server.Weight && server.AcquireConnection() {
				wrr.delivered++
				return server, nil
			}
			wrr.delivered = 0
		}
		wrr.current = (wrr.current + 1) % len(wrr.servers)
	}

	return nil, ErrNoServerAvailable
}

var _ Server = (*WeightedServerInstance)(nil)

func NewWeightedServerInstance(id string, host string, port int, maxConns int, weight int) (*WeightedServerInstance, error) {
	ServerInstance, err := NewServerInstance(id, host, port, maxConns)

	if err != nil {
		return nil, err
	}

	if weight < 1 || weight > 100 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidWeight, weight)
	}

	return &WeightedServerInstance{
		ServerInstance: *ServerInstance,
		Weight:         weight,
	}, nil
}
