Feature: Load Balancer Timeout and Retry Handling

  As a system administrator,
  I want to configure timeouts and retries for backend requests,
  So that I can optimize performance and prevent excessive delays.

  Scenario: Request succeeds within timeout
    Given the load balancer is running
    And the following backend servers are registered:
      | server  | response_time | timeout | retry |
      | server1 | 500           | 1000    | 1     |
    When a client sends a request to "server1"
    Then the request should succeed

  Scenario: Request times out and fails
    Given the load balancer is running
    And the following backend servers are registered:
      | server  | response_time | timeout | retry |
      | server1 | 2000          | 1000    | 0     |
    When a client sends a request to "server1"
    Then the request should fail with a "timeout" error

  Scenario: Request is retried after timeout
    Given the load balancer is running
    And the following backend servers are registered:
      | server  | response_time | timeout | retry |
      | server1 | 2000          | 1000    | 1     |
      | server2 | 500           | 1000    | 1     |
    When a client sends a request to "server1"
    Then the request should be retried on "server2"
    And the request should succeed