# Branch Test Map: `TestHTTPAPIRuntimeFailureRemainsUnknownData`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 131 | `if runtime.EffectiveKnown {` entered and complementary path | TestHTTPAPIRuntimeFailureRemainsUnknownData | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 136 | `if position.AdoptionStatus != "UNKNOWN" \|\| position.StatusKnown \|\| position.DesignationKnown \|\|` entered and complementary path | TestHTTPAPIRuntimeFailureRemainsUnknownData | a052 RED contract or pre-existing regression | verified by focused package suite |
