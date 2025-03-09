package loadbalancer

import (
	"context"
	"math/rand/v2"
)

type RandomLoadBalancer struct {
	BaseLoadBalancer
	attempts int
	random   rand.Rand
}

var _ LoadBalancer = (*RandomLoadBalancer)(nil)

func NewRandomLoadBalancer() LoadBalancer {
	return &RandomLoadBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(),
		attempts:         10,
		random:           *rand.New(rand.NewPCG(1, 2)),
	}
}

func (r *RandomLoadBalancer) NextServer(ctx context.Context) (Server, error) {
	r.Lock()
	defer r.Unlock()

	if len(r.servers) == 0 {
		return nil, ErrNoServerAvailable
	}

	for i := 0; i < r.attempts; i++ {
		selectedServer := r.servers[r.random.IntN(len(r.servers))].(*ServerInstance)

		if selectedServer.Active && selectedServer.AcquireConnection() {
			return selectedServer, nil
		}
	}
	return nil, ErrServerNotAvailable
}
