# Function Logic Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context, mutation plan, Guardian reference | canonical mutation and durable decision reference | caller supplies shape only; journal row is authority | invalid request or typed refusal; broker 0 calls |
| symbol in-flight state | one mutation per market/symbol | in-process claim plus journal pending/unresolved reads | non-journalled symbol-in-flight refusal |
| decision and reservation | current, unspent, unexpired, exact preimage/class/client key and HELD q_final authority | journal re-read | journalled `NOT_DISPATCHED` refusal |
| protection/entry/preflight | current and exposure-compatible | sealed readiness, entry gate and preflight | journalled `NOT_DISPATCHED` refusal |
| strategy dispatch input | exact CLAIMED lease CAS plus opaque frozen FX evidence | journal current owner/lease/binding and officialfx seal | exact pre-transport refusal consumes the claim and releases all six holds; broker 0 calls |
| strategy terminal result | durable core attempt state, raw broker id and first-leg binding | journal only; never a caller-supplied class/digest | core attempt, lease, risk order, campaign and strategy lineage settle in one transaction or all remain unresolved |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | prepare, in-process claim, pending-symbol and spent-decision errors/refusals | no broker call; attempt may not yet exist | invalid request / typed refusal | existing gateway replay/symbol/decision suites |
| B9-B15 | durable Prepare and protection/entry/preflight/decision/reservation checks | RECORDED attempt, then possible NOT_DISPATCHED settlement | typed refusal | existing failclosed/protection/reservation suites |
| B16-B20 | final callback re-read, decision/key/protection/q_final authority drift | dispatch transaction exists but send tracker and broker do not | `DispatchNotSent` | existing last-moment race suites |
| added S1 | q_final/first-leg exposure-raising request lacks exact strategy lease authority | NOT_DISPATCHED; a claim is released only after exact owner/revision/full binding proof | typed missing-lease refusal | paired KR/US `TestStrategyGatewayRequiresClaimedLeaseBeforeBroker` |
| added S2 | opaque FX is missing/stale/forged/wrong pair or differs from decision envelope | no `SUBMITTING`, no send tracker, no broker | typed FX-binding refusal | paired KR/US account-base Gateway matrix |
| added S3 | any protection/entry/preflight/decision/reservation/FX refusal while the exact lease is still CLAIMED | attempt `NOT_DISPATCHED` plus atomic lease `REFUSED+RELEASED` and aggregate+five-bucket release | original typed refusal; release failure is surfaced, never hidden | paired KR/US pre-transport matrix |
| added S4 | exact current FX and CLAIMED CAS | strategy-specific pre-send transaction spends nonce, marks dispatch started and commits `CLAIMED->SUBMITTING` before the send tracker exists | existing mutation classification | paired KR/US exact-current success |
| added S5 | ACK/roundtrip yields a terminal classification | bounded caller-cancellation-independent settlement derives the class from the durable attempt; one transaction closes core+lease+risk+campaign+lineage and backfills any ACK-window fill | terminal outcome or fail-closed unresolved state | paired KR/US cancellation/fill-race matrix |
| added S6 | caller supplies or tampers with an outcome class, digest or broker id | no such production input exists; raw broker id is read and compared byte-exact from the durable attempt | refuse/recovery-only path | paired KR/US authority-substitution tests |
| B21-B25 | ordinary dispatch/settlement error, nonce reuse, confirmed/refused result branches | ordinary Place/Cancel/Amend retain the existing durable attempt settlement | outcome/error | existing dispatch and roundtrip suites |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loadDecision`, `prepareRequest` | derive durable attempt binding from journal decision | fail closed; no caller risk authority | CodeGraph + AST |
| `claimSymbol`, `checkSymbolFree`, `checkDecisionUnspent` | preserve one mutation/decision use | no retry or fallback | current source/tests |
| `Journal.Prepare` | commit RECORDED before possible send | synchronous durable journal write | current source/tests |
| `checkProtection`, `checkEntry`, preflight | exposure-raising safety gates | exits bypass entry-only gates | current source/tests |
| `checkDecision`, `checkReservation` | repeat exact current decision and q_final holds | read failure is refusal | current source/tests |
| account-base FX validator | validate opaque evidence at Gateway clock against persisted decision envelope | no scalar fallback | paired KR/US account-base tests |
| `Journal.RefuseClaimedStrategyDispatchPreTransport` | consume one exact provably-not-sent claim and release aggregate plus five buckets | exact current owner/revision/full plan/first-leg proof; one transaction | paired KR/US refusal tests |
| strategy-specific dispatch method / `plan.call` | atomically bind dispatch start to SUBMITTING, then perform exactly one official mutation | ordinary dispatch API unchanged; broker errors classified and never blindly retried | paired strategy dispatch plus ordinary regression tests |
| journal composite terminalizer | derive result from durable attempt and transfer the exact first-leg lifecycle atomically | bounded settlement context; crash recovery requires sealed official attestation and never resends | paired KR/US cancellation, rollback and fill-race tests |

## State mutations and fallbacks

- Mutates in-process symbol ownership and durable intent/attempt state before dispatch.
- The strategy branch adds only the exact journal `SUBMITTING` transition immediately before send
  tracking and an atomic terminal lifecycle transfer after observation. It may not add a broker,
  paper, shadow or caller-supplied outcome-authority path.
- Before SUBMITTING, every exact claimed strategy refusal consumes the lease and releases all six
  holds atomically. After SUBMITTING, no pre-transport release API is reachable.
- After the broker call, settlement uses a short bounded context detached from caller cancellation;
  failure leaves durable SUBMITTING/ACKED/IN_DOUBT evidence for sealed no-resend recovery.
- Reductions/cancels and non-raising amends retain their current branch order and never require FX.

## Safety conclusion

- Safe edit boundary: private strategy capability and a strategy-only dispatch branch on the existing
  `mutationPlan`; ordinary mutations preserve current behavior, while q_final first-leg entries fail
  closed without exact lifecycle authority.
- High-risk impact: **yes** — last pre-broker fence and real-money mutation path.
