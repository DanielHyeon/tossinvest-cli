# Function Logic Map: `candidateVetoThresholds`

- Source: `cmd/tossctl/vetothresholds.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| No runtime input | none | OpenSpec a046 review: dormant `unapproved/passed=0` | Return all threshold strings absent so every veto is fail-closed/unmeasured |
| Legacy `near_high=2.0` | provenance only; not active | a046 design decision 8 and review finding 3 | Never copy the legacy value into the runtime `VetoThresholds` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unconditional leaf return | none; returns a new value | dormant zero-value `candidate.VetoThresholds` | `TestTheTwoSurfacesApplyTheSameThresholds`, a046 dormant threshold source test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | constructor is a pure leaf | no error, timeout, retry, I/O, polling, or mutation | CodeGraph callers: `candidateCycleOptions`, `consoleSignalsMarket`; AST `calls: null` |

## State mutations and fallbacks

- No mutation. Each call returns a fresh value.
- No fallback. An absent approved registry remains absent rather than falling back to the legacy 2.0 value or another market.

## Safety conclusion

- Safe edit boundary: remove the legacy numeric field from this constructor only; immutable approved sets are loaded by the new pure candidate contract and are not wired by this dormant change.
- High-risk impact: no. This is a read-only candidate assessment input and the dependency guards prevent order, RiskIntent, Guardian, ledger, or broker access. The safe direction is stricter: `near_high` becomes unmeasured, never passed.
