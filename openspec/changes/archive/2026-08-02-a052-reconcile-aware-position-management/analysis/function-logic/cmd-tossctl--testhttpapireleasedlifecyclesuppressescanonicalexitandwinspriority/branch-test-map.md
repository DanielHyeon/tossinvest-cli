# Branch Test Map: `TestHTTPAPIReleasedLifecycleSuppressesCanonicalExitAndWinsPriority`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 156 | `if out.ExitLine.CurrentProtection != "195" {` entered and complementary path | TestHTTPAPIReleasedLifecycleSuppressesCanonicalExitAndWinsPriority | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 166 | `if out.AdoptionStatus != "UNMANAGED" \|\| out.AdoptionReason != "OPERATOR_RELEASED" \|\| out.Eligible \|\|` entered and complementary path | TestHTTPAPIReleasedLifecycleSuppressesCanonicalExitAndWinsPriority | a052 RED contract or pre-existing regression | verified by focused package suite |
