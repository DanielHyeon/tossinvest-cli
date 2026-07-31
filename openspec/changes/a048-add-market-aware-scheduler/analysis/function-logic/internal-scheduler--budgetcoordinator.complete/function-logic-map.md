# Function Logic Map: `BudgetCoordinator.Complete`

- Source: `internal/scheduler/budget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| key | exact endpoint budget key used at issuance | caller plus token key digest | mismatch has no mutation |
| token | opaque capability returned by this coordinator for one low-priority grant | `TryAcquire` | zero, forged, copied across scope/generation or replayed token returns false |
| commitment record | exact capability, class, generation, completion state, diagnostic completion instant and monotonic completion sequence | endpoint state under coordinator mutex | missing/already completed/reconciled record, unavailable clock, or sequence exhaustion returns false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil coordinator or zero token | none | false | zero/nil capability tests |
| B2 | endpoint key absent | none | false | cross-key test |
| B3 | coordinator identity, key digest or active generation differs | none | false | forge/cross-coordinator/key/generation tests |
| B4 | capability absent, replayed, class changed or record generation differs | none | false | forge/class/repeated completion tests |
| B5 | completion clock unavailable | none | false | clock failure test |
| B6 | completion clock returns zero | none | false | clock failure test |
| success | every binding matches an in-flight record, clock returns nonzero and completion sequence can advance | marks completed-at plus completion sequence but retains the record in commitment count | true | lifecycle/outcome/unlimited-reuse/race tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mutex lock/unlock | make token validation and lifecycle transition atomic with acquire/observe | always unlocked by defer | CodeGraph + AST |
| `sha256.Sum256` | recompute exact endpoint-key binding | pure; mismatch refuses | cross-key test |
| injected completion clock | diagnostic completion time and fail-closed clock seam only | nil/zero fails closed; production uses `time.Now`; reconciliation does not trust wall ordering | clock and wall-rollback tests |

## State mutations and fallbacks

- Completion never deletes a record and never increments available capacity.
- Success, request error and cancellation use the same conservative state transition.
- An opaque observation cycle begun after completion owns reconciliation; reset generation change requires causal coverage of every old commitment and makes every old token unusable.

## Safety conclusion

- Safe edit boundary: low-priority budget commitment lifecycle only; safety-class grants never carry a token.
- High-risk impact: yes, because a forgeable or prematurely released commitment could consume the safety reserve.
