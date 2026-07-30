# Function Logic Map: `Journal.AdoptPosition`

- Source: `internal/journal/adoption.go`
- Qualified function: `Journal.AdoptPosition`
- AST evidence: `ast.json` (`6adc78fcb71ddfc90ee929f0df2acc005105677d93a61a96c26932bb7a90dcf4`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/adoption.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/journal/adoption.go:180` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B2 | `if` at `internal/journal/adoption.go:185` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B3 | `if` at `internal/journal/adoption.go:199` — if errors.Is(err, sql.ErrNoRows) { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B4 | `if` at `internal/journal/adoption.go:202` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B5 | `if` at `internal/journal/adoption.go:205` — if strings.TrimSpace(decisionID.String) != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B6 | `if` at `internal/journal/adoption.go:208` — if stored := strings.TrimSpace(existing.String); stored != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B7 | `if` at `internal/journal/adoption.go:213` — if stored == adoption.ID { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B8 | `if` at `internal/journal/adoption.go:220` — if _, err := tx.ExecContext(ctx, ' | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B9 | `if` at `internal/journal/adoption.go:229` — if isUniqueViolation(err) { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B10 | `if` at `internal/journal/adoption.go:243` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B11 | `if` at `internal/journal/adoption.go:248` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B12 | `if` at `internal/journal/adoption.go:252` — if affected != 1 { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B13 | `if` at `internal/journal/adoption.go:256` — if err := tx.Commit(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `req.record`, `j.db.BeginTx`, `fmt.Errorf`, `tx.Rollback`, `strings.TrimSpace`, `Scan`, `tx.QueryRowContext`, `errors.Is`, `readAdoptionTx`, `tx.ExecContext`, `nullableString`, `isUniqueViolation` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 9 assignment(s) and 14 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
