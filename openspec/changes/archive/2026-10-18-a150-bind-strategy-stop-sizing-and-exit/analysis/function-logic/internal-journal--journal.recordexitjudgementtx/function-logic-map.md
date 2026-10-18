# Function Logic Map: `Journal.recordExitJudgementTx`

Source: `internal/journal/exit_state.go`. The pre-edit AST is `pre-edit-ast.json`; review and scope approval pending.

## Inputs and invariants

Saved and recomputed snapshots must retain consistent generation, policy and stop provenance; proposals arm atomically before submission. Pending broker orders cannot be cancelled as a side effect of a stale candidate.

## Branches and early returns

Full condition inventory is in the current AST and branch map. A stale floor candidate against a newer valid saved snapshot currently enters the ambiguous selector and commits quarantine; /tmp/tossos-a150-stale-floor-red.log reproduces it.

## Calls and live bindings

The observer clears working orders before the journal write; the journal validates provenance, checks lifecycle, chooses a recovery snapshot and either commits judgement/proposal or quarantine. No mutex may span network calls.

## State mutations and fallbacks

Judgement, state, proposal and exit event are transactional; the existing selector commits quarantine on ambiguous comparison. A new stale-conflict path must leave the valid old snapshot and pending protection intact.

## Safety conclusion

High risk. Parent design decision required before edits.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.



## Implemented resolution

The approved floor-specific expected snapshot and evaluator-input tuple CAS rejects a stale candidate before selection; floor-specific selector refusal also rolls back without quarantining. Floor reconciliation is suppressed before any order clear, so rejected writes cannot cancel pending protection. A new outcome distinguishes this suppression. The first order decision is deferred to the next actual observation; the normal path then reevaluates. See `/tmp/tossos-a150-floor-final.log`.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

The tuple includes legacy watermark, rung/level, quantity, taken ratio and pending proposal fields; an empty snapshot ID alone does not fence legacy writers. RED/GREEN: `/tmp/tossos-a150-seed-cas-red.log`, `/tmp/tossos-a150-seed-cas-green.log`.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

The floor CAS now reuses the existing `scanExitState`/`exitStateSelect` reader in the same transaction; guarded fill fields still have exactly one writer. The first aggregate gate exposed the prior direct-read structural guard failure; the unchanged guard and the floor race tests pass after this repair.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
