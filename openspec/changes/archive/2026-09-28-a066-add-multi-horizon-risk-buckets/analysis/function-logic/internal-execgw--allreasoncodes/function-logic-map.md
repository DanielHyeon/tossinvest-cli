# Function Logic Map: `AllReasonCodes`

- Source: `internal/execgw/failclosed.go`
- AST evidence: `ast.json` (5.5 pre-edit, extracted 2026-09-27 at HEAD `c1d1e295`; lines 254–302, 0 branches, 2 returns —
  the sort comparator's and the function's)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| none | — | the literal list in the body | — |
| output | every reason code this build emits, sorted, unique | `testdata/reason_codes.golden` via `TestReasonCodeEnumIsStable` | golden diff fails the test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path — no branch (AST: 0) | none | sorted slice | `TestReasonCodeEnumIsStable` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sort.Slice` | stable golden order | none | AST |

## State mutations and fallbacks

- Pure. No state.

## Safety conclusion

- Safe edit boundary (5.5): append the new `ReasonEntryLossLockActive` to the literal list and regenerate the golden
  with the sanctioned generator (`TOSSOS_UPDATE_GOLDEN=1 go test -run TestWriteReasonCodeGolden`). A code the Gateway
  can emit but the list omits is the pre-existing gap recorded in the list's own comment ("declared … but never
  registered"); the new code is not added to that gap.
- High-risk impact: no behaviour — vocabulary pin only.
