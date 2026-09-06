# Function Logic Map: `insertExactStrategyDecision`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `lineage.ConsumedEvidenceSnapshotID/Digest` | both empty, or `snapshot-<64 lowercase hex>` paired with that digest | `validConsumedEvidenceReference` | typed `ErrStrategyEvidenceReferenceInvalid` before any statement runs |
| `lineage` (remaining fields) | the caller's sealed decision lineage | `StrategyAtomicPlan` | read-back mismatch is a `StrategyCollisionError` |
| `tx` | an open journal transaction at schema >= 21 | `Journal.Open` | statement errors surface unwrapped so the caller rolls back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | consumed evidence reference is partial or malformed | none — refused before the INSERT | `ErrStrategyEvidenceReferenceInvalid` | `TestStrategyEvidenceGoGuardRefusesBeforeSQL` |
| B2 | the lineage INSERT itself fails | none committed; the caller rolls the transaction back | the driver error, unwrapped | `TestFirstLegAtomicAdmissionLateStatementFailuresRollbackEveryFamily` |
| B3 | the read-back does not match the requested lineage exactly | none committed | `StrategyCollisionError{Stage: "decision lineage"}` | `TestStrategyEvidenceLineageReplayIsExact` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validConsumedEvidenceReference` | refuse a partial or malformed snapshot reference in Go, before the storage layer sees it | pure; no I/O | AST + `TestStrategyEvidenceGoGuardRefusesBeforeSQL` |
| `tx.ExecContext` (INSERT OR IGNORE) | append the decision lineage idempotently | driver error returned unwrapped; the v21 triggers may abort it | AST + `TestV21TriggerRefusesEveryMalformedReferenceShape` |
| `tx.QueryRowContext` | read the row back and compare every column | `sql.ErrNoRows` and mismatches both become a collision | AST + `TestStrategyEvidenceLineageReplayIsExact` |

## State mutations and fallbacks

- One `INSERT OR IGNORE` into `strategy_decision_lineage`, inside the caller's transaction.
- No broker, dispatch, Guardian, toggle or evidence-store call occurs.
- There is no fallback: a refused reference never becomes NULL/NULL, and a divergent replay is never overwritten.

## Safety conclusion

- High-risk: this is the row an entry decision's evidence lineage is read from later.
- The completion pass changed one thing — the Go refusal now carries `ErrStrategyEvidenceReferenceInvalid` instead of an anonymous `errors.New`. The v21 SQL triggers refuse the same six inputs, so an untyped refusal made the two layers indistinguishable and the Go guard deletable in silence (issues.md I4).
- Deleting the guard now fails `TestStrategyEvidenceGoGuardRefusesBeforeSQL`; measured under `go test -overlay` on 2026-09-07.
