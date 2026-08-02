# Branch Test Map: `TestManagedPositionPrimaryRowShowsTheProtectionDecision`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 80 | `if start < 0 {` true/entered and complementary path | TestManagedPositionPrimaryRowShowsTheProtectionDecision | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 84 | `if end < 0 {` true/entered and complementary path | TestManagedPositionPrimaryRowShowsTheProtectionDecision | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `range` line 88 | `for _, want := range []string{` true/entered and complementary path | TestManagedPositionPrimaryRowShowsTheProtectionDecision | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 98 | `if !strings.Contains(row, want) {` true/entered and complementary path | TestManagedPositionPrimaryRowShowsTheProtectionDecision | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `range` line 102 | `for _, raw := range []string{"HALF_RISK", "intent-77"} {` true/entered and complementary path | TestManagedPositionPrimaryRowShowsTheProtectionDecision | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B6 | `if` line 103 | `if strings.Contains(row, raw) {` true/entered and complementary path | TestManagedPositionPrimaryRowShowsTheProtectionDecision | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B7 | `if` line 107 | `if !strings.Contains(row, "원장 기준선 <strong>69500</strong>") \|\|` true/entered and complementary path | TestManagedPositionPrimaryRowShowsTheProtectionDecision | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B8 | `if` line 112 | `if !strings.Contains(page, '<caption>보유 종목과 보호 상태</caption>') \|\|` true/entered and complementary path | TestManagedPositionPrimaryRowShowsTheProtectionDecision | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
