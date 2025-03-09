package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type ServerHealth struct {
	LastCheck time.Time
	LatencyNs int64
	IsHealthy bool
}

type HealthChecker interface {
	Start()
	Stop()
	GetServerHealth(serverID string) (*ServerHealth, bool)
	GetAllHealth() map[string]*ServerHealth
}

var _ HealthChecker = (*healthChecker)(nil)

type healthChecker struct {
	lb           LoadBalancer
	healthStatus map[string]*ServerHealth
	interval     time.Duration
	timeout      time.Duration
	client       *http.Client
	stop         chan struct{}
	*sync.RWMutex
}

func NewHealthChecker(lb LoadBalancer, interval, timeout time.Duration) HealthChecker {
	return &healthChecker{
		lb:           lb,
		healthStatus: make(map[string]*ServerHealth),
		interval:     interval,
		timeout:      timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		stop:    make(chan struct{}),
		RWMutex: &sync.RWMutex{},
	}
}

func (hc *healthChecker) Start() {
	go hc.checkLoop()
}

func (hc *healthChecker) Stop() {
	close(hc.stop)
}

func (hc *healthChecker) GetServerHealth(serverID string) (*ServerHealth, bool) {
	hc.RLock()
	defer hc.RUnlock()

	health, exists := hc.healthStatus[serverID]
	return health, exists
}

func (hc *healthChecker) GetAllHealth() map[string]*ServerHealth {
	hc.RLock()
	defer hc.RUnlock()

	result := make(map[string]*ServerHealth)
	for k, v := range hc.healthStatus {
		result[k] = v
	}
	return result
}

func (hc *healthChecker) checkLoop() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.stop:
			return
		case <-ticker.C:
			hc.checkAll()
		}
	}
}

func (hc *healthChecker) checkAll() {
	servers := hc.lb.GetServers()
	var wg sync.WaitGroup

	for _, server := range servers {
		wg.Add(1)

		go func(srv Server) {
			defer wg.Done()
			hc.checkServer(srv.(*ServerInstance))
		}(server)
	}

	wg.Wait()
}

func (hc *healthChecker) checkServer(srv *ServerInstance) {
	start := time.Now()
	url := fmt.Sprintf("http://%s/health-check", srv.GetHostPort())

	resp, err := hc.client.Get(url)
	latency := time.Since(start).Nanoseconds()

	hc.Lock()
	defer hc.Unlock()

	h := &ServerHealth{
		LastCheck: time.Now(),
		LatencyNs: latency,
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		h.IsHealthy = false
		log.Printf("health check failed for server %s: %v", srv.ID, err)
	}

	hc.healthStatus[srv.ID] = h
	log.Printf("health check for server %s: %d, %v", srv.ID, h.LatencyNs, h.IsHealthy)
}
