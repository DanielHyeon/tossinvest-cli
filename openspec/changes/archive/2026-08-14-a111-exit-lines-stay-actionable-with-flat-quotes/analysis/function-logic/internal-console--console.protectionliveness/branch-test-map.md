# Branch Test Map: `Console.protectionLiveness`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branch-free happy path: compatibility wrapper performs one marker read and evaluates it at the supplied same-time authority | `TestA111ConsoleConsumesTheSharedFreshnessVerdict` | preservation or intentional RED as named | focused suite GREEN |
