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
	ErrInvalidConnPoolSize     = errors.New("Invalid connection pool size")
	ErrConnectionExhausted     = errors.New("connection pool exhausted")

	serverNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
)

type Server interface {
	GetHostPort() string
	AcquireConnection() bool
	ReleaseConnection()
}

type ServerInstance struct {
	ID           string
	Host         string
	Port         int
	Active       bool
	MaxConns     int
	ConnPoolSize int
	ActiveConns  chan struct{}
	connections  chan struct{}
}

var _ Server = (*ServerInstance)(nil)

func NewServerInstance(id string, host string, port int, maxConns int, connPoolSize int) (*ServerInstance, error) {
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

	if connPoolSize < 1 && connPoolSize > maxConns {
		return nil, fmt.Errorf("%w: %d", ErrInvalidConnPoolSize, connPoolSize)
	}

	s := &ServerInstance{
		ID:           id,
		Host:         host,
		Port:         port,
		Active:       true,
		MaxConns:     maxConns,
		ConnPoolSize: connPoolSize,
		ActiveConns:  make(chan struct{}, maxConns),
		connections:  make(chan struct{}, maxConns),
	}

	for range s.ConnPoolSize {
		s.connections <- struct{}{}
	}

	return s, nil
}

func (s *ServerInstance) GetHostPort() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *ServerInstance) AcquireConnection() bool {
	if len(s.connections) > len(s.ActiveConns) {
		s.ActiveConns <- struct{}{}
		return true
	}
	success := false
	select {
	case s.connections <- struct{}{}:
		s.ActiveConns <- struct{}{}
		success = true
	default:
		success = false
	}
	return success
}

func (s *ServerInstance) ReleaseConnection() {
	if len(s.ActiveConns) <= s.ConnPoolSize {
		<-s.ActiveConns
		return
	}
	select {
	case <-s.connections:
		<-s.ActiveConns
	default:
	}
}

func (s *ServerInstance) GetConnectionAmount() int {
	return len(s.connections)
}
