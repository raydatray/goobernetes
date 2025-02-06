Feature: Configure Maximum Connections Per Backend Server
  As a system administrator
  I want to set maximum connection limits for each backend server
  So that I can control server load and prevent overloading

  Background:
    Given the load balancer is running
    And I am authenticated as an administrator
    And a backend server exists with address "192.168.1.10" and port "8080"

  Scenario: Set valid maximum connections for a backend server
    When I set the maximum connections to 1000 for the backend server
    Then the maximum connections should be updated successfully
    And I should see a confirmation message
    And the backend server should accept connections up to the new limit

  Scenario: Attempt to set invalid maximum connections
    When I try to set the maximum connections to -50 for the backend server
    Then I should receive an error message "Maximum connections must be a positive number"
    And the maximum connections setting should remain unchanged

  Scenario: Update existing maximum connections
    Given the backend server has a maximum connection limit of 1000
    When I update the maximum connections to 1500
    Then the maximum connections should be updated to 1500
    And existing connections should not be affected

  Scenario: Set maximum connections near system limits
    When I try to set the maximum connections to 1000000
    Then I should receive a warning about high connection limits
    But the maximum connections should still be updated