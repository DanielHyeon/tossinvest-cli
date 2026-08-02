# Branch Test Map: `TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 126 | `if err != nil {` true/entered and complementary path | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `range` line 130 | `for _, want := range []string{` true/entered and complementary path | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 137 | `if !strings.Contains(text, want) {` true/entered and complementary path | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `range` line 141 | `for _, forbidden := range []string{"accountRef", "capability", "token", "command"} {` true/entered and complementary path | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `if` line 142 | `if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) {` true/entered and complementary path | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
