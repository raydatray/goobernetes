package tests

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
	"github.com/raydatray/goobernetes/pkg/loadbalancer"
)

type userInputValidationTest struct {
	lb        loadbalancer.LoadBalancer
	lastError error
}

func (t *userInputValidationTest) reset() {
	t.lastError = nil
}

func (t *userInputValidationTest) theLoadBalancerServiceIsRunning() error {
	t.reset()
	return nil
}

func (t *userInputValidationTest) iAmAuthenticatedAsAnAdministrator() error {
	return nil
}

func (t *userInputValidationTest) theBackendServerPoolIsAccessible() error {
	return nil
}

func (t *userInputValidationTest) theValidationServiceIsAvailable() error {
	return nil
}

func (t *userInputValidationTest) iAmConfiguringANewBackendServer() error {
	return nil
}

func (t *userInputValidationTest) iEnterTheFollowingServerDetails(table *godog.Table) error {
	return nil
}

func (t *userInputValidationTest) theSystemShouldValidate() error {
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
	ctx.Step(`^the validation service is available$`, test.theValidationServiceIsAvailable)
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
