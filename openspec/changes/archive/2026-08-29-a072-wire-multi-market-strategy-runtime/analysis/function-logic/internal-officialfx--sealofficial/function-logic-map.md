# Function Logic Map: `sealOfficial`

- Source: `internal/officialfx/evidence.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| official response/pair/policy | exact positive decimals, exact pair, valid window, sealed canonical haircut policy current for response | official endpoint + private policy authority | ErrInvalidEvidence; no partial evidence |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | any response or policy invariant fails | none | ErrInvalidEvidence | invalid table + policy tamper tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| decimal/time parsing | lossless bounded validation | pure; no fallback | current source |
| digest/seal | bind complete canonical preimage | SHA-256 integrity, not caller authority | current source |

## State mutations and fallbacks

- Creates immutable evidence only after response and policy are independently valid.

## Safety conclusion

- Safe edit boundary: preserve raw official decimals; canonicalize haircut through policy.
- High-risk impact: yes — monetary conversion.
