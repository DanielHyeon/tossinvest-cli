# Function Logic Map: `BuildExitLine`

- Source: `internal/operatorview/exit_line.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a067-screens-show-what-they-already-know/base-commit.txt`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `source.Snapshot` | nil, or one already-selected persisted snapshot | the caller | nil -> status `unknown`, every actionable value an em dash |
| `source.UnknownReason` | a typed reason string | journal or caller | unrecognised -> the generic "cannot be shown safely" sentence |
| `source.StaleReason` | a typed reason string | the **caller's** freshness policy | non-empty -> status `stale`, values hidden, provenance kept |
| `source.RemainingQuantity` | the ledger quantity | caller | only used for the one-share wording |
| `source.ObservationSource`, `ObservedAt`, `EffectiveSource` | provenance | caller | empty -> em dash |

**Invariant**: this adapter recomputes no price, rung, action, quantity or
identity. It maps a snapshot or it hides it. That is unchanged by a067.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `source.Snapshot == nil` | status `unknown`, `Reason = reasonText(UnknownReason)` | **early return**, all values em dash | existing unknown tests |
| B2 | `StaleReason` non-empty | status `stale`, **`StatusText = staleStatusText(reason)`** (a067), `Reason = reasonText(reason)` | **early return**, all values em dash | a067 2.3, 2.4; existing stale test |
| B3 | one share, state-only, projected zero | the one-share wording | falls through to return | existing one-share test |

**a067 changes B2 only**, and only its `StatusText`. Before: the literal
`오래된 평가` for every stale reason. After: the same literal by default, with
`판정 격리` and `엔진 정지` for the two reasons a067 introduces. The status code
stays `stale`, so no template branch and no JSON consumer has to learn a new
state. The set of values hidden is identical.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `emptyView` | every actionable field starts as an em dash | pure | AST |
| `staleStatusText` (a067) | the short verdict for a hidden line | pure, total; unknown reasons fall to the previous literal | `internal/operatorview/exit_line_status.go` |
| `reasonText` | the sentence under the verdict | pure, total | AST |
| `policyText`, `generationText`, `stageText`, `projectedText`, `actionText`, `dashIfEmpty` | provenance and value rendering | pure | AST |

No clock, no I/O, no package state. The package doc says this file does not read
a clock and a067 keeps that true: the freshness decision is the caller's.

## State mutations and fallbacks

- Builds and returns one value. Mutates nothing.
- The fallback of `staleStatusText` is the exact string the function used to
  return unconditionally, so an unrecognised stale reason renders as it did
  before a067.

## Safety conclusion

- Safe edit boundary: one status string. No value that was hidden is now shown,
  and no value that was shown is now hidden.
- High-risk impact: no. Presentation only.
- Safety invariant 0.3 untouched.
