# Function Logic Map: `newHarness`

- Source: `internal/console/console_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| option tweaks | zero or more test-only callbacks | individual tests | setup fails test immediately |
| temporary evidence paths | private test directory | `t.TempDir` | cleaned by testing runtime |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | apply option tweaks | test-only option mutation | continues | console suite |
| B2 | console construction fails | none | fatal test | console suite |
| B3 | cookie jar construction fails | none | fatal test | console suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `New` | build real console under test | fatal on setup error | CodeGraph + AST |
| `httptest.NewServer` | preserve legacy plaintext loopback coverage | cleanup closes server | AST |

## State mutations and fallbacks

- The helper intentionally remains plaintext for non-secret legacy route tests.

## Safety conclusion

- Safe edit boundary: unchanged; a separate TLS helper owns credential ingress tests.
- High-risk impact: no, test infrastructure only.
