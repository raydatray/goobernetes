package tests

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/cucumber/godog"
	"github.com/raydatray/goobernetes/pkg/loadbalancer"
)

type roundRobinTest struct {
	lb        loadbalancer.LoadBalancer
	requests  []*loadbalancer.ServerInstance
	lastError error
}

func (t *roundRobinTest) reset() {
	t.lb = loadbalancer.NewRoundRobinLoadBalancer()
	t.requests = make([]*loadbalancer.ServerInstance, 0)
	t.lastError = nil
}

func (t *roundRobinTest) theLoadBalancerIsRunning() error {
	t.reset()
	return nil
}

func (t *roundRobinTest) theFollowingBackendServersAreConfigured(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		serverID := row.Cells[0].Value
		host := row.Cells[2].Value
		port, _ := strconv.Atoi(row.Cells[3].Value)
		maxConn, _ := strconv.Atoi(row.Cells[4].Value)

		server, err := loadbalancer.NewServerInstance(serverID, host, port, maxConn)
		if err != nil {
			return fmt.Errorf("failed to create server: %v", err)
		}

		if err := t.lb.AddServer(server); err != nil {
			return fmt.Errorf("failed to add server: %v", err)
		}
	}
	return nil
}

func (t *roundRobinTest) allBackendServersAreHealthy() error {
	for _, server := range t.lb.GetServers() {
		if !server.(*loadbalancer.ServerInstance).Active {
			t.lastError = loadbalancer.ErrServerNotAvailable
			return loadbalancer.ErrServerNotAvailable
		}
	}

	return nil
}

func (t *roundRobinTest) aClientMakesConsecutiveRequests(requestCount int) error {
	t.requests = make([]*loadbalancer.ServerInstance, 0)
	for i := 0; i < requestCount; i++ {
		server, err := t.lb.NextServer()
		if err != nil {
			t.lastError = err
			return err
		}
		t.requests = append(t.requests, server.(*loadbalancer.ServerInstance))
	}
	return nil
}

func (t *roundRobinTest) theRequestsShouldBeRoutedInThisOrder(table *godog.Table) error {
	if len(t.requests) != len(table.Rows)-1 {
		return fmt.Errorf("expected %d requests but got %d requests", len(table.Rows)-1, len(t.requests))
	}

	for i, row := range table.Rows[1:] { // skip first row
		expectedServerID := row.Cells[1].Value
		receivedServerID := t.requests[i].ID
		if expectedServerID != receivedServerID {
			return fmt.Errorf("request %d: expected %s but got %s", i+1, expectedServerID, receivedServerID)
		}
	}
	return nil
}

func (t *roundRobinTest) serverBecomesUnavailable(serverID string) error {
	t.lb.SetServerStatus(serverID, false)
	return nil
}

func (t *roundRobinTest) allBackendServersAreUnavailable() error {
	servers := t.lb.GetServers()
	for _, server := range servers {
		if err := t.lb.SetServerStatus(server.(*loadbalancer.ServerInstance).ID, false); err != nil {
			return err
		}
	}
	return nil
}

func (t *roundRobinTest) aClientMakesARequest() error {
	_, err := t.lb.NextServer()
	t.lastError = err
	return nil
}

func (t *roundRobinTest) theLoadBalancerShouldReturnAServiceUnavailableResponse() error {
	if t.lastError != loadbalancer.ErrNoServerAvailable {
		return fmt.Errorf("expected ErrNoServerAvailable but got %v", t.lastError)
	}
	return nil
}

func intializeID001Scenario(ctx *godog.ScenarioContext) {
	test := &roundRobinTest{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		test.reset()
		return ctx, nil
	})

	ctx.Step(`^the load balancer is running$`, test.theLoadBalancerIsRunning)
	ctx.Step(`^the following backend servers are configured:$`, test.theFollowingBackendServersAreConfigured)
	ctx.Step(`^all backend servers are healthy$`, test.allBackendServersAreHealthy)
	ctx.Step(`^a client makes (\d+) consecutive requests$`, test.aClientMakesConsecutiveRequests)
	ctx.Step(`^the requests should be routed in this order:$`, test.theRequestsShouldBeRoutedInThisOrder)
	ctx.Step(`^"([^"]*)" becomes unavailable$`, test.serverBecomesUnavailable)
	ctx.Step(`^all backend servers are unavailable$`, test.allBackendServersAreUnavailable)
	ctx.Step(`^a client makes a request$`, test.aClientMakesARequest)
	ctx.Step(`^the load balancer should return a "503 Service Unavailable" response$`, test.theLoadBalancerShouldReturnAServiceUnavailableResponse)
}

func TestID001(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: intializeID001Scenario,
		Options: &godog.Options{
			Format:    "pretty",
			Paths:     []string{"../features/ID001_Implement_Round_Robin.feature"},
			Randomize: 0,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("ID001 test failure")
	}
}
