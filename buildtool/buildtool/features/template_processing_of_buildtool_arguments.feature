Feature: Template processing of BuildTool Arguments
    Users should be able to interpolate strings and apply simple
    functions to the BuildToolsArguments for a BuildTool in a Target.


    Scenario: Parse arguments merges things
        Given an Build.Inputs of "sun,moon"
        And a BuildToolArguments of "{{ .Build.Inputs | Merge }}"
        When ParseArguments is called
        Then it returns BuildToolArguments
        And the list has 2 members
        And the members in order are "sun,moon"