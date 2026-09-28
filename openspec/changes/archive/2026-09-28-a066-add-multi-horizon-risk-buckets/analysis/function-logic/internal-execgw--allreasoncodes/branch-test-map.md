# Branch Test Map: `AllReasonCodes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path — no branch (AST 0); the output list is pinned whole | `TestReasonCodeEnumIsStable` | golden diff when a code is added without regenerating | GREEN at `c1d1e295` (pre-edit) |
