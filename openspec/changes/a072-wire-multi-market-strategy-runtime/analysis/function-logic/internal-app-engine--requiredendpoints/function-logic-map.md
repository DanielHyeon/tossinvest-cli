# Function Logic Map: `RequiredEndpoints`

- Source: `internal/app/engine/interlock.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| global engine dependency catalog | exact method/path for shared safety-loop reads and mutations only | official runtime calls | capability attestation missing any item refuses startup; strategy-local reads are excluded |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | direct return | none; returns a fresh slice without the strategy-only exchange-rate read | complete shared startup endpoint set | endpoint coverage and strategy FX isolation tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | static catalog only | no error, timeout, retry, I/O or fallback | CodeGraph + AST |

## State mutations and fallbacks

- No durable or network mutation. The US strategy FX loader still verifies the official read and refuses
  US entry locally; global attestation no longer lets that strategy-only dependency stop safety loops.

## Safety conclusion

- Safe edit boundary: separate one strategy-local read from shared runtime startup evidence.
- High-risk impact: **yes** — startup capability attestation is fail-closed.
