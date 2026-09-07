# Function Logic Map: `(*Journal).SetApplyHooks`

Current binding: `Journal.SetApplyHooks`.

- Source: `internal/journal/apply_hook.go`
- Pre-edit SHA-256: `8d096983d114cc9e4a1bfc1b30e76195a83de9cf6f2440fd30218027b0cfb7c3`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `hooks` | at least one domain apply function is non-nil | apply-hook atomicity contract | `ErrInvalidRequest`, no binding |
| `j.applyHooks` | unbound before this one-time call | `Journal.applyMu` protected state | second binding refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | Project and Exit are both nil | none | `ErrInvalidRequest` | `TestApplyHooksAreBoundOnce` plus Campaign-only extension test |
| B2 | either existing Project or Exit is non-nil | none | `ErrInvalidRequest` | `TestApplyHooksAreBoundOnce` |
| end | valid first binding | assigns `j.applyHooks` under mutex | nil | apply-hook atomic tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `applyMu.Lock/Unlock` | serialize wiring against fill reads | no retry; one critical section | CodeGraph + AST |
| `fmt.Errorf` | wrap invalid wiring as typed invalid request | caller must fail wiring | AST |

## State mutations and fallbacks

- The only mutation is the one-time in-memory hook binding.
- No broker, journal row, position, exit, toggle, or live setting is touched.
- a065 may extend the nil/rebind predicates with Campaign, but must preserve one-time binding.

## Safety conclusion

- Safe edit boundary: add Campaign to the existing all-nil and already-bound checks only.
- High-risk impact: yes; incorrect binding could split fill/campaign atomicity.

## Post-edit verification (2026-08-03)

- The all-nil and rebind predicates now include `Campaign`.
- No journal row, broker call, toggle, or activation was added to wiring.
- `TestCampaignHookRunsBetweenProjectionAndExitAndRollsBackAtomically` and the full journal suite pass.
