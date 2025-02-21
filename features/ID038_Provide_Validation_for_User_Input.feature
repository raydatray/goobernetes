Feature: Load Balancer User Input Validation

  As a system administrator
  I want to validate user-provided configuration inputs,
  So that I can ensure the load balancer is configured correctly and securely.

  Background:
    Given the load balancer service is running
    And I am authenticated as an administrator
    And the backend server pool is accessible

  Scenario: Validate Backend Server Registration
    Given I am configuring a new backend server
    When I enter the following server details:
      | Field             | Value                |
      | IP Address        | <ip_address>         |
      | Port              | <port>               |
      | Weight            | <weight>             |
      | Server Name       | <server_name>        |
    Then the system should validate:
      | Field       | Rules                                                    |
      | IP Address  | - Must be valid IPv4/IPv6 format                         |
      |             | - Must not be in restricted ranges (127.0.0.1, 0.0.0.0)  |
      |             | - Must be reachable                                      |
      | Port        | - Must be between 1-65535                                |
      |             | - Must be a number                                       |
      | Weight      | - Must be between 1-100                                  |
      |             | - Must be a positive integer                             |
      | Server Name | - Must be 1-64 characters                                |
      |             | - Must contain only alphanumeric, hyphens, underscores   |
      |             | - Must be unique in the backend pool                     |

    Examples:
      | ip_address    | port  | weight | server_name     |
      | 192.168.1.100 | 8080  | 50     | app-server-1    |
      | 127.0.0.1     | 80    | 75     | localhost       |
      | 256.1.2.3     | 70000 | -5     | bad@server      |
