# Function Logic Map: `newConsoleOptimizationCommander`

- Source: `cmd/tossctl/optimization.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context | caller-owned startup context | Cobra `runConsole` | registry/store open error is returned; no retry |
| active profile directory | directory containing the selected journal path | `consoleJournalPath` | empty/unresolvable path prevents construction |
| evidence provider | narrow `optimization.EvidenceProvider`; may be nil | `internal/optimizationevidence` over `performance.Dashboard` | nil remains explicit unavailable evidence and cannot authorize an evidence-backed candidate |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | core registry construction fails | none | return error before control DB open | registry/store package tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyopt.CoreRegistry` | assemble owner-controlled finite option registry | one attempt; validation error returned | CodeGraph + AST |
| `strategyopt.Open` | open the separate optimization control DB with the supplied read-only evidence capability | one attempt; error returned | CodeGraph + AST + cmd wiring tests |

## State mutations and fallbacks

- Opens only `optimization-control.db`; it never opens or writes `journal.db`.
- The evidence dependency exposes only `ReadEvidence`; it cannot collect performance, mutate the journal, trade, or change LIVE/lane/gate state.
- Missing evidence remains visible and fail-closed while server-defined conservative preset preview remains governed by the existing lifecycle contract.

## Safety conclusion

- Safe edit boundary: add a narrow evidence-provider argument to the existing store construction.
- High-risk impact: yes; evidence can gate recommendations, so missing/error states must remain unavailable and no trading authority may be added.
