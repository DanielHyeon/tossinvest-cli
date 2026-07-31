# Function Logic Map: `talliedAs`

- Source: `internal/candidate/tally_alarm_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test totals and passed count | fixture-controlled integers | tally alarm tests | initializes all D3 map entries consistently |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | initialize raised map | fixture map mutation | continue | tally alarm suite |
| B2 | initialize not-measured map | fixture map mutation | return fixture | tally alarm suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes` | mirror production D3 keys | pure test helper | AST |

## State mutations and fallbacks

- Test-local fixture construction only; no production state.

## Safety conclusion

- Safe edit boundary: accessor migration in a fixture helper.
- High-risk impact: no.
