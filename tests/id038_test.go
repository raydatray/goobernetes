package tests

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/cucumber/godog"
	"github.com/raydatray/goobernetes/pkg/loadbalancer"
)

type userInputValidationTest struct {
	lb        loadbalancer.LoadBalancer
	server    *loadbalancer.ServerInstance
	lastError error
}

func (t *userInputValidationTest) reset() {
	t.lb = loadbalancer.NewRoundRobinLoadBalancer()
	t.server = nil
	t.lastError = nil
}

func (t *userInputValidationTest) theLoadBalancerServiceIsRunning() error {
	t.reset()
	return nil
}

func (t *userInputValidationTest) iAmAuthenticatedAsAnAdministrator() error {
	// TODO - Implement once authentication & role features are added
	return nil
}

func (t *userInputValidationTest) theBackendServerPoolIsAccessible() error {
	if t.lb == nil {
		return errors.New("Backend server pool is not accessible")
	}

	return nil
}

func (t *userInputValidationTest) iAmConfiguringANewBackendServer() error {
	t.server = nil
	t.lastError = nil

	return nil
}

func (t *userInputValidationTest) iEnterTheFollowingServerDetails(table *godog.Table) error {
	serverDetails := make(map[string]string)

	for _, row := range table.Rows[1:] {
		field := row.Cells[0].Value
		value := row.Cells[1].Value
		serverDetails[field] = value
	}

	// Convert port and weight to integers
	port := 0
	if portStr := serverDetails["Port"]; portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	weight := 0
	if weightStr := serverDetails["Weight"]; weightStr != "" {
		if w, err := strconv.Atoi(weightStr); err == nil {
			weight = w
		}
	}

	server, err := loadbalancer.NewServerInstance(
		serverDetails["Server Name"],
		serverDetails["IP Address"],
		port,
		weight,
	)

	if err != nil {
		t.lastError = err
		return nil
	}

	t.server = server
	return nil
}

func (t *userInputValidationTest) theSystemShouldValidate(table *godog.Table) error {
	if t.lastError != nil {
		switch {
		case errors.Is(t.lastError, loadbalancer.ErrInvalidIP):
			return nil
		case errors.Is(t.lastError, loadbalancer.ErrInvalidPort):
			return nil
		case errors.Is(t.lastError, loadbalancer.ErrInvalidMaxConns):
			return nil
		default:
			return fmt.Errorf("Unexpected validation error: %v", t.lastError)
		}
	}

	if t.server == nil {
		return fmt.Errorf("Unexpected validation error: %v", t.lastError)
	}

	return nil
}

func initializeID038Scenario(ctx *godog.ScenarioContext) {
	test := &userInputValidationTest{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		test.reset()
		return ctx, nil
	})

	ctx.Step(`^the load balancer service is running$`, test.theLoadBalancerServiceIsRunning)
	ctx.Step(`^I am authenticated as an administrator$`, test.iAmAuthenticatedAsAnAdministrator)
	ctx.Step(`^the backend server pool is accessible$`, test.theBackendServerPoolIsAccessible)
	ctx.Step(`^I am configuring a new backend server$`, test.iAmConfiguringANewBackendServer)
	ctx.Step(`^I enter the following server details:$`, test.iEnterTheFollowingServerDetails)
	ctx.Step(`^the system should validate:$`, test.theSystemShouldValidate)
}

func TestID038(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeID038Scenario,
		Options: &godog.Options{
			Format:    "pretty",
			Paths:     []string{"../features/ID038_Provide_Validation_for_User_Input.feature"},
			Randomize: 0,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("ID038 test failure")
	}
}
