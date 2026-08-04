# Function Logic Map: `ReadOfficial`

- Source: `internal/officialfx/evidence.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context/client/pair/policy | non-nil, canonical cross-currency pair, official-origin client, sealed current haircut policy | official client + officialfx private policy authority | ErrInvalidEvidence/read error; no authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid request scope or policy | none | ErrInvalidEvidence | invalid/policy tests |
| B2 | authoritative read fails | read-only request at most | wrapped error | source/origin tests |
| B3 | authoritative origin specifically unavailable | no HTTP transport invoked | ErrInvalidEvidence | configured zero-hit test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Client.AuthoritativeExchangeRate | prove sealed origin and read under one immutable boundary | no fallback; context/error propagation | official boundary tests |
| sealOfficial | bind response plus sealed policy | pure validation | focused tests |

## State mutations and fallbacks

- No mutation; configured clients fail before token/data transport, and no separate origin token can go stale before the GET.

## Safety conclusion

- Safe edit boundary: never accept raw caller haircut or caller-selected source labels.
- High-risk impact: yes — FX feeds q_final monetary reserve.
