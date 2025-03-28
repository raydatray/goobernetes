package loadbalancer

import (
	"errors"
	"fmt"
	"net"
	"regexp"
)

var (
	ErrInvalidServerNameLength = errors.New("Invalid server name length")
	ErrInvalidCharInServerName = errors.New("Invalid character in server name")
	ErrInvalidIP               = errors.New("Invalid IP address")
	ErrInvalidPort             = errors.New("Invalid port number")
	ErrInvalidMaxConns         = errors.New("Invalid max connections")

	serverNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
)

type Server interface {
	GetHostPort() string
	AcquireConnection() bool
	ReleaseConnection()
}

type ServerInstance struct {
	ID          string
	Host        string
	Port        int
	Active      bool
	MaxConns    int
	connections chan struct{}
}

var _ Server = (*ServerInstance)(nil)

func NewServerInstance(id string, host string, port int, maxConns int) (*ServerInstance, error) {
	if len(id) < 1 || len(id) > 64 {
		return nil, ErrInvalidServerNameLength
	}

	if !serverNameRegex.MatchString(id) {
		return nil, ErrInvalidCharInServerName
	}

	ip := net.ParseIP(host)

	if ip == nil || ip.IsUnspecified() {
		return nil, fmt.Errorf("%w: %s", ErrInvalidIP, host)
	}

	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidPort, port)
	}

	if maxConns < 1 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidMaxConns, maxConns)
	}

	return &ServerInstance{
		ID:          id,
		Host:        host,
		Port:        port,
		Active:      true,
		MaxConns:    maxConns,
		connections: make(chan struct{}, maxConns),
	}, nil
}

func (s *ServerInstance) GetHostPort() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *ServerInstance) AcquireConnection() bool {
	success := false
	select {
	case s.connections <- struct{}{}:
		success = true
	default:
		success = false
	}
	return success
}

func (s *ServerInstance) ReleaseConnection() {
	select {
	case <-s.connections:
	default:
	}
}

func (s *ServerInstance) GetConnectionAmount() int {
	return len(s.connections)
}
