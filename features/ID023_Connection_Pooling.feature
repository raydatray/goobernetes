Feature: Connection Pooling in Load Balancer

  As a system administrator,
  I want the load balancer to use connection pooling,
  so that backend resources are efficiently utilized, reducing the overhead of opening and closing connections for each request.

  Scenario: Reuse existing connections from the pool
    Given the load balancer is running
    And the connection pool size is 5
    And 5 connections are established
    When a client sends a request
    Then the request should use an existing connection
    And the connection should be returned to the pool after use

  Scenario: Create a new connection if the pool is empty
    Given the load balancer is running
    And the connection pool size is 2
    And 2 connections are in use
    And the max connections limit is 3
    When a client sends a request
    Then a new connection should be created
    And the connection should be closed after use

  Scenario: Fail request when connection pool is exhausted
    Given the load balancer is running
    And the connection pool size is 2
    And 2 connections are in use
    And the max connections limit is 2
    When a client sends a request
    Then the request should fail with "no server available" error
