# Function Logic Map: `ValidateCommand`

Source `internal/positioncampaign/command.go`; base 28a800f5, AST `ast.json`; syntax `ast-grep.txt`; risk `risk-pattern-report.md`.

## Inputs and invariants

Preserve legacy event and protection states; new reconciliation blocks entries, never disables valid protective exits or reopens closed campaigns.

## Branches and early returns

| Branch | Line | Existing branch | Regression requirement |
|---|---|---|---|
| B1 | 23 | `switch CommandKind(kind) {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B2 | 24 | `case CommandCreate, CommandPlanLeg, CommandLinkOrder, CommandCancelProspective,` | Existing fixtures plus targeted a150 failure/replay tests. |
| B3 | 26 | `default:` | Existing fixtures plus targeted a150 failure/replay tests. |
| B4 | 29 | `if len(key) == 0 \|\| len(key) > 128 {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B5 | 32 | `for i := 0; i < len(key); i++ {` | Existing fixtures plus targeted a150 failure/replay tests. |
| B6 | 34 | `if (c >= 'a' && c <= 'z') \|\| (c >= 'A' && c <= 'Z') \|\|` | Existing fixtures plus targeted a150 failure/replay tests. |

## Calls and live bindings

AST calls: CommandKind, fmt.Errorf, len. See source and discovery plan for existing side effects and conditions.

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
