# Branch Test Map: `TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range` line 91 | `for _, currency := range []string{"KRW", ""} {` true/entered and complementary path | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 98 | `if cycle.Adopted != 0 \|\| cycle.Deferred != 1 {` true/entered and complementary path | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 102 | `if p.Adopted() {` true/entered and complementary path | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
