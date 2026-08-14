# Function Logic Map: `exitObservationRefreshGuardTx`

- Source: `internal/journal/apply_hook.go`
- Post-edit AST evidence: `ast.json` (1 branch; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| transaction | `RefreshExitObservation`'s active `BEGIN IMMEDIATE` transaction | caller | query error propagates for rollback |
| position ID | current durable exit-state identity | refresh request | missing row is a query error; no fallback or write |
| output authority | snapshot status plus derived `proposalPending` boolean only | one transactional SELECT | raw guarded state and transaction are not exported; helper can never mutate |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | status/pending SELECT or scan fails at `internal/journal/apply_hook.go:688` | none; read-only helper | empty outputs plus exact query error | `TestA111RefreshFailureRollsBackTheWholeTuple`, `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` |
| Return | SELECT succeeds | trims pending action and returns a boolean; no durable mutation | status, pending flag, nil | `TestA111RefreshRejectsARealCurrentPendingProposalAndPreservesItsEvidence` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `tx.QueryRowContext(...).Scan` | read status and pending inside the refresh transaction | one read; any error propagates; no retry | AST + pending/lifecycle REDs |
| `strings.TrimSpace` | derive presence without returning raw guarded value | whitespace-only pending is false | AST |

## State mutations and fallbacks

- Performs no `UPDATE`, append, proposal arm, event, commit, or order operation.
- Placement in `apply_hook.go` preserves the package rule that guarded fill-time columns are named only at the atomic apply boundary.
- `RefreshExitObservation` consumes only the derived boolean, then separately enforces current managed lifecycle and temporal CAS inside the same transaction.

## Safety conclusion

- Safe edit boundary: unexported read-only authority adapter for an existing transactional guard.
- High-risk impact: yes—proposal/lifecycle gate visibility; structural and behavioral REDs prove no writer expansion.
