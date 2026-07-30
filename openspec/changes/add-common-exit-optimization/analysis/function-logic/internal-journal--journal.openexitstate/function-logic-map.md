# Function Logic Map: `Journal.OpenExitState`

- Source: `internal/journal/exit_state.go`
- Qualified function: `Journal.OpenExitState`
- AST evidence: `ast.json` (`941f908df3758d1efc70be01dd276e47267fae61a5dcf8f8cfa1d2ba45ee924f`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/exit_state.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/journal/exit_state.go:108` — if id == "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B2 | `if` at `internal/journal/exit_state.go:112` — if kind == "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B3 | `if` at `internal/journal/exit_state.go:115` — if kind != ExitPolicyRatchet && kind != ExitPolicyLadder { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B4 | `switch` at `internal/journal/exit_state.go:120` — switch kind { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B5 | `case` at `internal/journal/exit_state.go:121` — case ExitPolicyRatchet: | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B6 | `if` at `internal/journal/exit_state.go:122` — if policyID != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B7 | `case` at `internal/journal/exit_state.go:125` — case ExitPolicyLadder: | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B8 | `if` at `internal/journal/exit_state.go:126` — if policyID == "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B9 | `else` at `internal/journal/exit_state.go:128` — } else if _, ok := exitpolicy.CommonPolicyByID(policyID); !ok && policyID != "default_v1" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B10 | `if` at `internal/journal/exit_state.go:128` — } else if _, ok := exitpolicy.CommonPolicyByID(policyID); !ok && policyID != "default_v1" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B11 | `if` at `internal/journal/exit_state.go:136` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B12 | `if` at `internal/journal/exit_state.go:142` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B13 | `if` at `internal/journal/exit_state.go:151` — if errors.Is(err, sql.ErrNoRows) { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B14 | `if` at `internal/journal/exit_state.go:154` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B15 | `if` at `internal/journal/exit_state.go:157` — if !position.ExitEligible(decisionID.String, adoptionID.String) { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B16 | `if` at `internal/journal/exit_state.go:161` — if _, err := tx.ExecContext(ctx, ' | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B17 | `if` at `internal/journal/exit_state.go:168` — if isUniqueViolation(err) { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B18 | `if` at `internal/journal/exit_state.go:173` — if err := appendExitEventTx(ctx, tx, exitEventRow{ | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B19 | `if` at `internal/journal/exit_state.go:179` — if err := tx.Commit(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace`, `fmt.Errorf`, `exitpolicy.CommonPolicyByID`, `exitpolicy.OpenRatchetState`, `j.nowString`, `j.db.BeginTx`, `tx.Rollback`, `Scan`, `tx.QueryRowContext`, `errors.Is`, `position.ExitEligible`, `tx.ExecContext` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 13 assignment(s) and 14 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
