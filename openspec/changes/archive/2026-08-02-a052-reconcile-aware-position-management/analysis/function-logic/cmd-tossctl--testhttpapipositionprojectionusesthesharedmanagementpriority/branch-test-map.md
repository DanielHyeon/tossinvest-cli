# Branch Test Map: `TestHTTPAPIPositionProjectionUsesTheSharedManagementPriority`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 112 | `if managed.AdoptionStatus != "MANAGED" \|\| !managed.StatusKnown \|\| !managed.Excluded \|\| managed.Candidate \|\|` entered and complementary path | TestHTTPAPIPositionProjectionUsesTheSharedManagementPriority | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 121 | `if blocked.AdoptionStatus != "RECONCILE_BLOCKED" \|\| blocked.AdoptionReason != "RECONCILE_BLOCK" \|\|` entered and complementary path | TestHTTPAPIPositionProjectionUsesTheSharedManagementPriority | a052 RED contract or pre-existing regression | verified by focused package suite |
