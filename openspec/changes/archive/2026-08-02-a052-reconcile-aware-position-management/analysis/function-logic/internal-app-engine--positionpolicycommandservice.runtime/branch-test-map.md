# Branch Test Map: `PositionPolicyCommandService.Runtime`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 109 | `if s.blocks == nil {` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `range` line 112 | `for _, block := range s.blocks.Blocks() {` true/entered and complementary path | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
