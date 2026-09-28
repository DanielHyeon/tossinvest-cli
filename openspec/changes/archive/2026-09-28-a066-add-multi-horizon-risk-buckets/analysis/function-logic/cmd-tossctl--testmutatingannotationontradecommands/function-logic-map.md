# Function Logic Map: `TestMutatingAnnotationOnTradeCommands`

- Source: `cmd/tossctl/help_convention_test.go`
- AST evidence: `ast.json` (post-edit, 96–150, 3 branches)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `wantMutating` | the command paths that must declare `mutating: true` | this test (a frozen list) | a listed command without the annotation fails; an unlisted command with it fails |

## Branches and early returns

| Branch | Condition (AST line) | Effect | Test |
|---|---|---|---|
| B1 | range over every leaf command (140) | — | this test |
| B2 | listed but not mutating (143) | `t.Errorf` | this test (mutation C06: `entry-lock-release` declared false → RED) |
| B3 | mutating but not listed (146) | `t.Errorf` | this test |

The a066 5.5 edit is data only: two entries, `tossctl engine entry-lock-release` and `tossctl engine risk-latch-release`
(user decision 2026-09-28: "tossctl mutating 명령"). The `engine run` line was re-aligned by gofmt. No branch changed.

## Calls and live bindings

- `leafCommands(newRootCmd())` — the real command tree.

## State mutations and fallbacks

- None (test).

## Safety conclusion

- Test-only edit. It widens the frozen list by exactly the two release commands; `risk-latch-show` stays off the list.
