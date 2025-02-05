Feature: Weighted Round Robin Load Balancing

  As a system administrator,
  I want the load balancer to distribute requests using the Weighted Round Robin strategy,
  So that backend servers receive traffic based on their assigned weights.

  Background:
    Given the load balancer is running
    And the following backend servers are configured:
      | server_id | weight | address      | port | max_connections |
      | server1   | 3      | 192.168.1.10 | 8080 | 1000            |
      | server2   | 1      | 192.168.1.11 | 8080 | 1000            |
      | server3   | 2      | 192.168.1.12 | 8080 | 1000            |

  Scenario: Normal Flow - Requests are distributed based on weights
    When a client makes 6 consecutive requests
    Then the requests should be routed in this order:
      | request | server  |
      | 1       | server1 |
      | 2       | server1 |
      | 3       | server1 |
      | 4       | server2 |
      | 5       | server3 |
      | 6       | server3 |

  Scenario: Alternative Flow - A higher-weight server goes down
    Given "server1" becomes unavailable
    When a client makes 6 consecutive requests
    Then the requests should be routed in this order:
      | request | server  |
      | 1       | server2 |
      | 2       | server3 |
      | 3       | server3 |
      | 4       | server2 |
      | 5       | server3 |
      | 6       | server3 |

  Scenario: Error Flow - No backend servers available
    Given all backend servers are unavailable
    When a client makes a request
    Then the load balancer should return a "503 Service Unavailable" response
