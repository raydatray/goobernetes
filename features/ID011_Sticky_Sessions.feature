Feature: Implement Sticky Sessions

  As a system administrator
  I want the load balancer to implement sticky sessions
  So that a client is consistently directed to the same backend server during their session.

  Background:
    Given the load balancer is running with sticky session support enabled
    And the backend servers are configured to handle requests from clients
    And the load balancer uses a session cookie to identify clients

  Scenario: Successful request with sticky session
    Given a client sends a request to the load balancer for "resource-A"
    When the load balancer routes the request to a backend server
    Then the client should receive a session cookie identifying the backend server
    And subsequent requests from the client should be routed to the same backend server using the session cookie

  Scenario: Sticky session fails if no backend is available
    Given the load balancer has a session cookie for a client identifying "server-1"
    And "server-1" is down
    When the client sends a request
    Then the load balancer should route the request to an available backend server
    And the client should receive a new session cookie for the new backend server

  Scenario: Route client to the same backend server during the session
    Given a client sends multiple requests to the load balancer for "resource-A"
    And the first request is routed to "server-1"
    When the client sends a second request
    Then the load balancer should route the request to "server-1" based on the session cookie

  Scenario: Timeout of sticky session
    Given the load balancer is running with sticky sessions and session timeout set to 30 minutes
    When the client has not made a request for 35 minutes
    Then the load balancer should treat the client as a new session on the next request
    And the client should be routed to an available backend server based on the load balancing algorithm
