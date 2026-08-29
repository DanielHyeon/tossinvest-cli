# Function Logic Map: `ReconcileDriver.blocked`

- Source: `internal/app/engine/adoption.go`
- Frozen source span: lines 159–164 at base `3615f793c4a9dbe027fdbe88b3ed01e140b05cc9`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md` (no configured pattern matched)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market`, `symbol` | normalized downstream by tracker gate lookup | authoritative holding candidate | covered block defers adoption before any quote read |
| `d.opts.Tracker` | nil only in isolated/no-tracker configuration | engine wiring | nil means no reconciliation block is asserted by this helper |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | tracker is nil | none | `false` | no-tracker driver test |
| Return 1 | tracker has any permanent account block | none | `true` for every market/symbol | `TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile` |
| Return 2 | no permanent block; scoped `EntryAllowed` rejects this candidate | none | `true` for covered candidate, otherwise `false` | reconcile-loop adoption-block tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Tracker.Permanent` | account-wide permanent precedence | in-memory read under tracker mutex | AST return expression |
| `Tracker.EntryAllowed` | checks ordinary scoped block | returns rejection object or nil; no broker I/O | AST return expression |

## State mutations and fallbacks

- Pure predicate; it neither releases reconciliation nor changes adoption settings.
- A permanent promotion produced by `Tracker.Observe` therefore blocks every automatic-adoption candidate before price observation. This is the direct link from the a110 defect to missing adoption baselines.

## Safety conclusion

- Safe edit boundary: **not edited by a110**. It is mapped as supporting high-risk evidence and must remain fail-closed.
- High-risk impact: **yes**, because weakening it would allow adoption under an unresolved account disagreement. a110 changes only what evidence earns `Permanent()`.
