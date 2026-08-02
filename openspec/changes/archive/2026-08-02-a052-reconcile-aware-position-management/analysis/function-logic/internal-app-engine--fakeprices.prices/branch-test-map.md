# Branch Test Map: `fakePrices.Prices`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 53 | `if f.err != nil {` true/entered and complementary path | fakePrices.Prices | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `range` line 57 | `for _, s := range symbols {` true/entered and complementary path | fakePrices.Prices | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 58 | `if v, ok := f.last[s]; ok {` true/entered and complementary path | fakePrices.Prices | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 60 | `if configured, ok := f.currencies[s]; ok {` true/entered and complementary path | fakePrices.Prices | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
