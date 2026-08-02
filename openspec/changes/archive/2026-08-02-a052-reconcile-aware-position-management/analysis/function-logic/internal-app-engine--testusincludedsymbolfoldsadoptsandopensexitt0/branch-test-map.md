# Branch Test Map: `TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 70 | `if cycle.Err != nil \|\| cycle.Folded != 1 \|\| cycle.Adopted != 1 \|\| cycle.Unmanaged != 0 {` true/entered and complementary path | TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 75 | `if !p.Adopted() \|\| !p.ExitEligible() \|\| provenance != positionpolicy.ProvenanceExternalAdoption \|\|` true/entered and complementary path | TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 80 | `if err != nil \|\| adoption.ObservedPrice != "200" \|\| adoption.SyntheticStop != "190" {` true/entered and complementary path | TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 84 | `if err != nil \|\| exitState.EntryPrice != "200" \|\| exitState.HighWater != "200" \|\|` true/entered and complementary path | TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
