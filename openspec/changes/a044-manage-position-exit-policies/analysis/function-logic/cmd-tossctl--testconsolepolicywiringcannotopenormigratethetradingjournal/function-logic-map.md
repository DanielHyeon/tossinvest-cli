# Function Logic Map: `TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| console source | production `console.go` text | repository file | test fails if unreadable |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | forbidden constructor/write symbol appears | none | test failure | this test |\n| B3-B4 | required descriptor/Dial symbol absent | none | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.ReadFile` | inspect exact production source | local deterministic read | AST |

## State mutations and fallbacks

- Static assertion only; no runtime state mutation.

## Safety conclusion

- Safe edit boundary: keep the console process structurally unable to construct a writable journal path.
- High-risk impact: no
