Feature: Caching Mechanism for Frequently Accessed Responses

  As a system administrator
  I want the load balancer to implement a caching mechanism
  So that frequently accessed responses are served faster, reducing backend load

  Background:
    Given the caching mechanism is enabled
    And the cache has a default expiration time of 5 minutes
    And the backend server is responsive

  Scenario: Cache a frequently accessed response
    Given a client sends a request to the backend with "resource-X"
    When the backend responds with "resource-X" for the first time
    Then the response should be cached for "resource-X"
    And the cached response should be served on subsequent requests to "resource-X" within the cache expiration time

  Scenario: Serve cached response if the resource is requested within cache expiration
    Given a client sends a request to the backend with "resource-Y"
    When the backend responds with "resource-Y" for the first time
    And the response is cached for "resource-Y"
    When another client sends a request to "resource-Y" before the cache expiration
    Then the cached response for "resource-Y" should be served
    And no new request should be made to the backend for "resource-Y"

  Scenario: Cache expiration and backend request after expiration
    Given a client sends a request to the backend with "resource-Z"
    When the backend responds with "resource-Z" for the first time
    And the response is cached for "resource-Z"
    And the cache for "resource-Z" expires
    When another client sends a request to "resource-Z" after cache expiration
    Then a new request should be made to the backend for "resource-Z"
    And the new response should be cached for "resource-Z"

  Scenario: Handle cache miss for a new resource
    Given a client sends a request to the backend with "resource-A"
    When the backend responds with "resource-A"
    Then the response should be cached for "resource-A"
    And the cached response for "resource-A" should be served on subsequent requests

  Scenario: Handle cache evictions when the cache is full
    Given the cache is full with cached responses
    When a new resource "resource-B" is requested
    Then the least recently used cached response should be evicted
    And the new response for "resource-B" should be cached