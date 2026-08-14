# Function Logic Map: `Journal.RefreshExitObservation`

- Source: `internal/journal/exit_observation_refresh.go`
- Post-edit AST evidence: `ast.json` (26 branches; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| refresh candidate | complete provenance-bound, nonorderable EVALUATED line | evaluator output plus stored working set | typed invalid/conflict/stale; never proposal/order authority |
| observation identity | `quote_fetched_at` or positive `cycle:N`, nonzero time | observer validation and durable tuple | malformed/older/ambiguous evidence is no-write |
| lifecycle/proposal state | current managed generation, not completed, no pending proposal | journal transaction | conflict or completed error before update |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `41`: blank position id | reject invalid request before transaction | fail-closed result | `TestA111RefreshRejectsSeedAndEveryStrongerOrIncompatibleCandidateWithoutWriting` |
| B2 | `44`: provenance validation or zero provenance | reject incomplete provenance before transaction | fail-closed result | `TestA111RefreshRejectsSeedAndEveryStrongerOrIncompatibleCandidateWithoutWriting` |
| B3 | `45`: provenance validator returned an error | propagate validation error | fail-closed result | `TestA111RefreshRejectsSeedAndEveryStrongerOrIncompatibleCandidateWithoutWriting` |
| B4 | `50`: request provenance differs from snapshot provenance | reject tuple mismatch before transaction | fail-closed result | `TestA111RefreshRejectsSeedAndEveryStrongerOrIncompatibleCandidateWithoutWriting` |
| B5 | `56`: observed_at is zero | reject missing evidence time | fail-closed result | `TestA111RefreshRejectsSeedAndEveryStrongerOrIncompatibleCandidateWithoutWriting` |
| B6 | `60`: observation source is neither official nor valid cycle:N | reject malformed temporal identity | fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B7 | `63`: snapshot is explicitly orderable or carries any executable proposal, including an exact replay | return conflict before generic snapshot validation and before opening a transaction | fail-closed result | `TestA111RefreshExplicitlyRejectsAnIdenticalExecutableSnapshotWithoutWriting` |
| B8 | `76`: complete candidate fails judgement/snapshot validation | reject invalid tuple before transaction | fail-closed result | `TestA111RefreshRejectsSeedAndEveryStrongerOrIncompatibleCandidateWithoutWriting` |
| B9 | `80`: BEGIN IMMEDIATE fails | return storage error; no write | fail-closed result | `TestA111RefreshFailureRollsBackTheWholeTuple` |
| B10 | `85`: current exit progress cannot be read | rollback and propagate error | fail-closed result | `TestA111RefreshFailureRollsBackTheWholeTuple` |
| B11 | `88`: current exit state is completed | reject without reopening | fail-closed result | `TestA111RefreshRejectsCompletedStateWithoutReopeningIt` |
| B12 | `91`: current effective snapshot is absent/SEED | reject non-evaluated state | fail-closed result | `TestA111RefreshRejectsSeedAndEveryStrongerOrIncompatibleCandidateWithoutWriting` |
| B13 | `95`: read-only `exitObservationRefreshGuardTx` status/pending query fails | rollback and propagate error; guarded columns remain named only in `apply_hook.go` | fail-closed result | `TestA111RefreshFailureRollsBackTheWholeTuple`, `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` |
| B14 | `98`: helper reports status is not EVALUATED or a proposal is pending | reject refresh without touching proposal | fail-closed result | `TestA111RefreshRejectsARealCurrentPendingProposalAndPreservesItsEvidence` |
| B15 | `103`: caller lifecycle generation is zero | bind expected generation to current durable generation | fail-closed result | `TestA111RefreshPreservesTheExistingPolicyAndGenerationOnAValidCandidate` |
| B16 | `106`: expected generation differs from current exit generation | reject lifecycle conflict | fail-closed result | `TestA111RefreshRejectsSeedAndEveryStrongerOrIncompatibleCandidateWithoutWriting` |
| B17 | `113`: no lifecycle row exists | use legacy managed generation-one compatibility fallback | fail-closed result | `TestA111RefreshAtomicallyReplacesTheCompleteTupleWithoutAnEvent` |
| B18 | `115`: lifecycle query did return a row/error path | distinguish compatibility fallback from query error | fail-closed result | `TestA111RefreshRejectsARealReleasedLifecycleAndPreservesItsTuple` |
| B19 | `115`: lifecycle query returned a non-NoRows error | rollback and propagate error | fail-closed result | `TestA111RefreshFailureRollsBackTheWholeTuple` |
| B20 | `118`: latest lifecycle is released/unmanaged or generation changed | reject without writing | fail-closed result | `TestA111RefreshRejectsARealReleasedLifecycleAndPreservesItsTuple` |
| B21 | `122`: temporal CAS comparator rejects candidate | propagate stale/conflict; no tuple write | fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B22 | `125`: candidate is an exact evidence replay | true no-op including updated_at | fail-closed result | `TestA111RefreshIsIdempotentAndDurableAcrossRestart` |
| B23 | `128`: position generation or any operational semantic field differs | reject stronger/incompatible candidate | fail-closed result | `TestA111OperationalEqualityChecksEveryD1FieldUsingEvaluatorGeneratedDonors` |
| B24 | `133`: complete snapshot encoding fails | rollback; no flattened/JSON partial write | fail-closed result | `TestA111RefreshFailureRollsBackTheWholeTuple` |
| B25 | `149`: atomic exit_states UPDATE fails | rollback all tuple fields | fail-closed result | `TestA111RefreshFailureRollsBackTheWholeTuple` |
| B26 | `152`: post-update write hook fails | rollback transaction | fail-closed result | `TestA111RefreshFailureRollsBackTheWholeTuple` |
| Return | every admitted replacement | commit one complete JSON+flattened tuple; no event/proposal/order | exact transaction result | `TestA111RefreshAtomicallyReplacesTheCompleteTupleWithoutAnEvent` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ExecutableProposal` | explicit authority firewall before generic validation | any executable result returns conflict before transaction | B7 + exact-replay RED |
| `validateJudgementSnapshot` | validate complete persisted tuple | invalid request; no transaction | B8 |
| `scanExitProgress` | read current complete exit progress inside `BEGIN IMMEDIATE` | transactional rollback on read error/conflict | B10-B12 |
| `exitObservationRefreshGuardTx` | read snapshot status and derive a pending boolean without naming guarded columns in the refresh file | read-only query in `apply_hook.go`; error rolls back, pending/status mismatch conflicts | B13-B14 + structural writer-authority RED |
| lifecycle query | bind latest managed lifecycle generation | released/generation mismatch conflicts; query error rolls back | B15-B20 + released/lifecycle RED |
| `compareObservationEvidence` | temporal CAS and exact replay classification | stale/conflict/no-op/replace only | B21-B22 |
| `sameExitOperationalLine` | prohibit semantic change on heartbeat path | conflict; no write | B23 |
| `encodeStoredSnapshot` / `ExecContext` / hook / commit | replace JSON and flattened tuple atomically | any error rolls back | B24-B26 |

## State mutations and fallbacks

- The executable/orderable guard is deliberately before generic snapshot validation, so even an identical executable replay returns the refresh-specific conflict and never reaches a transaction.
- Only the final UPDATE mutates non-guarded observation state. It replaces every flattened observation field and `effective_snapshot_json` together; it appends no `exit_events` and never touches proposal state.
- The refresh file no longer names `pending_action`. Its unexported helper lives beside the apply-hook authority, returns only status plus a derived boolean, and issues no write; `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` preserves the package-wide single-writer boundary.
- Equal evidence identity is a true no-op, and lifecycle/pending/semantic conflicts leave the prior tuple byte-for-byte intact.

## Safety conclusion

- Safe edit boundary: observation-only heartbeat of an already EVALUATED, exact operationally equal, nonorderable line.
- High-risk impact: yes — journal and order-authority boundary; B7 is the explicit firewall.
