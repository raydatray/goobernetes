package loadbalancer

import (
	"errors"
	"sync"

	"github.com/raydatray/goobernetes/pkg/server"
)

var (
	ErrServerAlreadyExists = errors.New("server alrady exists")
	ErrServerNotFound      = errors.New("server not found")
	ErrNoServerAvailable   = errors.New("no server available")
)

type BaseLoadBalancer struct {
	servers []*server.ServerInstance
	mutex   sync.RWMutex
}

func NewBaseLoadBalancer() BaseLoadBalancer {
	return BaseLoadBalancer{
		servers: make([]*server.ServerInstance, 0),
	}
}

func (b *BaseLoadBalancer) AddServer(server *server.ServerInstance) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, s := range b.servers {
		if s.ID == server.ID {
			return ErrServerAlreadyExists
		}
	}

	b.servers = append(b.servers, server)
	return nil
}

func (b *BaseLoadBalancer) RemoveServer(serverID string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for i, server := range b.servers {
		if server.ID == serverID {
			b.servers = append(b.servers[:i], b.servers[i+1:]...)
			return nil
		}
	}

	return ErrServerNotFound
}

func (b *BaseLoadBalancer) GetServers() []*server.ServerInstance {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	serversCopy := make([]*server.ServerInstance, len(b.servers))
	copy(serversCopy, b.servers)
	return serversCopy
}

func (b *BaseLoadBalancer) SetServerStatus(serverID string, active bool) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, server := range b.servers {
		if server.ID == serverID {
			server.Active = active
			return nil
		}
	}
	return ErrServerNotFound
}
