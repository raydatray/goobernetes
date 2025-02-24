Feature: Least Connections Load Balancing Algorithm
  As a load balancer
  I want to distribute traffic based on active connection counts
  So that I can prevent server overload and ensure optimal resource usage

  Background:
    Given the load balancer is running
    And the following backend servers are configured:
      | server_id | weight | address      | port | max_connections |
      | server1   | 1      | 192.168.1.10 | 8080 | 1000            |
      | server2   | 1      | 192.168.1.11 | 8080 | 1000            |
      | server3   | 1      | 192.168.1.12 | 8080 | 1000            |

  Scenario: Select server with least active connections
    Given the backend servers have the following active connections:
      | server_id | active_connections |
      | server1   | 100                |
      | server2   | 50                 |
      | server3   | 75                 |
    When a new client request arrives
    Then the request should be routed to "server2"
    And the active connection count for "server2" should increase by 1

  Scenario: Handle connection completion
    Given server "server2" has 51 active connections
    When a client completes their request to "server2"
    Then the active connection count for "server2" should decrease by 1

  Scenario: Handle server at max connections
    Given server "server1" is at maximum connections
    When a new client request arrives
    Then the request should be routed to the next server with least connections
    And a warning should be logged about server "server1" reaching capacity

  Scenario: All servers busy
    Given all servers are at 90% of their maximum connections
    When a new client request arrives
    Then the request should still be routed to the server with least connections
    And a critical alert should be generated about high connection load