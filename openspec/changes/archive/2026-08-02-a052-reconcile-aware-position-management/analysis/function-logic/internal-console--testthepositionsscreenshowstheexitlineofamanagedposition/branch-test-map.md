# Branch Test Map: `TestThePositionsScreenShowsTheExitLineOfAManagedPosition`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range` line 284 | `for _, want := range []string{` true/entered and complementary path | TestThePositionsScreenShowsTheExitLineOfAManagedPosition | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 290 | `if !strings.Contains(page, want) {` true/entered and complementary path | TestThePositionsScreenShowsTheExitLineOfAManagedPosition | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `range` line 295 | `for _, staleRaw := range []string{"HALF_RISK", "intent-77"} {` true/entered and complementary path | TestThePositionsScreenShowsTheExitLineOfAManagedPosition | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 296 | `if strings.Contains(row, staleRaw) {` true/entered and complementary path | TestThePositionsScreenShowsTheExitLineOfAManagedPosition | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `range` line 300 | `for _, evidence := range []string{"원장 기록 · 실효 미확인", "원장 기준선 <strong>69500</strong>",` true/entered and complementary path | TestThePositionsScreenShowsTheExitLineOfAManagedPosition | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B6 | `if` line 302 | `if !strings.Contains(row, evidence) {` true/entered and complementary path | TestThePositionsScreenShowsTheExitLineOfAManagedPosition | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
