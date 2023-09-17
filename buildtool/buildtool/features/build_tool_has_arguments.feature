Feature: A build tool has arguments
    BuildTools pass arguments on to some executable.

    Scenario: Exact arguments are passed
    Given a literal string "arg1"
    When I call BuildToolArguments
    Then a list of 1 string should be returned
    And "arg1" should be the only member