# Function Logic Map: `Journal.recordExitJudgementTx`

- Source: `internal/journal/exit_state.go`
- Frozen span: lines 392-625 at HEAD `3355df0fe9c82c3bb8c522e2d79abf107dd5f2c3`
- AST evidence: `ast.json` (51 branches, 69 calls, 31 returns)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| position/lifecycle/generation | nonblank, current managed lifecycle, exact generation | journal row + caller snapshot | typed invalid/stale/completed error before write |
| provenance/proposal | complete matching tuple or wholly legacy zero | immutable evaluator snapshot | reject partial/cross-snapshot proposal |
| saved/recomputed snapshot | decodable, digest-valid, monotone-selectable | effective JSON + judgement | corruption error or atomic quarantine |
| JSON/flattened tuple | exact same effective snapshot | encoder + one UPDATE statement | transaction rollback on any failure |
| proposal/event | arm only selected effective proposal; append one judgement history row | transaction | rollback together; commit once |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| V1-V7 | blank ID, provenance/suppression/proposal invalid or mismatched | none | validation error | identity/proposal tests |
| V8 | full provenance | build and validate recomputed stored snapshot | validation error or continue | snapshot validation tests |
| T1-T6 | begin/read/completed/lifecycle/generation failures | open tx rolls back | typed/storage error | lifecycle and rollback tests |
| D1-D3 | duplicate decision | no write | `ErrProposalPending` | duplicate decision test |
| L1-L2 | legacy scalar path descends | no write | monotonicity error | descending baseline/watermark tests |
| R1-R3 | recovery selection ambiguous | quarantine + commit | `ErrExitSnapshotQuarantined` | ambiguous recovery/quarantine tests |
| R4 | saved monotone wins | select saved, strip recomputed proposal/suppression | continue | saved-monotone test |
| R5 | recomputed wins | copy selected recomputed tuple | continue | persistence test |
| W1 | effective snapshot present | encode digest JSON; one UPDATE writes scalar + flattened + JSON | error rolls back | tuple-integrity tests |
| W2 | legacy snapshot absent | scalar UPDATE only | error rolls back | legacy compatibility tests |
| W3-W5 | hook after state/arm/event fails | rollback all prior stages | hook error | staged rollback test |
| A1 | proposal present | arm inside tx | arm error rolls back | proposal/crash tests |
| E1 | always | append event, including evaluation candidates when complete | error rolls back | history tests |
| C1 | commit fails/succeeds | atomic durable boundary | wrapped error or result projection | crash test |
| O1-O3 | saved / armed / suppressed result | none after commit | typed outcome | outcome tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| provenance and snapshot validators | reject partial/forged evidence | before transaction/write | CodeGraph callees + AST |
| `scanExitProgress` | read JSON and generation | corruption is typed and fail-closed | CodeGraph/AST |
| `SelectRecoverySnapshot` | monotone whole-candidate selection | ambiguity quarantines and commits | CodeGraph/AST |
| `encodeStoredSnapshot` | validate and seal JSON with output digest | failure aborts tx | CodeGraph/AST |
| `tx.ExecContext(updateSQL,args...)` | atomically replace all effective representations | one statement | source lines 551-577 |
| `armExitProposalTx` / `appendExitEventTx` | proposal and history in same tx | any error rolls back | CodeGraph/AST |
| `tx.Commit` | sole ordinary durability boundary | post-commit result only | AST |

## State mutations and fallbacks

- Flattened fields and `effective_snapshot_json` are co-written in one UPDATE, not separately.
- Every ordinary judgement appends an event; therefore A111 narrow refresh cannot reuse this function unchanged.
- A refresh transaction must re-read status/generation/current effective evidence, validate exact operational equality, replace JSON and every flattened column atomically, and never arm/append.
- Conflict must be typed; silent fallback to full judgement would risk stale overwrites or duplicate history.

## Safety conclusion

- Safe edit boundary: add a sibling narrow transaction reusing validation/encoding helpers; do not insert a refresh flag into 51-branch proposal transaction without independent proof.
- High-risk impact: yes — authoritative protection state, crash consistency, quarantine, and order arming.
