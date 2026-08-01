# Function Logic Map: `TestRiskPreviewWaitsThreeSecondsAndAppliesSameCapabilityExactlyOnce`

- Source: `internal/console/optimization_review_block_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| real optimization store + fake clock | base v1 and one server preset candidate | temp SQLite + deterministic clock | any premature/replayed mutation fails |
| rendered preview | CSP-bound static script, read-only diff, one checkbox/button | console preview template | arbitrary input or capability mutation fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B23 | response/CSP/DOM/input/countdown checks, early 425, t+3 success, replay idempotency | test assertions and temp-store writes only | exact invariant failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| store Preview/Apply/Read | verify one-shot capability lifecycle | deterministic local DB; no broker/LIVE/order calls | AST |
| HTML parser + SHA-256 | bind script bytes to CSP and inspect controls | pure | AST |

## State mutations and fallbacks

- Mutates a temporary optimization settings DB only. The early and replay branches prove audit/version do not advance.

## Safety conclusion

- Safe edit boundary: server preset capability confirmation only.
- High-risk impact: no trading side effect; high-risk settings approval test.
