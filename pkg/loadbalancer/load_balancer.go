package loadbalancer

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type LoadBalancer interface {
	AddServer(server Server) error
	RemoveServer(serverID string) error
	NextServer(ctx context.Context) (Server, error)
	SetServerStatus(serverID string, active bool) error
	GetServers() []Server
	UpdateServerMaxConn(serverID string, maxConn int) error
	// UpdateServerMetrics(serverID string) error
	// HealthCheck() error
}

var (
	ErrServerAlreadyExists = errors.New("server alrady exists")
	ErrServerNotFound      = errors.New("server not found")
	ErrNoServerAvailable   = errors.New("no server available")
	ErrServerNotAvailable  = errors.New("a server is unavailable")
	ErrBadServerInterface  = errors.New("server is not a valid interface")
)

type contextKey string

type BaseLoadBalancer struct {
	servers []Server
	*sync.RWMutex
}

func NewBaseLoadBalancer() BaseLoadBalancer {
	return BaseLoadBalancer{
		servers: make([]Server, 0),
		RWMutex: &sync.RWMutex{},
	}
}

func (b *BaseLoadBalancer) AddServer(srv Server) error {
	b.Lock()
	defer b.Unlock()

	server, ok := srv.(*ServerInstance)
	if !ok {
		return ErrBadServerInterface
	}

	for _, s := range b.servers {
		if s.(*ServerInstance).ID == server.ID {
			return ErrServerAlreadyExists
		}
	}

	b.servers = append(b.servers, server)
	return nil
}

func (b *BaseLoadBalancer) RemoveServer(serverID string) error {
	b.Lock()
	defer b.Unlock()

	for i, s := range b.servers {
		if s.(*ServerInstance).ID == serverID {
			b.servers = append(b.servers[:i], b.servers[i+1:]...)
			return nil
		}
	}

	return ErrServerNotFound
}

func (b *BaseLoadBalancer) GetServers() []Server {
	b.RLock()
	defer b.RUnlock()

	serversCopy := make([]Server, 0, len(b.servers))
	for _, s := range b.servers {
		serversCopy = append(serversCopy, s.(*ServerInstance))
	}

	return serversCopy
}

func (b *BaseLoadBalancer) SetServerStatus(serverID string, active bool) error {
	b.Lock()
	defer b.Unlock()

	for _, s := range b.servers {
		if s.(*ServerInstance).ID == serverID {
			s.(*ServerInstance).Active = active
			return nil
		}
	}
	return ErrServerNotFound
}

func (b *BaseLoadBalancer) UpdateServerMaxConn(serverID string, maxConn int) error {
	b.Lock()
	defer b.Unlock()

	if maxConn < 1 {
		return fmt.Errorf("%w: %d", ErrInvalidMaxConns, maxConn)
	}

	for _, s := range b.servers {
		fmt.Printf("%s, %s\n", s.(*ServerInstance).ID, serverID)
		if s.(*ServerInstance).ID == serverID {
			s.(*ServerInstance).MaxConns = maxConn
			close(s.(*ServerInstance).connections)
			newChan := resizeChannel(s.(*ServerInstance).connections, maxConn)
			s.(*ServerInstance).connections = newChan
			return nil
		}
	}

	return ErrServerNotFound
}

func resizeChannel(oldChan <-chan struct{}, newChanSize int) chan struct{} {
	newChan := make(chan struct{}, newChanSize)

	go func() {
		for val := range oldChan {
			newChan <- val
		}
	}()

	return newChan
}
