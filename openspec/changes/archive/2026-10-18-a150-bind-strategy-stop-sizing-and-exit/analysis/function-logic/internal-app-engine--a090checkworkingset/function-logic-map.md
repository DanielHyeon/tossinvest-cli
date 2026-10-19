# Function Logic Map: `a090CheckWorkingSet`

Source: `internal/app/engine/a090_workingset_exits_test.go`. Base AST retained: the helper body is unchanged and the adjacent comment is classified against the base.

## Inputs and invariants

The unchanged helper compares every loop exit and verifies mark/unmark placement. The caller's census now pins the approved post-mark exit.

## Branches and early returns

Base AST enumerates the unchanged helper branches; the existing mutation tests remain enabled.

## Calls and live bindings

Test-only AST parser and census; no production or network operations.

## State mutations and fallbacks

Only local temporary test values.

## Safety conclusion

Helper unchanged; `/tmp/tossos-a150-census-green.log` records census and mutation checks. Full gate repeat pending.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
