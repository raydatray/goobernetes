package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/raydatray/goobernetes/pkg/loadbalancer"
)

type testMaxConn struct {
	id        string
	host      string
	port      int
	maxConn   int
	server    *loadbalancer.ServerInstance
	lb        loadbalancer.LoadBalancer
	lastError error
}

func (t *testMaxConn) reset() error {
	t.lb = loadbalancer.NewRoundRobinLoadBalancer()
	t.lastError = nil
	return nil
}

func (t *testMaxConn) theLoadBalancerIsRunning() error {
	t.reset()
	return nil
}

func (t *testMaxConn) aBackendServerExistsWithAddressAndPort(id string, host string, port int) error {
	t.id = id
	t.host = host
	t.port = port
	return nil
}

func (t *testMaxConn) iSetTheMaximumConnectionsForABackendServer(maxConn int) error {
	t.maxConn = maxConn
	server, err := loadbalancer.NewServerInstance(t.id, t.host, t.port, t.maxConn)
	t.server = server
	t.lastError = err
	return err
}

func (t *testMaxConn) theMaximumConnectionsShouldBeUpdatedSuccessfully() error {
	if t.maxConn != t.server.MaxConns {
		return fmt.Errorf("expected %d but got %d", t.maxConn, t.server.MaxConns)
	}
	return nil
}

func (t *testMaxConn) theBackendServerShouldAcceptConnectionsUpToTheLimit() error {
	_ = t.lb.AddServer(t.server)
	for i := 0; i < t.maxConn; i++ {
		_, err := t.lb.NextServer()
		if err != nil {
			return err
		}
	}

	if _, err := t.lb.NextServer(); err != loadbalancer.ErrNoServerAvailable {
		return fmt.Errorf("expected \"no server available\" but got %v", err)
	}

	return nil
}

func (t *testMaxConn) iTryToSetTheMaximumConnectionsForTheBackendServer(maxConn int) error {
	t.maxConn = maxConn
	_, err := loadbalancer.NewServerInstance(t.id, t.host, t.port, t.maxConn)
	t.lastError = err
	return nil
}

func (t *testMaxConn) IShouldReceiveAnErrorMessage(errorMessage string) error {
	if t.lastError != loadbalancer.ErrInvalidMaxConns {
		return fmt.Errorf("expected \"Invalid max connections\" but got %v", t.lastError)
	}
	return nil
}

func (t *testMaxConn) theBackendServerHasAMaximumConnectionLimit(maxConn int) error {
	t.maxConn = maxConn
	server, err := loadbalancer.NewServerInstance(t.id, t.host, t.port, t.maxConn)
	t.server = server
	t.lastError = err
	return err
}

func (t *testMaxConn) iUpdateTheMaximumConnections(newMaxConn int) error {
	t.maxConn = newMaxConn
	t.lb.NextServer() // ensure that len(s.connections) is 1
	err := t.lb.UpdateServerMaxConn(t.id, t.maxConn)
	t.lastError = err
	return err
}

func (t *testMaxConn) theMaximumConnectionsShouldBeUpdated() error {
	server := t.lb.GetServers()[0].(*loadbalancer.ServerInstance)
	if t.maxConn != server.MaxConns {
		return fmt.Errorf("expected %d but got %d", t.maxConn, server.MaxConns)
	}
	return nil
}

func (t *testMaxConn) existingConnectionsShouldNotBeAffected() error {
	server := t.lb.GetServers()[0].(*loadbalancer.ServerInstance)
	if server.GetConnectionAmount() != 1 {
		return fmt.Errorf("expected 1 connection but got %d", server.GetConnectionAmount())
	}
	return nil
}

func initializeID012Scenario(ctx *godog.ScenarioContext) {
	test := &testMaxConn{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		test.reset()
		return ctx, nil
	})

	ctx.Step(`^the load balancer is running$`, test.theLoadBalancerIsRunning)
	ctx.Step(`^a backend server exists with id "([^"]*)", address "([^"]*)" and port (\d+)$`, test.aBackendServerExistsWithAddressAndPort)
	ctx.Step(`^I set the maximum connections to (\d+) for the backend server$`, test.iSetTheMaximumConnectionsForABackendServer)
	ctx.Step(`^the maximum connections should be updated successfully$`, test.theMaximumConnectionsShouldBeUpdatedSuccessfully)
	ctx.Step(`^the backend server should accept connections up to the limit$`, test.theBackendServerShouldAcceptConnectionsUpToTheLimit)
	ctx.Step(`^I try to set the maximum connections to (-\d+) for the backend server`, test.iTryToSetTheMaximumConnectionsForTheBackendServer)
	ctx.Step(`^I should receive an error message "([^"]*)"$`, test.IShouldReceiveAnErrorMessage)
	ctx.Step(`^the backend server has a maximum connection limit of (\d+)$`, test.theBackendServerHasAMaximumConnectionLimit)
	ctx.Step(`^I update the maximum connections to (\d+)$`, test.iUpdateTheMaximumConnections)
	ctx.Step(`^the maximum connections should be updated to 1500$`, test.theMaximumConnectionsShouldBeUpdated)
	ctx.Step(`^existing connections should not be affected`, test.existingConnectionsShouldNotBeAffected)
}

func TestID012(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeID012Scenario,
		Options: &godog.Options{
			Format:    "pretty",
			Paths:     []string{"../features/ID012_Set_Max_Connections_per_Backend.feature"},
			Randomize: 0,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("ID012 test failure")
	}
}
