# Function Logic Map: `TestTheStatusColumnHeaderSaysAdoption`

- Source: `internal/console/portfolio_label_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| rendered positions page | authenticated fixture with settings and journal | template + operator-console spec | test error on missing/legacy header |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | accessible `관리 편입` column header absent | records test error | continue | this test |
| B2 | legacy bare `관리` header present | records test error | return normally after assertion | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| settingsHarness/seedJournal | build isolated fixture inputs | temp files and fake seams only | test setup |
| strings.Contains | assert exact accessible markup | pure scan | AST B1-B2 |

## State mutations and fallbacks

- Mutates only temporary test journal/settings fixtures. No network or live account binding.

## Safety conclusion

- Safe edit boundary: expected HTML changed only to include `scope="col"`.
- High-risk impact: no; assertion-only test change.
