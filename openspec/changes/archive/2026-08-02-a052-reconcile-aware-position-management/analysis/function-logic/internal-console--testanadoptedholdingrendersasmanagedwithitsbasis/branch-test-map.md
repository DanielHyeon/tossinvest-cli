# Branch Test Map: `TestAnAdoptedHoldingRendersAsManagedWithItsBasis`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 874 | `if !strings.Contains(page, "편입 기록") {` true/entered and complementary path | TestAnAdoptedHoldingRendersAsManagedWithItsBasis | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 877 | `if !strings.Contains(page, "진입 결정") {` true/entered and complementary path | TestAnAdoptedHoldingRendersAsManagedWithItsBasis | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 882 | `if strings.Contains(row, "관리 외(미편입)") {` true/entered and complementary path | TestAnAdoptedHoldingRendersAsManagedWithItsBasis | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 885 | `if !strings.Contains(row, "엔진 관리") {` true/entered and complementary path | TestAnAdoptedHoldingRendersAsManagedWithItsBasis | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `if` line 890 | `if !strings.Contains(row, "원장 기록 · 실효 미확인") \|\|` true/entered and complementary path | TestAnAdoptedHoldingRendersAsManagedWithItsBasis | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
