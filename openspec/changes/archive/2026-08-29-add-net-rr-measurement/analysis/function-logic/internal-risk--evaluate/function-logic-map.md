# Function Logic Map: `Evaluate`

- Source: `internal/risk/chain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `add-net-rr-measurement`

## Why this function is in scope

Runs the entry chain and returns the first refusal. This change added one thing to it: each refusal is stamped with the step that produced it, so the observation record can say which rung refused without inferring it from the reason code (a many-to-one map — ReasonInputUnavailable is raised at 42 sites in this package).

**No verdict moved.** `at()` copies the step onto an already-built Decision and returns an ALLOW untouched, so `Allowed`, `Reason` and `Detail` are byte-identical to before for every input. The rung functions were not edited at all.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.Now` | non-zero instant | caller (Guardian uses its clock) | preflight refuses INPUT_UNAVAILABLE |
| `in.Intent` | BUY or SELL, account/symbol set, decimal prices | caller | preflight refuses INPUT_UNAVAILABLE |
| `in.Account` | aggregates and latches the rungs read | caller | individual rungs refuse with their own codes |
| `in.Policy` | `Policy.Validate()` clean | Guardian construction | preflight refuses INPUT_UNAVAILABLE |
| `in.Costs` | `Configured()` true on the entry path | Guardian construction | preflight refuses INPUT_UNAVAILABLE |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`if` @ internal/risk/chain.go:127) | `preflight(in)` refused | none | the refusal, stamped `StepPreflight` | `TestPreflightAndReductionReportDistinctSteps` |
| B2 (`if` @ internal/risk/chain.go:130) | `in.Intent.Side == SideSell` | none | `evaluateReduction`'s verdict, stamped `StepReduction` when refused | `TestPreflightAndReductionReportDistinctSteps` |
| B3 (`range` @ internal/risk/chain.go:133) | iterating the 12 `entryChain` rungs in order | none | falls through to `allow()` when none refuse | `TestEveryRungReportsItsOwnName` |
| B4 (`if` @ internal/risk/chain.go:134) | a rung refused | none | that rung's refusal, stamped `s.name` | `TestEveryRungReportsItsOwnName`, `TestFirstFailureStopsTheChain` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `preflight` | usability of the input before any policy judgement | pure; no I/O, no timeout | AST B1 + chain.go |
| `evaluateReduction` | the reduce-only branch | pure | AST B2 |
| `s.eval` (12 rungs) | the policy judgement itself | pure; each returns a Decision | AST B3/B4 + `EntryChainSteps` |
| `at` | stamps the step onto a refusal | pure; returns an ALLOW unchanged | AST B1/B2/B4 |

## State mutations and fallbacks

- None. `Evaluate` is a pure function: it takes `Input` by value, mutates no field of it, and writes nothing. The new `at()` call operates on the returned `Decision` value, not on `in`.
- No side effects, no I/O, no clock read (the instant is an input), no fallback path.
- No live binding. `internal/risk` imports neither the journal nor execgw, which `execgw/observation_scope_test.go:TestTheRiskPackageStaysFreeOfTheJournal` enforces by scan.

## Safety conclusion

- Safe edit boundary: the stamp is applied at the single return site of each arm, after the verdict exists. Nothing reads `Step` to decide anything — only the observation record writes it.
- High-risk impact: **yes** (Guardian path). §0.9 is not engaged: no ALLOW became a REFUSE or the reverse. The evidence is the pre-existing chain suite passing **unmodified** (`internal/risk`, 127 tests) plus `TestTheStepIsAdditiveToTheVerdict`.
- Upstream inheritance: none. This chain is TossOS-new; the 650 inherited tests do not reach it.
