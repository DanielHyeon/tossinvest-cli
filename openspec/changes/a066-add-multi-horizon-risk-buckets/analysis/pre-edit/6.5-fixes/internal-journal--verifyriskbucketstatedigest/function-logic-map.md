# Function Logic Map: `verifyRiskBucketStateDigest`

- Source: `internal/journal/risk_bucket_fill.go`
- AST evidence: `ast.json` (pre-edit, extracted 2026-09-28 at HEAD `4fb0b26e`, 1093–1106, 3 branches; copy in `analysis/pre-edit/6.5-fixes/`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| owner key | canonical | caller | — |
| reconstructed state digest vs last sealed `state_digest` | equal | `loadRiskBucketState` vs `risk_bucket_state_snapshots` | mismatch → `ErrRiskBucketReplayMismatch`; **missing seal → raw `sql.ErrNoRows`** |

## Branches and early returns (pre-edit)

| Branch | Condition (AST source line) | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | reconstruction error (1095) | none | err (already typed by `loadRiskBucketState`) | fill-layer tests |
| B2 | seal query error (1099) — includes `sql.ErrNoRows` when no seal row exists | none | **raw err** | none before 6.5 — `TestA066FillWithAMissingStateSealLatchesAndKeepsTheFill` (RED at 4fb0b26e: the fill transaction aborts) |
| B3 | digest mismatch (1102) | none | `ErrRiskBucketReplayMismatch` | state digest drift tests |

## Calls, callers and the planned edit

Callers (grep internal/journal non-test): admission (CRA :167, commitFresh :443), revalidation (:593, `verifyQFinalAdmissionRows` :363), fill apply (:217), terminal release (:268), actual completion (:312), release (:373), order registration (:114), owner bind (owner.go:393; the caller at :547 latches on semantic **or** `sql.ErrNoRows`), owner release (owner.go:846), and strategy registration (strategy_dispatch_runtime.go:1116). Every non-fill caller refuses on any error. Only the fill callers branch on `isRiskBucketSemanticError`.

Planned 6.5 edit (finding 4): a missing seal (`sql.ErrNoRows`) becomes `ErrRiskBucketReplayMismatch` ("state seal missing"). On the fill path this becomes a latch and the fill commits (the 2.7 contract). Elsewhere it is still a refusal, so the outcome is unchanged there. Other errors stay raw (storage). R2's other two raw returns (`riskBucketFillEventDigest` and the held read in `releaseRiskBucketOrderInTx`) sit behind this check on every path (walked: applyRiskBucketFillInTx :217 before :243; releaseTerminal :268 and ReleaseRiskBucketOrder :373 before `releaseRiskBucketOrderInTx`). They are backstops and not edited.

## Safety conclusion

- Safe edit boundary: the named branches only; entry refusal / fill latch in the conservative direction.
- High-risk impact: yes (admission / submit revalidation / fill path).
