# Function Logic Map: `TestApplyFillRetryIsIdempotent`

- Source: `internal/riskbucket/fill_test.go`
- AST evidence: `ast.json` (at the landing `e9355a82` = HEAD content, 32–47, 3 branches)
- Risk scan: `risk-pattern-report.md`

## Why this bundle exists

a066 commit `4a364caf` ("aggregate risk fills across owner decisions") inserted a new test right after this function.
The diff hunk is `@@ -48,0 +49,27 @@`, a pure insertion anchored on the blank line after the closing brace. Step 5 of
the gate counts that insertion against this function. **The function body is byte-identical between the old base
`23794f86` and the landing `e9355a82`** (lines 32–47 compared, 2026-09-29). This bundle records that; there was no
edit to map.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `fillFixture("100","50")` event at cumulative 4 with actual evidence | fixture | this test | — |

## Branches and early returns

| Branch | Condition (AST line) | Effect | Test |
|---|---|---|---|
| B1 | first `ApplyFill` error (37) | `t.Fatal` | this test |
| B2 | retry `ApplyFill` error (41) | `t.Fatal` | this test |
| B3 | retry not a duplicate or state changed (44) | `t.Fatalf` | this test |

## Calls and live bindings

- `ApplyFill` (the riskbucket fill reducer) twice with the same event; `reflect.DeepEqual` on the two states.

## State mutations and fallbacks

- None (pure test on value state).

## Safety conclusion

- No a066 edit to this function. It pins that a retried fill is a no-op duplicate.
