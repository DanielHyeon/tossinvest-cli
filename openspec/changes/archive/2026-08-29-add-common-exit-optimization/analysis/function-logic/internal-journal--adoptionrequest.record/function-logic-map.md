# Function Logic Map: `AdoptionRequest.record`

- Source: `internal/journal/adoption.go`
- Qualified function: `AdoptionRequest.record`
- AST evidence: `ast.json` (`6adc78fcb71ddfc90ee929f0df2acc005105677d93a61a96c26932bb7a90dcf4`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/adoption.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` at `internal/journal/adoption.go:367` — switch { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B2 | `case` at `internal/journal/adoption.go:368` — case strings.TrimSpace(r.PositionID) == "": | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B3 | `case` at `internal/journal/adoption.go:371` — case a.Symbol == "" \|\| a.Market == "": | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B4 | `case` at `internal/journal/adoption.go:374` — case a.ObservedAt == "": | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B5 | `if` at `internal/journal/adoption.go:379` — if a.ExitPolicyID != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B6 | `if` at `internal/journal/adoption.go:380` — if _, ok := exitpolicy.CommonPolicyByID(a.ExitPolicyID); !ok { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B7 | `if` at `internal/journal/adoption.go:387` — if a.Quantity, err = canonicalQuantity("adopted quantity", orZero(r.Quantity)); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B8 | `if` at `internal/journal/adoption.go:390` — if isZeroDecimal(a.Quantity) { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B9 | `if` at `internal/journal/adoption.go:394` — if a.ObservedPrice, err = positivePrice("observed price", a.ObservedPrice); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B10 | `if` at `internal/journal/adoption.go:397` — if a.SyntheticStop, err = positivePrice("synthetic stop", a.SyntheticStop); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B11 | `if` at `internal/journal/adoption.go:405` — if cerr != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B12 | `if` at `internal/journal/adoption.go:409` — if cmp >= 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B13 | `if` at `internal/journal/adoption.go:416` — if a.CostBasis != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `normaliseSymbol`, `normaliseMarket`, `strings.TrimSpace`, `fmt.Errorf`, `exitpolicy.CommonPolicyByID`, `canonicalQuantity`, `orZero`, `isZeroDecimal`, `positivePrice`, `riskcalc.CompareDecimal`, `adoptionDigest` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 10 assignment(s) and 11 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
