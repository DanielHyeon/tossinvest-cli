# Branch Test Map: `Console.routes`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | optional monitoring routes remain conditional | console static route suite | existing coverage | pass |
| B2 | optional decision routes remain conditional | console static route suite | existing coverage | pass |
| I1 | optimization lifecycle request body is capped at 4096 bytes | `TestOptimizationRequestBodyIsBounded` | route omitted the limit | pass |
