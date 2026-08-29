# Function Logic Map: `TestClientBaseURL`

- Source: `internal/official/client_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| explicitly configured base URL | normalized URL without trailing slash | WithBaseURL | test fails on mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | configured base differs from expected normalization | none | `t.Fatalf` | TestClientBaseURL |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Client.BaseURL | observe normalized test configuration | pure getter | AST |

## State mutations and fallbacks

- Test only; preserves ordinary configurable read-client behavior.

## Safety conclusion

- Safe edit boundary: explicit configuration remains usable but loses authority.
- High-risk impact: no production mutation.
