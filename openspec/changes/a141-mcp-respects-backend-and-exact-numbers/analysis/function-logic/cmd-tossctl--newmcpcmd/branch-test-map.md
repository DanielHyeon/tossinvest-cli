# Branch Test Map: `newMCPCmd`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | config load error returns before construction | Existing command/config coverage; retain | no | pending |
| B2 | OpenAPI path resolution error returns before construction | Existing path coverage; retain | no | pending |
| B3 | credential load error returns before construction | Existing credential coverage; retain | no | pending |
| B4 | Stored credentials + disabled OpenAPI | New offline command wiring test: official constructor/dependency unavailable | pending | pending |
| B9 | Effective backend is WTS, auto, or openapi | New offline command wiring matrix: WTS/disabled deny official and auto/openapi retain it | pending | pending |
| B5 | malformed WTS session load returns error | Existing session coverage; retain | no | pending |
| B6 | WTS session present | New offline command wiring test: WTS operation remains callable | pending | pending |
| B8 | no session builds a sessionless router only | New offline command wiring test: no official bypass | pending | pending |
| B7 | neither effective credential available | Existing/new offline login-guidance test | pending | pending |
| B10 | update cache path resolves | Existing non-blocking update behavior; retain | no | pending |
| B11 | update available appends instructions | Existing non-blocking update behavior; retain | no | pending |
