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
	server    *loadbalancer.WeightedServerInstance
	lastError error
	expected  string
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

	port, _ := strconv.Atoi(serverDetails["Port"])
	max_connections, _ := strconv.Atoi(serverDetails["Max Connections"])
	weight, _ := strconv.Atoi(serverDetails["Weight"])

	server, err := loadbalancer.NewWeightedServerInstance(
		serverDetails["Server Name"],
		serverDetails["IP Address"],
		port,
		max_connections,
		weight,
		0,
	)

	t.server = server
	t.lastError = err
	t.expected = serverDetails["Expected"]
	return nil
}

func (t *userInputValidationTest) theSystemShouldValidate(table *godog.Table) error {
	if t.expected == "Success" {
		if t.lastError != nil {
			return fmt.Errorf("Expected Success but got error: %v", t.lastError)
		}
		if t.server == nil {
			return fmt.Errorf("Expected Success but server is nil")
		}
		return nil
	}

	if t.lastError == nil {
		return fmt.Errorf("Expected error '%s' but got no error", t.expected)
	}

	actualError := t.lastError.Error()
	if actualError != t.expected {
		return fmt.Errorf("Expected '%s', but got '%s'", t.expected, actualError)
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
