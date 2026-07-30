# Function Logic Map: `ReconcileDriver.adoptOne`

- Source: `internal/app/engine/adoption.go`
- AST evidence: `ast.json` (`2a1ce278711c340c4de8f53d21576c72e53f1d8b457f8f38adf6d3207c3e7a43`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | fresh observed price and validated adoption stop configuration | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 283 | only the branch's documented state transition | existing return/error contract | `TestAdoptOnePolicySnapshot` |
| B2 | existing if branch at line 302 | only the branch's documented state transition | existing return/error contract | `TestAdoptOnePolicySnapshot` |
| B3 | existing if branch at line 307 | only the branch's documented state transition | existing return/error contract | `TestAdoptOnePolicySnapshot` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SyntheticStop, AdoptPosition | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- snapshot startup common policy into request; adoption itself never submits an order.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: snapshot startup common policy into request; adoption itself never submits an order.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
