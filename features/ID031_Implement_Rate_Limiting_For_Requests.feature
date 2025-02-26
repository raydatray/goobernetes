Feature: Configure Request Rate Limiting
  As a system administrator
  I want to set rate limits for backend requests
  So that I can prevent abuse and ensure fair resource distribution

  Background:
    Given the load balancer service is running
    And I am authenticated as an administrator
    And a rate limiter middleware is configured with the following settings:
      | requests_per_minute | window_size_seconds |
      | 50                 | 60                  |

  Scenario: Main Flow - Requests within rate limit
    When I send 45 requests within a minute
    Then all requests should be successful
    And the rate limiter response should show remaining requests

  Scenario: Main Flow - Update rate limit configuration
    When I update the rate limit settings to:
      | requests_per_minute | window_size_seconds |
      | 100                | 60                  |
    Then the rate limiter should be updated successfully
    And I should be able to send up to 100 requests successfully
    And the next request should be rejected

  Scenario: Alternative Flow - Rate limit exceeded
    When I send 55 requests within a minute
    Then the first 50 requests should be successful
    And the remaining requests should receive a "429 Too Many Requests" response
    And the response should include rate limit information

  Scenario: Error Flow - Invalid rate limit configuration - requests
    When I attempt to set invalid rate limit values:
      | requests_per_minute | window_size_seconds |
      | -50                 | 60                  |
    Then the system should return a "400 Bad Request" response
    And the rate limit settings should remain unchanged

  Scenario: Error Flow - Invalid rate limit configuration - time frame
    When I attempt to set invalid rate limit values:
      | requests_per_minute | window_size_seconds |
      | 50                  | -60                 |
    Then the system should return a "400 Bad Request" response
    And the rate limit settings should remain unchanged
