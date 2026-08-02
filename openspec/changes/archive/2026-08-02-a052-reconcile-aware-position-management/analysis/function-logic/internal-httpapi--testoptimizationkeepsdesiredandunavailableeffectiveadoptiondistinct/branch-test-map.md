# Branch Test Map: `TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 180 | `if err != nil {` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 190 | `if !management.Desired.Enabled \|\| management.Desired.DefaultStopPct != 0.03 \|\|` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 195 | `if management.EffectiveKnown \|\| management.Effective != nil {` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 198 | `if !management.AutoEnabledDesired \|\| management.StopDesired != "3%" \|\|` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `if` line 206 | `if management.AutoEnabledEffective \|\| management.StopEffective != "5%" \|\| management.Effective == nil {` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
