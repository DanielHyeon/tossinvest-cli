# Branch Test Map: `queryPositionPolicies`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 87 | `if err != nil {` true/entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration; TestReadOnlyPositionPoliciesPreservesReleasedLifecycle | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `for` line 92 | `for rows.Next() {` true/entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration; TestReadOnlyPositionPoliciesPreservesReleasedLifecycle | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 94 | `if err != nil {` true/entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration; TestReadOnlyPositionPoliciesPreservesReleasedLifecycle | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 99 | `if err := rows.Err(); err != nil {` true/entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration; TestReadOnlyPositionPoliciesPreservesReleasedLifecycle | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
