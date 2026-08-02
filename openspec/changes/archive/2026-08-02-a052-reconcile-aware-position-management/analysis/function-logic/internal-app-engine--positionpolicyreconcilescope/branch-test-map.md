# Branch Test Map: `positionPolicyReconcileScope`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `switch` line 121 | `switch scope {` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `case` line 122 | `case reconcile.ScopeAccount:` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `case` line 124 | `case reconcile.ScopeMarket:` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `case` line 126 | `case reconcile.ScopeSymbol:` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `case` line 128 | `default:` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
