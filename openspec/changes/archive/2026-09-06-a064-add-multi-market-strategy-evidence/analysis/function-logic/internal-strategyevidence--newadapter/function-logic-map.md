# Function Logic Map: `NewAdapter`

- Source: `internal/strategyevidence/source.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `policy` | any `SourcePolicy` — this constructor validated nothing | the caller | none; refusal was deferred to `Fetch` |
| `transport`, `credentials` | any value, `nil` included | the caller | none here |
| `shared` rate budget | left `nil` by this constructor | n/a | `acquire`/`consumeCall` fell back to per-`Adapter` counters |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless: build an `*Adapter` with a real clock, a timer waiter and no shared budget | allocates one `Adapter` | `*Adapter` | `TestSharedRateBudgetCapsEveryAdapterOnOneContract` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Now` (stored, not called) | bind the adapter's clock | none | AST |
| `timerWaiter{}` | bind the default Retry-After waiter | none | AST |

## State mutations and fallbacks

- Allocation only; no I/O, no shared state.
- The fallback that mattered was the absence of one: with `shared == nil`, each `Adapter` counted its own calls, so N adapters on the same official contract gave N × the policy rate.

## Safety conclusion

- This bundle is `revision: base`. The completion pass unexported this constructor to `newAdapter`, so the exported name no longer exists at HEAD and the function's logic is mapped at the frozen comparison base.
- Why unexport rather than add a check: the process-wide budget already exists (`NewOfficialAdapter` requires a non-nil `*SharedRateBudget`). Leaving a second exported door that skips it made 'one budget per official contract' a convention instead of a type-level property, and nothing pinned it (issues.md I17). The rename removes the door; `NewOfficialAdapter` is now the only way in from outside the package.
- Behaviour change outside the package: none — the exported constructor had no caller anywhere in the module (measured 2026-09-07).
