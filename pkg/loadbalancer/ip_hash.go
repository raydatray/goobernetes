package loadbalancer

import (
	"context"
	"errors"
	"hash/fnv"
)

const ClientIPKey contextKey = "client_ip"

var ErrNoClientIP = errors.New("no client ip found in ctx")

type IPHashLoadBalancer struct {
	BaseLoadBalancer
}

var _ LoadBalancer = (*IPHashLoadBalancer)(nil)

func NewIPHashLoadBalancer() LoadBalancer {
	return &IPHashLoadBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(),
	}
}

func (ip *IPHashLoadBalancer) NextServer(ctx context.Context) (Server, error) {
	ip.Lock()
	defer ip.Unlock()

	if len(ip.servers) == 0 {
		return nil, ErrNoServerAvailable
	}

	clientIP, ok := ctx.Value(ClientIPKey).(string)
	if !ok {
		return nil, ErrNoClientIP
	}

	h := fnv.New32a()
	h.Write([]byte(clientIP))
	hash := h.Sum32()

	selectedServer := ip.servers[hash%uint32(len(ip.servers))].(*ServerInstance)

	if selectedServer.Active && selectedServer.AcquireConnection() {
		return selectedServer, nil
	}
	return nil, ErrServerNotAvailable
}
