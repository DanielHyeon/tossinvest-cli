# Function Logic Map: `WithAccountSeq`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account sequence option/client seal | any integer during construction; immutable afterward | `official.New` | replay after construction is ignored |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | configuration already sealed | none | no-op | retained-option replay test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `atomic.Int64.Store` | publish constructor-selected account sequence | no I/O | AST |

## State mutations and fallbacks

- Writes account selection only under `configMu` before the client is sealed.

## Safety conclusion

- Safe edit boundary: preserve constructor options while forbidding post-construction replay.
- High-risk impact: yes, the same retained pointer could otherwise mutate account scope.
