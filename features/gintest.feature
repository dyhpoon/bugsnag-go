Feature: Running a request via gin

Background:
  Given I set environment variable "API_KEY" to "a35a2a72bd230ac0aa0f52715bbdc6aa"
  And I configure the bugsnag endpoint

Scenario Outline: An error report was sent from the gin framework
  Given I set environment variable "GIN_VERSION" to "<gin version>"
  And I start the service "gin-default"
  And I wait for the app to open port "4511"
  And I wait for 1 seconds
  Then I open the URL "http://localhost:4511/basic"
  And I wait for 1 seconds
  Then I should receive 2 requests
  And the request is valid for the error reporting API
  And the request contained the api key "a35a2a72bd230ac0aa0f52715bbdc6aa"

  Examples:
  | gin version |
  | v1.3.0      |
  | v1.2.0      |

