# Function Logic Map: `BudgetCoordinator.observeLocked`

- Source: `internal/scheduler/budget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| observation | endpoint path, allowance/reset provenance, response timestamp | `official.RateBudget` | older/invalid evidence cannot gain authority |
| cycle record | nil for manual ingestion or validated one-shot record bound by `ObserveCycle` | coordinator cycle registry | nil has no reconcile/generation/reset authority |
| endpoint state | observation, commitments, issuance memories, trusted reset/generation | coordinator mutex | conflicts fail closed; generation never wraps |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | endpoint exists and incoming observation is older | none | return | out-of-order test |
| B2 | timestamp equal but provenance differs | marks provenance conflict | return | equal-time conflict test |
| B3 | timestamp equal with identical provenance | may lower remaining; valid same-window cycle may reconcile | return | conservative correction/cycle tests |
| B4 | newer trusted observation is initial/same/next/conflict relation | initializes, reconciles, conditionally advances, or marks conflict | continue/return | reset relation tests |
| B5 | next window has nil cycle | preserves generation and both issued memories even with empty commitments | continue storing evidence | manual-cap-bypass test |
| B6 | next window has valid cycle | reconciles only watermark-covered completions; advances only if none remain | continue | held-response/new-window tests |
| B7 | newer evidence is invalid | preserves trusted reset and all authority memories | stores evidence for fail-closed admission | invalid reset tests |
| B8 | endpoint is new | initializes empty state and trusted reset only when valid | stores generation 1 | initial budget tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sameBudgetProvenance` | enforce exact equal-time provenance | mismatch marks conflict | CodeGraph + AST + conflict tests |
| `trustedBudgetWindow` / `classifyBudgetWindow` | validate official parser-equivalent reset and fixed window relation | invalid/conflict cannot reconcile | CodeGraph + AST + parser/window tests |
| `reconcileCompleted` | delete only commitments covered by cycle watermark | nil or later/in-flight commitments remain | CodeGraph + AST + chronology tests |
| `advanceBudgetGeneration` | clear per-generation state after causal proof | wrap marks exhausted | AST + cap/generation tests |

## State mutations and fallbacks

- Manual evidence may refresh the displayed observation but cannot change generation authority or issuance memory.
- A valid cycle is necessary but not sufficient for advance: every prior commitment must also be absent after watermark reconciliation.
- Safety classes bypass this coordinator state in `TryAcquire` and remain unblocked.

## Safety conclusion

- Safe edit boundary: per-endpoint budget evidence and generation state.
- High-risk impact: yes, because accidental generation advance reopens low-priority capacity.
