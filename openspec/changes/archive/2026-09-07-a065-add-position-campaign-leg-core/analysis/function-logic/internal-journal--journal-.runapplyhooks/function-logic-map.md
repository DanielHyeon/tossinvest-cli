# Function Logic Map: `(*Journal).runApplyHooks`

Current binding: `Journal.runApplyHooks`.

- Source: `internal/journal/apply_hook.go`
- Pre-edit SHA-256: `8d096983d114cc9e4a1bfc1b30e76195a83de9cf6f2440fd30218027b0cfb7c3`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | live fill transaction | `RecordFill` | hook error rolls entire fill back |
| `fill` | newly applied/corrected/terminal observation | `RecordFill` cumulative snapshot | immutable input to all hooks |
| hooks | zero or one bound Project/Exit set | `SetApplyHooks` | all nil is a no-op |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | Project and Exit nil | none | nil | no-hook fill tests |
| B2/B3 | Project set / Project returns error | Position projection in same tx / rollback | wrapped error | `TestAFailingApplyHookRollsBackTheFill` |
| B4/B5 | Exit set / Exit returns error | exit projection in same tx / rollback | wrapped error | `TestAFailingExitHookRollsBackTheProjectionToo` |
| end | all configured hooks succeed | handle invalidated before return | nil | `TestApplyHooksRunInsideTheFillTransaction` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Project` | Position remains sole fill projection authority | any error aborts fill transaction | CodeGraph + AST |
| `Exit` | resolve exit projection after Position | any error aborts fill transaction | CodeGraph + AST |
| `handle.invalidate` | prevent tx handle escape | always deferred | AST |

## State mutations and fallbacks

- Hook order is Position first, Exit second.
- There is no fallback or partial commit: any configured hook error aborts caller transaction.
- a065 inserts Campaign after Position and before Exit so campaign sees the authoritative Position delta while exit safety still observes the same commit.

## Safety conclusion

- Safe edit boundary: extend the all-nil predicate and add one ordered Campaign call using the same handle/error rollback pattern.
- High-risk impact: yes; predecessor late fill must never be dropped, truncated, or committed apart from Position.

## Post-edit verification (2026-08-03)

- Effective order is `Project → Campaign → Exit`, on one `ApplyTx` and one deferred invalidation.
- Campaign errors are wrapped and returned before Exit; the caller's fill transaction rolls Position,
  campaign watermark, fill event and snapshot back together.
- Delta-zero terminal evidence is not a no-op: it can close a residual leg, while its identical retry is a no-op.
- Successor lookup errors propagate; they are never interpreted as “no successor”.
