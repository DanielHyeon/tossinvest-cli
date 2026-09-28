# Function Logic Map: `TestTheReadOnlyHandleHasNoWriteMethods`

- Source: `internal/journal/readonly_test.go`
- AST evidence: `ast.json` (post-edit, 116–162, 2 branches)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `allowed` | the method names `*ReadOnly` may expose | this test (a frozen list) | an unlisted method fails |

## Branches and early returns

| Branch | Condition (AST line) | Effect | Test |
|---|---|---|---|
| B1 | for every method of `*ReadOnly` (155) | — | this test |
| B2 | method not in `allowed` (157) | `t.Errorf` | this test |

The a066 5.5 edit is data only: `ReadEntryLossLocks` and `ReadRiskOwnerLatches`, the two SELECT-only readers that
`engine risk-latch-show` uses. No branch changed.

## Calls and live bindings

- `reflect.TypeOf(&ReadOnly{})` — the real method set.

## State mutations and fallbacks

- None (test).

## Safety conclusion

- Test-only edit. The two readers issue SELECTs on the `mode=ro` + `query_only` connection. They refuse a pre-v35
  journal with `ErrSchemaTooOld` (`TestReadOnlyRelaxationViewsRefuseAPreV35Journal`). The releases themselves exist only
  on the engine-owned `*Journal`.
