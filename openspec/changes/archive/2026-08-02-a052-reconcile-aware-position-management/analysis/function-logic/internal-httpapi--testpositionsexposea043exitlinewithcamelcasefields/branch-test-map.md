# Branch Test Map: `TestPositionsExposeA043ExitLineWithCamelCaseFields`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range` line 100 | `for _, want := range []string{'"exitLine"', '"currentProtection":"68000"', '"statusText":"평가 완료"'} {` entered and complementary path | TestPositionsExposeA043ExitLineWithCamelCaseFields | pre-existing regression at frozen base | verified by current package suite |
| B2 | `if` line 101 | `if !strings.Contains(body, want) {` entered and complementary path | TestPositionsExposeA043ExitLineWithCamelCaseFields | pre-existing regression at frozen base | verified by current package suite |
| B3 | `range` line 105 | `for _, forbidden := range []string{'"ExitLine"', '"CurrentProtection"', '"StatusText"'} {` entered and complementary path | TestPositionsExposeA043ExitLineWithCamelCaseFields | pre-existing regression at frozen base | verified by current package suite |
| B4 | `if` line 106 | `if strings.Contains(body, forbidden) {` entered and complementary path | TestPositionsExposeA043ExitLineWithCamelCaseFields | pre-existing regression at frozen base | verified by current package suite |
