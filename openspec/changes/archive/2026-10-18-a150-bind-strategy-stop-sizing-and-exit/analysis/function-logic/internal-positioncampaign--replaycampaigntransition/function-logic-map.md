# Function Logic Map: `replayCampaignTransition`

Source `internal/positioncampaign/replay.go`; base 28a800f5, AST `ast.json`; syntax `ast-grep.txt`; risk `risk-pattern-report.md`.

## Inputs and invariants

Preserve legacy event and protection states; new reconciliation blocks entries, never disables valid protective exits or reopens closed campaigns.

## Branches and early returns

| Branch | Line | Existing branch | Regression requirement |
|---|---|---|---|
| B1 | 345 | `if event.Sequence == 1 {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B2 | 346 | `if event.EventKind != "CREATED" \|\| event.CampaignState != CampaignPlanned \|\| event.EntryBlocked {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B3 | 353 | `switch event.EventKind {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B4 | 354 | `case "LEG_PLANNED":` | Existing fixtures plus targeted a150 failure/replay tests. |
| B5 | 356 | `case "ORDER_LINKED":` | Existing fixtures plus targeted a150 failure/replay tests. |
| B6 | 358 | `case "ORDER_LINK_REFUSED":` | Existing fixtures plus targeted a150 failure/replay tests. |
| B7 | 360 | `case "PROSPECTIVE_CANCELLED":` | Existing fixtures plus targeted a150 failure/replay tests. |
| B8 | 362 | `case "AMBIGUOUS_ORDER_FILL":` | Existing fixtures plus targeted a150 failure/replay tests. |
| B9 | 363 | `if state.CampaignState == CampaignClosed {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B10 | 365 | `} else if state.CampaignState == CampaignReconcile {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B11 | 365 | `} else if state.CampaignState == CampaignReconcile {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B12 | 367 | `} else {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B13 | 370 | `case "ORDER_WATERMARK_ADVANCED":` | Existing fixtures plus targeted a150 failure/replay tests. |
| B14 | 371 | `switch {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B15 | 372 | `case state.CampaignState == CampaignClosed:` | Existing fixtures plus targeted a150 failure/replay tests. |
| B16 | 374 | `case event.CampaignState == CampaignClosed:` | Existing fixtures plus targeted a150 failure/replay tests. |
| B17 | 376 | `case event.CampaignState == CampaignReconcile:` | Existing fixtures plus targeted a150 failure/replay tests. |
| B18 | 377 | `if state.CampaignState == CampaignReconcile {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B19 | 379 | `} else {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B20 | 382 | `case state.CampaignState == CampaignExiting:` | Existing fixtures plus targeted a150 failure/replay tests. |
| B21 | 384 | `default:` | Existing fixtures plus targeted a150 failure/replay tests. |
| B22 | 387 | `case "STOP_COMPOSED":` | Existing fixtures plus targeted a150 failure/replay tests. |
| B23 | 393 | `default:` | Existing fixtures plus targeted a150 failure/replay tests. |

## Calls and live bindings

AST calls: TransitionCampaign, LatchEntryBlocked. See source and discovery plan for existing side effects and conditions.

## State mutations and fallbacks

`workingSet` processes existing states before initializing missing ones; quarantine excludes stop judgement. Replay derives transitions from immutable events; a new supported event must preserve sequence and monotone entry blocking.

## Safety conclusion

High-risk yes; branch test mapping is planned, not an assertion of execution.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
