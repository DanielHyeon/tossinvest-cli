# Function Logic Map: `BudgetCoordinator.TryAcquire`

- Source: `internal/scheduler/budget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| class | one of seven closed `PollClass` constants | scheduler priority contract | unknown refuses before safety bypass |
| provenance | reported, observed, nonzero reset with exact epoch/delta kind and valid `0 <= remaining <= limit` bounds | official rate headers | missing/unknown/stale/skew/invalid bounds refuse low priority |
| commitments | in-flight plus completed/unreconciled grants under endpoint mutex | coordinator state + authoritative observation lifecycle | completion alone never permits stale capacity reuse |
| capability entropy / issued memory | production `crypto/rand.Reader`, deterministic reader only through package-private test seam; absolute per-endpoint/per-generation issuance cap independent of reported limit | coordinator | entropy failure, generation exhaustion, repeated capability, or cap exhaustion refuses low priority; safety classes remain unblocked |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unknown poll class | none | `UNKNOWN_POLL_CLASS` | unknown-class test |
| B2 | known safety class | none | allowed regardless of provenance | safety continuity test |
| B3 | nil coordinator | none | missing provenance | missing test |
| B4 | equal-time provenance conflict recorded | none | `CONFLICTING_PROVENANCE` | conflicting-provenance test |
| B5 | endpoint/report/reset/time missing | none | missing provenance | missing/reset test |
| B6 | reset kind not exact epoch/delta | none | invalid reset kind | unknown-reset test |
| B7-B9 | clock moved back, reset elapsed, or observation aged 2m | none | skew/stale | stale/skew tests |
| B10 | limit/remaining bounds are invalid | none | `INVALID_BUDGET_BOUNDS` | negative/exceeds-limit/MaxInt test |
| B11 | discretionary capacity is nonpositive or outstanding count reaches it | no subtraction that can underflow | safety reserve | reserve/in-flight/repeated-completion tests |
| B12-B13 | analytics outstanding at half discretionary share | none on denial | entry priority denial or continue | analytics-share test |
| B14 | entropy unavailable, generation exhausted, or absolute issued-cap reached | none | token unavailable | entropy/generation/MaxInt-cap boundary tests |
| B15 | capability read fails | permanently marks entropy unavailable | token unavailable | entropy failure test |
| B16-B17 | endpoint commitment/issued maps absent | initializes maps | continue | defensive branch |
| B18 | capability was already issued in this generation | none; never reissues authority | token unavailable | repeated-capability test |
| success | discretionary capacity remains | stores a cryptographically random capability bound to exact coordinator/key/class/generation | granted with opaque `CommitmentToken` | capability/budget tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `isKnownPollClass` / `isSafetyClass` | closed enum and safety priority | unknown never inherits safety | CodeGraph + AST |
| `SafetyReserve` | max(5, ceil(remaining/2)) | pure | CodeGraph + AST |
| `io.ReadFull` over injected entropy | issue fixed-width unpredictable coordinator/capability bytes | any short read/error fails closed | entropy tests |
| `sha256.Sum256` | bind capability to the exact endpoint key without exposing it | pure | cross-key test |
| mutex lock/unlock | atomic check-and-reserve | defer unlock | CodeGraph + AST |
| `countCommitments` | enforce analytics share over outstanding leases | pure | analytics-share test |

## State mutations and fallbacks

- Safety classes do not create commitments. Low-priority grants reserve atomically; `Complete` only marks consumed/unreconciled, a later authoritative same-window response reconciles completed calls, and a trusted reset clears the old generation.
- No sequential value is exposed or accepted as authority. Capability collision, entropy loss and generation exhaustion all fail closed.
- Reconciliation removes active commitments only. The per-generation issued set remains monotonic until a proven new window, so neither reported `MaxInt` nor repeated reconcile cycles can grow it without bound or reuse a capability.

## Safety conclusion

- Safe edit boundary: candidate/entry/analytics cadence; safety loop cadence remains upstream.
- High-risk impact: yes, because unknown values must not consume or bypass reserve.
