# Function Logic Map: `insertExactStrategyDecision`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | active caller-owned strategy issuance transaction | `planStrategyEntryForTest`, `RecordStrategyDecisionAndReserve` | SQL error rolls back caller transaction |
| `lineage` | complete normalized immutable strategy decision | a047 strategy lineage contract | exact collision error, never overwrite |
| consumed evidence reference | both snapshot ID and digest blank, or both present and bounded | a064 snapshot-only journal requirement | reject incomplete pair before insert |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (new) | only one of snapshot ID/digest is present, pair is malformed or identity is unbounded | no SQL statement | invalid lineage error | `TestStrategyEvidenceLineageRejectsPartialOrMalformedReference` |
| B2 | `INSERT OR IGNORE` returns error | no committed mutation | SQL error | existing atomic rollback tests |
| B3 | row read fails or any persisted preimage field differs | ignored insert remains inside caller transaction | `StrategyCollisionError` | exact replay and divergent snapshot tests |
| success | row is newly inserted or byte-exact replay | nullable ID/digest inserted with decision | rows affected 1 or 0 | `TestStrategyEvidenceLineagePersistsOnlyImmutableReference` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `tx.ExecContext` | atomic insert-or-ignore within existing risk reservation transaction | context/error returned; caller rolls back | CodeGraph callers + AST |
| `tx.QueryRowContext(...).Scan` | compare every immutable field after insert-or-ignore | mismatch is collision, no fallback | strategy lineage exact replay tests |
| `RowsAffected` | distinguish first insert from idempotent replay | advisory count only after exact compare | existing idempotency tests |

## State mutations and fallbacks

- Existing strategy decision rows are immutable by trigger.
- The edit adds two nullable scalar columns to the same insert/compare preimage.
- No payload, source response, revision, credential, Guardian, broker or toggle field is introduced.
- There is no fallback from a missing snapshot reference to `evidence_digest` or journal payload.

## Safety conclusion

- Safe edit boundary: append the exact optional pair to insert and equality verification; preserve transaction ownership and existing collision semantics.
- High-risk impact: yes — journal lineage is high-risk, mitigated by additive-nullable v21 migration, RED atomic/replay tests and unchanged dispatch authority.
