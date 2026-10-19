# Function Logic Map: `ExitObserver.record`

Source: `internal/app/engine/exitloop.go`. The pre-edit AST is `pre-edit-ast.json`; review and scope approval pending.

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


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
