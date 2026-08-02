# Branch Test Map: `NewPositionPolicyCommandService`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 82 | `if ectx == nil \|\| ectx.Journal == nil {` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks; TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 85 | `if clk == nil {` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks; TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
