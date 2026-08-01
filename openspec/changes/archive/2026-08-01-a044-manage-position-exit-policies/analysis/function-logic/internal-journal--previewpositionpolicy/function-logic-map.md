# Function Logic Map: `previewPositionPolicy`

- Source: `internal/journal/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| current/request | exact generation/version, server action/reason | transaction/list projection | typed refusal |
| provenance eligibility | release/readopt only external-adopted | position IDs | ineligible |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B20 | CAS/state/action transitions and effective reset identity | in-memory after only | preview/error | journal policy tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| typed state eligibility | authorize lifecycle transition | no fallback from generic exit eligibility | AST |

## State mutations and fallbacks

- Preview is pure. READOPT after effective policy must equal the reset observation policy.

## Safety conclusion

- Safe edit boundary: enforce provenance at journal boundary and compute exact deterministic after state.
- High-risk impact: yes
