# Branch Test Map: `TestTheReadOnlyHandleHasNoWriteMethods`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `for` line 140 | `for i := 0; i < typ.NumMethod(); i++ {` true/entered and complementary path | TestTheReadOnlyHandleHasNoWriteMethods | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 142 | `if !allowed[name] {` true/entered and complementary path | TestTheReadOnlyHandleHasNoWriteMethods | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
