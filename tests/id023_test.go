package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/raydatray/goobernetes/pkg/loadbalancer"
)

type connectionPoolTest struct {
	lb           loadbalancer.LoadBalancer
	lastError    error
	poolSize     int
	currentConns int
	maxConn      int
	ctx          context.Context
	serverHit    *loadbalancer.ServerInstance
	connInUse    int
	addConn      bool
}

func (t *connectionPoolTest) reset() {
	t.lb = loadbalancer.NewRoundRobinLoadBalancer()
	t.maxConn = 10
	t.lastError = nil
}

func (t *connectionPoolTest) theLoadBalancerIsRunning() error {
	t.reset()
	return nil
}

func (t *connectionPoolTest) theConnectionPoolSizeIs(size int) error {
	t.poolSize = size
	return nil
}

func (t *connectionPoolTest) connectionsAreEstablished(n int) error {
	t.currentConns = n
	return nil
}

func (t *connectionPoolTest) theMaxConnectionsLimitIs(max int) error {
	t.maxConn = max
	return nil
}

func (t *connectionPoolTest) connectionsAreInUse(inUse int) error {
	t.connInUse = inUse
	t.addConn = true
	return nil
}

func (t *connectionPoolTest) addServerToLB() {
	s, _ := loadbalancer.NewServerInstance("server1", "127.0.0.1", 8080, t.maxConn, t.poolSize)
	if t.addConn {
		for range t.connInUse {
			s.AcquireConnection()
		}
	}
	t.lb.AddServer(s)
}

func (t *connectionPoolTest) aClientSendsARequest() error {
	t.addServerToLB()
	s, err := t.lb.NextServer(t.ctx)
	t.lastError = err
	srv, ok := s.(*loadbalancer.ServerInstance)
	if ok {
		t.serverHit = srv
	}
	return nil
}

func (t *connectionPoolTest) theRequestShouldUseAnExistingConnection() error {
	if t.serverHit.GetConnectionAmount() != 5 && len(t.serverHit.ActiveConns) != 1 {
		return fmt.Errorf("Expected 1 active connection and 5 open connections but got %d active and %d open", len(t.serverHit.ActiveConns), t.serverHit.GetConnectionAmount())
	}
	return nil
}

func (t *connectionPoolTest) theConnectionShouldBeReturned() error {
	t.serverHit.ReleaseConnection()
	if t.serverHit.GetConnectionAmount() != 5 && len(t.serverHit.ActiveConns) != 0 {
		return fmt.Errorf("Expected 0 active connection and 5 open connections but got %d active and %d open", len(t.serverHit.ActiveConns), t.serverHit.GetConnectionAmount())
	}
	return nil
}

func (t *connectionPoolTest) aNewConnectionShouldBeCreated() error {
	if t.serverHit.GetConnectionAmount() != 3 && len(t.serverHit.ActiveConns) != 3 {
		return fmt.Errorf("Expected 3 active connections and 3 open connections but got %d active and %d open", len(t.serverHit.ActiveConns), t.serverHit.GetConnectionAmount())
	}
	return nil
}

func (t *connectionPoolTest) theConnectionShouldBeClosed() error {
	t.serverHit.ReleaseConnection()
	if t.serverHit.GetConnectionAmount() != 2 && len(t.serverHit.ActiveConns) != 2 {
		return fmt.Errorf("Expected 2 active connections and 2 open connections but got %d active and %d open", len(t.serverHit.ActiveConns), t.serverHit.GetConnectionAmount())
	}
	return nil
}

func (t *connectionPoolTest) theRequestShouldFailWith(errMsg string) error {
	if t.lastError == nil {
		return fmt.Errorf("Expect error but got none")
	}

	if t.lastError.Error() != errMsg {
		return fmt.Errorf("expected error '%s' but got '%s'", errMsg, t.lastError.Error())
	}

	return nil
}

func initializeID023Scenario(ctx *godog.ScenarioContext) {
	test := &connectionPoolTest{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		test.reset()
		return ctx, nil
	})

	ctx.Step(`^the load balancer is running$`, test.theLoadBalancerIsRunning)
	ctx.Step(`^the connection pool size is (\d+)$`, test.theConnectionPoolSizeIs)
	ctx.Step(`^(\d+) connections are established$`, test.connectionsAreEstablished)
	ctx.Step(`^a client sends a request$`, test.aClientSendsARequest)
	ctx.Step(`^the request should use an existing connection$`, test.theRequestShouldUseAnExistingConnection)
	ctx.Step(`^the connection should be returned to the pool after use$`, test.theConnectionShouldBeReturned)
	ctx.Step(`^(\d+) connections are in use$`, test.connectionsAreInUse)
	ctx.Step(`^the max connections limit is (\d+)$`, test.theMaxConnectionsLimitIs)
	ctx.Step(`^a new connection should be created$`, test.aNewConnectionShouldBeCreated)
	ctx.Step(`^the connection should be closed after use$`, test.theConnectionShouldBeClosed)
	ctx.Step(`^the request should fail with "([^"]*)" error$`, test.theRequestShouldFailWith)
}

func TestID023(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeID023Scenario,
		Options: &godog.Options{
			Format:    "pretty",
			Paths:     []string{"../features/ID023_Connection_Pooling.feature"},
			Randomize: 0,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("ID023 test failure")
	}
}
