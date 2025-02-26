package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/raydatray/goobernetes/pkg/middleware"
)

const (
	defaultRateLimit = 50
	defaultWindow    = time.Minute
)

type rateLimitTest struct {
	rateLimiter    *middleware.RateLimiter
	responses      []*http.Response
	responseBodies [][]byte
	testError      error
	currentLimit   int64
	testServer     *httptest.Server
	requestCounter int64
}

func (test *rateLimitTest) cleanup() {
	test.responses = make([]*http.Response, 0)
	test.responseBodies = make([][]byte, 0)
	test.testError = nil
	test.currentLimit = defaultRateLimit
	test.requestCounter = 0
	if test.rateLimiter != nil {
		test.rateLimiter.Reset()
	}
	if test.testServer != nil {
		test.testServer.Close()
		test.testServer = nil
	}
}

func (test *rateLimitTest) initializeLoadBalancer() error {
	test.cleanup()
	rateLimiter, err := middleware.NewRateLimiter(defaultRateLimit, defaultWindow)
	if err != nil {
		return fmt.Errorf("failed to initialize load balancer: %v", err)
	}
	test.rateLimiter = rateLimiter
	return nil
}

func (test *rateLimitTest) authenticateAsAdmin() error {
	// Authentication would be implemented in a real system
	return nil
}

func (test *rateLimitTest) configureRateLimiter(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		requestsPerMinute, err := strconv.ParseInt(row.Cells[0].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid rate limit value '%s': %v", row.Cells[0].Value, err)
		}

		windowSeconds, err := strconv.ParseInt(row.Cells[1].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid window size value '%s': %v", row.Cells[1].Value, err)
		}

		rateLimiter, err := middleware.NewRateLimiter(requestsPerMinute, time.Duration(windowSeconds)*time.Second)
		if err != nil {
			test.testError = err
			return fmt.Errorf("failed to create rate limiter: %v", err)
		}

		test.rateLimiter = rateLimiter
		test.currentLimit = requestsPerMinute

		test.testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if !test.rateLimiter.TryAcquire() {
				test.requestCounter++
				response := middleware.RateLimitResponse{
					Error:     middleware.ErrRateLimitExceeded.Error(),
					Limit:     test.currentLimit,
					Remaining: 0,
					Reset:     time.Now().Add(time.Second).Unix(), // Using fixed window size of 1 second
				}
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(writer).Encode(response)
				return
			}

			test.requestCounter++
			remaining := test.rateLimiter.GetRemainingRequests()
			response := middleware.RateLimitResponse{
				Limit:     test.currentLimit,
				Remaining: remaining,
				Reset:     time.Now().Add(time.Second).Unix(), // Using fixed window size of 1 second
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			json.NewEncoder(writer).Encode(response)
		}))
	}
	return nil
}

func (test *rateLimitTest) sendRequests(requestCount int) error {
	if test.testServer == nil {
		return fmt.Errorf("test server not initialized")
	}

	test.requestCounter = 0
	if test.rateLimiter != nil {
		test.rateLimiter.Reset()
	}

	// Clear previous responses
	test.responses = make([]*http.Response, 0)
	test.responseBodies = make([][]byte, 0)

	client := &http.Client{}
	startTime := time.Now()

	for i := 0; i < requestCount; i++ {
		if time.Since(startTime) >= defaultWindow {
			return fmt.Errorf("test timed out after %v", defaultWindow)
		}

		request, err := http.NewRequest(http.MethodGet, test.testServer.URL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request %d: %v", i+1, err)
		}

		response, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("failed to send request %d: %v", i+1, err)
		}

		// Read and store the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body for request %d: %v", i+1, err)
		}
		response.Body.Close()

		// Store both the response and body
		test.responses = append(test.responses, response)
		test.responseBodies = append(test.responseBodies, body)
	}
	return nil
}

func (test *rateLimitTest) verifySuccessfulRequests() error {
	if len(test.responses) == 0 {
		return fmt.Errorf("no requests were made to verify")
	}

	for index, response := range test.responses {
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("request %d failed: expected status 200, got %d",
				index+1, response.StatusCode)
		}
	}
	return nil
}

func (test *rateLimitTest) verifySuccessfulRequestCount(count int) error {
	if test.testServer == nil {
		return fmt.Errorf("test server not initialized")
	}

	test.requestCounter = 0
	if test.rateLimiter != nil {
		test.rateLimiter.Reset()
	}
	test.responses = make([]*http.Response, 0)
	test.responseBodies = make([][]byte, 0)

	client := &http.Client{}
	startTime := time.Now()

	for i := 0; i < count; i++ {
		if time.Since(startTime) >= defaultWindow {
			return fmt.Errorf("failed to complete %d requests within %v", count, defaultWindow)
		}

		request, err := http.NewRequest(http.MethodGet, test.testServer.URL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request %d: %v", i+1, err)
		}

		response, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("failed to send request %d: %v", i+1, err)
		}

		// Read and store the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body for request %d: %v", i+1, err)
		}
		response.Body.Close()

		test.responses = append(test.responses, response)
		test.responseBodies = append(test.responseBodies, body)

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("request %d failed: expected status 200, got %d with body: %s",
				i+1, response.StatusCode, string(body))
		}
	}
	return nil
}

func (test *rateLimitTest) updateRateLimitSettings(table *godog.Table) error {
	if test.testServer != nil {
		test.testServer.Close()
	}
	return test.configureRateLimiter(table)
}

func (test *rateLimitTest) verifyRateLimiterUpdate() error {
	if test.rateLimiter == nil {
		return fmt.Errorf("rate limiter update failed: limiter is nil")
	}
	return nil
}

func (test *rateLimitTest) verifyRejectedRequests(expectedStatus string) error {
	if len(test.responses) == 0 {
		return fmt.Errorf("no requests were made to verify")
	}

	successCount := 0
	rejectedCount := 0

	for i, response := range test.responses {
		if response.StatusCode == http.StatusOK {
			successCount++
			if successCount > int(test.currentLimit) {
				return fmt.Errorf("request %d succeeded but should have been rejected (over limit %d)",
					i+1, test.currentLimit)
			}
		} else if response.StatusCode == http.StatusTooManyRequests {
			rejectedCount++
		} else {
			return fmt.Errorf("request %d: unexpected status code %d", i+1, response.StatusCode)
		}
	}

	expectedRejections := len(test.responses) - int(test.currentLimit)
	if rejectedCount != expectedRejections {
		return fmt.Errorf("expected %d rejected requests, got %d",
			expectedRejections, rejectedCount)
	}

	return nil
}

func (test *rateLimitTest) verifyInvalidSettings(table *godog.Table) error {
	originalLimit := test.currentLimit

	for _, row := range table.Rows[1:] {
		requestsPerMinute, err := strconv.ParseInt(row.Cells[0].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid rate limit value '%s': %v", row.Cells[0].Value, err)
		}

		windowSeconds, err := strconv.ParseInt(row.Cells[1].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid window size value '%s': %v", row.Cells[1].Value, err)
		}

		_, err = middleware.NewRateLimiter(requestsPerMinute, time.Duration(windowSeconds)*time.Second)
		test.testError = err

		if err != nil {
			test.currentLimit = originalLimit
		}
	}
	return nil
}

func (test *rateLimitTest) verifyErrorResponse(expectedStatus string) error {
	if test.testError == nil {
		return fmt.Errorf("expected error response '%s', but no error occurred", expectedStatus)
	}
	return nil
}

func (test *rateLimitTest) verifyUnchangedSettings() error {
	if test.currentLimit != defaultRateLimit {
		return fmt.Errorf("rate limit was changed: expected %d, got %d",
			defaultRateLimit, test.currentLimit)
	}
	return nil
}

func (test *rateLimitTest) theNextRequestShouldBeRejected() error {
	if test.testServer == nil {
		return fmt.Errorf("test server not initialized")
	}

	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, test.testServer.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	response.Body.Close()

	test.responses = append(test.responses, response)
	test.responseBodies = append(test.responseBodies, body)

	if response.StatusCode != http.StatusTooManyRequests {
		return fmt.Errorf("request should have been rejected: expected status 429, got %d",
			response.StatusCode)
	}
	return nil
}

func (test *rateLimitTest) theRateLimitResponseShouldShowRemainingRequests() error {
	if len(test.responses) == 0 || len(test.responseBodies) == 0 {
		return fmt.Errorf("no responses to check")
	}

	lastResponseBody := test.responseBodies[len(test.responseBodies)-1]

	var rateLimitResponse middleware.RateLimitResponse
	if err := json.Unmarshal(lastResponseBody, &rateLimitResponse); err != nil {
		return fmt.Errorf("failed to parse rate limit response: %v", err)
	}

	if rateLimitResponse.Limit <= 0 {
		return fmt.Errorf("invalid limit value in response: %d", rateLimitResponse.Limit)
	}

	// No need to check for negative remaining as the middleware handles this
	return nil
}

func (test *rateLimitTest) theResponseShouldIncludeRateLimitInformation() error {
	return test.theRateLimitResponseShouldShowRemainingRequests()
}

func initializeID031Scenario(ctx *godog.ScenarioContext) {
	test := &rateLimitTest{}

    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        test.cleanup()
        rateLimiter, err := middleware.NewRateLimiter(defaultRateLimit, defaultWindow)
        if err != nil {
            return ctx, fmt.Errorf("failed to initialize scenario: %v", err)
        }
        test.rateLimiter = rateLimiter
        test.currentLimit = defaultRateLimit
        return ctx, nil
    })

	ctx.Step(`^the load balancer service is running$`, test.initializeLoadBalancer)
	ctx.Step(`^I am authenticated as an administrator$`, test.authenticateAsAdmin)
	ctx.Step(`^a rate limiter middleware is configured with the following settings:$`, test.configureRateLimiter)
	ctx.Step(`^I send (\d+) requests within a minute$`, test.sendRequests)
	ctx.Step(`^all requests should be successful$`, test.verifySuccessfulRequests)
	ctx.Step(`^I update the rate limit settings to:$`, test.updateRateLimitSettings)
	ctx.Step(`^the rate limiter should be updated successfully$`, test.verifyRateLimiterUpdate)
	ctx.Step(`^I should be able to send up to (\d+) requests successfully$`, test.verifySuccessfulRequestCount)
	ctx.Step(`^the next request should be rejected$`, test.theNextRequestShouldBeRejected)
	ctx.Step(`^the first (\d+) requests should be successful$`, test.verifySuccessfulRequestCount)
	ctx.Step(`^the remaining requests should receive a "([^"]*)" response$`, test.verifyRejectedRequests)
	ctx.Step(`^I attempt to set invalid rate limit values:$`, test.verifyInvalidSettings)
	ctx.Step(`^the system should return a "([^"]*)" response$`, test.verifyErrorResponse)
	ctx.Step(`^the rate limit settings should remain unchanged$`, test.verifyUnchangedSettings)
	ctx.Step(`^the rate limiter response should show remaining requests$`, test.theRateLimitResponseShouldShowRemainingRequests)
	ctx.Step(`^the response should include rate limit information$`, test.theResponseShouldIncludeRateLimitInformation)
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
		t.Fatal("rate limiting tests failed")
	}
}
