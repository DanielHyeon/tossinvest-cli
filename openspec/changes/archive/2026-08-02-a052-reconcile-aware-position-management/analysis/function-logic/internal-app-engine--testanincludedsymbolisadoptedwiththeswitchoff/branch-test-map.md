# Branch Test Map: `TestAnIncludedSymbolIsAdoptedWithTheSwitchOff`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 47 | `if cycle.Adopted != 1 {` entered and complementary path | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff | pre-existing regression at frozen base | verified by current package suite |
| B2 | `if` line 50 | `if !h.position("005930").Adopted() {` entered and complementary path | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff | pre-existing regression at frozen base | verified by current package suite |
| B3 | `if` line 53 | `if h.position("000660").Adopted() {` entered and complementary path | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff | pre-existing regression at frozen base | verified by current package suite |
| B4 | `if` line 56 | `if cycle.Unmanaged != 1 {` entered and complementary path | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff | pre-existing regression at frozen base | verified by current package suite |
