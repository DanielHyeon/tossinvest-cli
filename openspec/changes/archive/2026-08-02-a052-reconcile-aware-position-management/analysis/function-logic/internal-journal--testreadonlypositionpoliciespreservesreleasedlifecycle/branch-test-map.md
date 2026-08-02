# Branch Test Map: `TestReadOnlyPositionPoliciesPreservesReleasedLifecycle`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 174 | `if _, err := j.ApplyPositionPolicy(context.Background(), policyRequest(positionpolicy.ActionRelease, 0)); err != nil {` true/entered and complementary path | TestReadOnlyPositionPoliciesPreservesReleasedLifecycle | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 178 | `if err != nil \|\| len(states) != 1 \|\| states[0].PositionID != "p-policy" \|\|` true/entered and complementary path | TestReadOnlyPositionPoliciesPreservesReleasedLifecycle | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
