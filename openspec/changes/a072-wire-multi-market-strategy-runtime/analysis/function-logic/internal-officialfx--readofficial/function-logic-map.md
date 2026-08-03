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
| B1 | invalid request/origin/policy | none | fail closed | invalid/overridden origin tests |
| B2 | official GET fails | read-only request only | wrapped error | source error tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Client.AuthorityOrigin | prove default official endpoint and transport | no network | official client provenance |
| Client.ExchangeRate | read official rate | context/error propagation | current source |
| sealOfficial | bind response plus sealed policy | pure validation | focused tests |

## State mutations and fallbacks

- No mutation; override clients may perform ordinary reads but must fail before the GET in this authority path.

## Safety conclusion

- Safe edit boundary: never accept raw caller haircut or caller-selected source labels.
- High-risk impact: yes — FX feeds q_final monetary reserve.
