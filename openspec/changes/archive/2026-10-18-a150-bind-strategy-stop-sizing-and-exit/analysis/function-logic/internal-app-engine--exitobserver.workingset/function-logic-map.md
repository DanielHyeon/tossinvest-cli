# Function Logic Map: `ExitObserver.workingSet`

Source `internal/app/engine/exitloop.go`; base 28a800f5, AST `ast.json`; syntax `ast-grep.txt`; risk `risk-pattern-report.md`.

## Inputs and invariants

Preserve legacy event and protection states; new reconciliation blocks entries, never disables valid protective exits or reopens closed campaigns.

## Branches and early returns

| Branch | Line | Existing branch | Regression requirement |
|---|---|---|---|
| B1 | 548 | `if err != nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B2 | 552 | `if err != nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B3 | 556 | `for _, result := range stateResults {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B4 | 561 | `for _, p := range positions {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B5 | 562 | `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B6 | 565 | `if !p.ExitEligible() {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B7 | 580 | `if !ok {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B8 | 582 | `if err != nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B9 | 583 | `if cycle.Err == nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B10 | 588 | `if opened.PositionID == "" {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B11 | 599 | `if result.Corruption != nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B12 | 602 | `if qerr != nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B13 | 603 | `if cycle.Err == nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B14 | 613 | `if q, active, qerr := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq); qerr != nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B15 | 618 | `} else if active && !q.NeedsReJudgement() {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B16 | 614 | `if cycle.Err == nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B17 | 618 | `} else if active && !q.NeedsReJudgement() {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B18 | 622 | `} else if active {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B19 | 622 | `} else if active {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B20 | 646 | `if identityErr != nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B21 | 649 | `if qerr != nil {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B22 | 650 | `if cycle.Err == nil {` | Existing fixtures plus targeted a150 failure/replay tests. |

## Calls and live bindings

AST calls: o.opts.Journal.Positions, o.opts.Journal.OpenExitStateResults, make, len, isZeroQuantity, p.ExitEligible, o.alertUnmanaged, o.markHeld, int64, o.openState, o.unmarkHeld, o.opts.Journal.QuarantineExitSnapshot, result.Corruption.Error, append, fmt.Errorf, o.announceQuarantine, o.opts.Journal.ActiveExitSnapshotQuarantine, q.NeedsReJudgement, o.log, managedPolicyIdentity, p.Adopted, identityErr.Error. See source and discovery plan for existing side effects and conditions.

## State mutations and fallbacks

`workingSet` processes existing states before initializing missing ones; quarantine excludes stop judgement. Replay derives transitions from immutable events; a new supported event must preserve sequence and monotone entry blocking.

## Safety conclusion

High-risk yes; branch test mapping is planned, not an assertion of execution.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

## Current implementation clarification

Current implementation verifies admission and position generation before using a policy; mismatches attempt a campaign latch and preserve an existing exit. For a verified same-generation admission it evaluates the greater of the current baseline and admitted stop without changing entry, initial R, policy or pending order. Tests cover foreign generation, failed latch, stop breach, restart and pending sale; snapshot selector refusal remains an error, never a fabricated recovery.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
