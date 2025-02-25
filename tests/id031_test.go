package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/raydatray/goobernetes/pkg/middleware"
)

type rateLimitTest struct {
	rateLimiter  *middleware.RateLimiter
	responses    []*http.Response
	lastError    error
	currentLimit int64
	testServer   *httptest.Server
	requestCount int64
}

func (t *rateLimitTest) reset() {
	t.responses = make([]*http.Response, 0)
	t.lastError = nil
	t.currentLimit = 50 // default limit
	t.requestCount = 0
	if t.testServer != nil {
		t.testServer.Close()
	}
}

func (t *rateLimitTest) theLoadBalancerServiceIsRunning() error {
	t.reset()
	return nil
}

func (t *rateLimitTest) iAmAuthenticatedAsAnAdministrator() error {
	return nil
}

func (t *rateLimitTest) setupTestServer() {
	var err error
	t.rateLimiter, err = middleware.NewRateLimiter(t.currentLimit, time.Minute)
	if err != nil {
		panic(fmt.Sprintf("Failed to create rate limiter: %v", err))
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !t.rateLimiter.TryAcquire() {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", t.currentLimit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		t.requestCount++
		remaining := t.currentLimit - t.requestCount

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", t.currentLimit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.WriteHeader(http.StatusOK)
	})

	t.testServer = httptest.NewServer(handler)
}

func (t *rateLimitTest) aRateLimiterMiddlewareIsConfiguredWithSettings(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		rpm, _ := strconv.ParseInt(row.Cells[0].Value, 10, 64)
		window, _ := strconv.ParseInt(row.Cells[1].Value, 10, 64)

		var err error
		t.rateLimiter, err = middleware.NewRateLimiter(rpm, time.Duration(window)*time.Second)
		if err != nil {
			return err
		}
		t.currentLimit = rpm
	}
	t.setupTestServer()
	return nil
}

func (t *rateLimitTest) iSendRequestsWithinAMinute(count int) error {
	client := &http.Client{}
	for i := 0; i < count; i++ {
		req, _ := http.NewRequest("GET", t.testServer.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		t.responses = append(t.responses, resp)
	}
	return nil
}

func (t *rateLimitTest) allRequestsShouldBeSuccessful() error {
	for i, resp := range t.responses {
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request %d failed with status code %d", i+1, resp.StatusCode)
		}
	}
	return nil
}

func (t *rateLimitTest) theRateLimiterHeadersShouldShowRemainingRequests() error {
	lastResp := t.responses[len(t.responses)-1]
	if lastResp.Header.Get("X-RateLimit-Remaining") == "" {
		return fmt.Errorf("rate limit remaining header not found")
	}
	return nil
}

func (t *rateLimitTest) iUpdateTheRateLimitSettingsTo(table *godog.Table) error {
	if t.testServer != nil {
		t.testServer.Close()
	}
	return t.aRateLimiterMiddlewareIsConfiguredWithSettings(table)
}

func (t *rateLimitTest) theRateLimiterShouldBeUpdatedSuccessfully() error {
	if t.rateLimiter == nil {
		return fmt.Errorf("rate limiter was not updated")
	}
	return nil
}

func (t *rateLimitTest) iShouldBeAbleToSendRequestsSuccessfully(count int) error {
	return t.iSendRequestsWithinAMinute(count)
}

func (t *rateLimitTest) theStRequestShouldBeRejected(count int) error {
	if t.rateLimiter != nil {
		t.rateLimiter.Reset()
	}
	t.requestCount = 0

	// First send count-1 requests
	err := t.iSendRequestsWithinAMinute(count - 1)
	if err != nil {
		return err
	}

	// Then send one more request that should be rejected
	client := &http.Client{}
	req, _ := http.NewRequest("GET", t.testServer.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		return fmt.Errorf("expected status 429, got %d", resp.StatusCode)
	}
	return nil
}

func (t *rateLimitTest) theFirstRequestsShouldBeSuccessful(count int) error {
	for i := 0; i < count && i < len(t.responses); i++ {
		if t.responses[i].StatusCode != http.StatusOK {
			return fmt.Errorf("request %d failed with status code %d", i+1, t.responses[i].StatusCode)
		}
	}
	return nil
}

func (t *rateLimitTest) theRemainingRequestsShouldReceiveAResponse(expectedStatus string) error {
	for i := int(t.currentLimit); i < len(t.responses); i++ {
		if t.responses[i].StatusCode != http.StatusTooManyRequests {
			return fmt.Errorf("expected status %s, got %d", expectedStatus, t.responses[i].StatusCode)
		}
	}
	return nil
}

func (t *rateLimitTest) theResponseShouldIncludeRateLimitHeaders() error {
	return t.theRateLimiterHeadersShouldShowRemainingRequests()
}

func (t *rateLimitTest) iAttemptToSetInvalidRateLimitValues(table *godog.Table) error {
	t.lastError = t.aRateLimiterMiddlewareIsConfiguredWithSettings(table)
	return nil
}

func (t *rateLimitTest) theSystemShouldReturnAResponse(expectedStatus string) error {
	if t.lastError == nil {
		return fmt.Errorf("expected error, got nil")
	}
	return nil
}

func (t *rateLimitTest) theRateLimitSettingsShouldRemainUnchanged() error {
	if t.currentLimit != 50 {
		return fmt.Errorf("rate limit changed to %d, expected 50", t.currentLimit)
	}
	return nil
}

func initializeID031Scenario(ctx *godog.ScenarioContext) {
	test := &rateLimitTest{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		test.reset()
		return ctx, nil
	})

	ctx.Step(`^the load balancer service is running$`, test.theLoadBalancerServiceIsRunning)
	ctx.Step(`^I am authenticated as an administrator$`, test.iAmAuthenticatedAsAnAdministrator)
	ctx.Step(`^a rate limiter middleware is configured with the following settings:$`, test.aRateLimiterMiddlewareIsConfiguredWithSettings)
	ctx.Step(`^I send (\d+) requests within a minute$`, test.iSendRequestsWithinAMinute)
	ctx.Step(`^all requests should be successful$`, test.allRequestsShouldBeSuccessful)
	ctx.Step(`^the rate limiter headers should show remaining requests$`, test.theRateLimiterHeadersShouldShowRemainingRequests)
	ctx.Step(`^I update the rate limit settings to:$`, test.iUpdateTheRateLimitSettingsTo)
	ctx.Step(`^the rate limiter should be updated successfully$`, test.theRateLimiterShouldBeUpdatedSuccessfully)
	ctx.Step(`^I should be able to send up to (\d+) requests successfully$`, test.iShouldBeAbleToSendRequestsSuccessfully)
	ctx.Step(`^the next request should be rejected$`, test.theStRequestShouldBeRejected)
	ctx.Step(`^the first (\d+) requests should be successful$`, test.theFirstRequestsShouldBeSuccessful)
	ctx.Step(`^the remaining requests should receive a "([^"]*)" response$`, test.theRemainingRequestsShouldReceiveAResponse)
	ctx.Step(`^the response should include rate limit headers$`, test.theResponseShouldIncludeRateLimitHeaders)
	ctx.Step(`^I attempt to set invalid rate limit values:$`, test.iAttemptToSetInvalidRateLimitValues)
	ctx.Step(`^the system should return a "([^"]*)" response$`, test.theSystemShouldReturnAResponse)
	ctx.Step(`^the rate limit settings should remain unchanged$`, test.theRateLimitSettingsShouldRemainUnchanged)
}

func TestID031(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeID031Scenario,
		Options: &godog.Options{
			Format:    "pretty",
			Paths:     []string{"../features/ID031_Implement_Rate_Limiting_For_Requests.feature"},
			Randomize: 0,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("ID031 test failure")
	}
}
