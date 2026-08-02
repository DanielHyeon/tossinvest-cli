# Branch Test Map: `TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 36 | `if err != nil {` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 39 | `if !got.EffectiveKnown \|\| !got.Effective.Enabled \|\| got.Effective.DefaultStopPct != .03 \|\|` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 45 | `if block.Scope != positionpolicy.ScopeAccount \|\| block.Reason != "RECONCILE_PERMANENT" \|\|` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
