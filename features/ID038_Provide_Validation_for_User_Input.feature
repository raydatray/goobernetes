Feature: Load Balancer User Input Validation

  As a system administrator
  I want to validate user-provided configuration inputs,
  So that I can ensure the load balancer is configured correctly and securely.

  Background:
    Given the load balancer service is running
    And I am authenticated as an administrator
    And the backend server pool is accessible

  Scenario Outline: Validate Backend Server Registration
    Given I am configuring a new backend server
    When I enter the following server details:
      | Field             | Value                |
      | Server Name       | <server_name>        |
      | IP Address        | <ip_address>         |
      | Port              | <port>               |
      | Max Connections   | <max_connections>    |
      | Weight            | <weight>             |
      | Expected          | <expected>           |

    Then the system should validate:
      | Field           | Rules                                                    |
      | Server Name     | - Must be 1-64 characters                                |
      |                 | - Must contain only alphanumeric, hyphens, underscores   |
      | IP Address      | - Must be valid IPv4 format.                             |
      |                 | - Must not be unspecified (0.0.0.0)                      |
      | Port            | - Must be between 1-65535                                |
      | Max Connections | - Must be a positive integer                             |
      | Weight          | - Must be between 1-100                                  |

    Examples:
      | ip_address    | port  | max_connections | weight | server_name     | expected                                              |
      | 192.168.1.100 | 8080  | 50              | 50     | app-server-1    | Success                                               |
      | 192.168.1.101 | 8080  | 50              | 0      | app-server-2    | Invalid weight (must be between 1-100 inclusive): 0   |
      | 192.168.1.102 | 8080  | 50              | 101    | app-server-3    | Invalid weight (must be between 1-100 inclusive): 101 |
      | 0.0.0.0       | 80    | 75              | 50     | localhost       | Invalid IP address: 0.0.0.0                           |
      | 256.0.0.1     | 8080  | 75              | 50     | invalid-ip      | Invalid IP address: 256.0.0.1                         |
      | 192.0.5.1     | 65536 | 75              | 50     | invalid-port    | Invalid port number: 65536                            |
      | 192.0.5.1     | 0     | 75              | 50     | invalid-port    | Invalid port number: 0                                |
      | 255.1.2.3     | 8080  | -5              | 50     | iloveG1         | Invalid max connections: -5                           |
      | 192.5.6.7     | 8080  | 0               | 50     | ilove@G2        | Invalid character in server name                      |
