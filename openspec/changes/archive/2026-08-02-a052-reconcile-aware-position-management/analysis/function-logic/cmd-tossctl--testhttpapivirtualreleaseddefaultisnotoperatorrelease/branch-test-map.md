# Branch Test Map: `TestHTTPAPIVirtualReleasedDefaultIsNotOperatorRelease`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 176 | `if managed \|\| released {` entered and complementary path | TestHTTPAPIVirtualReleasedDefaultIsNotOperatorRelease | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 183 | `if out.AdoptionStatus != "UNMANAGED" \|\| out.AdoptionReason != "NOT_SELECTED" {` entered and complementary path | TestHTTPAPIVirtualReleasedDefaultIsNotOperatorRelease | a052 RED contract or pre-existing regression | verified by focused package suite |
