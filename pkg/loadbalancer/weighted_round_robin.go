package loadbalancer

import "fmt"

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

func (wrr *WeightedRoundRobinLoadBalancer) NextServer() (Server, error) {
	wrr.Lock()
	defer wrr.Unlock()

	if len(wrr.servers) == 0 {
		return nil, ErrNoServerAvailable
	}

	for i := 0; i < len(wrr.servers); i++ {
		server := wrr.servers[wrr.current].(*WeightedServerInstance)
		fmt.Printf("%#v\n", server)
		if server.GetActive() {
			if wrr.delivered < server.GetWeight() {
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

func (ws *WeightedServerInstance) GetWeight() int {
	return ws.Weight
}

func NewWeightedServerInstance(id string, host string, port int, maxConns int, weight int) *WeightedServerInstance {
	return &WeightedServerInstance{
		ServerInstance: *NewServerInstance(id, host, port, maxConns),
		Weight:         weight,
	}
}
