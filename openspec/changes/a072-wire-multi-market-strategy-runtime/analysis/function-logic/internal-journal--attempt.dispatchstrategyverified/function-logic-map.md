# Function Logic Map: `Attempt.DispatchStrategyVerified`

- Source: `internal/journal/strategy_dispatch_runtime.go`
- Qualified function: `Attempt.DispatchStrategyVerified`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| attempt | non-nil journal-owned RECORDED attempt | durable journal plus sealed handle | invalid request or fenced transaction error |
| lease CAS | exact current CLAIMED lease/owner/revision/fencing token bound to this first leg | durable v26 strategy binding and current owner row | fail before dispatch callback |
| dispatch callback | one non-retrying official Gateway callback | caller provides capability, journal controls invocation point | nil refuses before mutation |
| existence verifier | mandatory official byte-exact broker readback | official Gateway verifier | nil refuses before mutation; failure yields IN_DOUBT |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | attempt/journal/dispatch/verifier is nil | none | invalid request | mandatory verifier paired regression |
| B2 | atomic attempt+lease pre-send transaction fails | rollback; broker callback not called | fenced/error | owner/revision/rollback suites |
| B3-B11 | classify dispatch result | see class rows below | authoritative result or IN_DOUBT | paired outcome matrix |
| B4 | provably NOT_SENT | target NOT_DISPATCHED | dispatch-not-sent reason | refusal/release suites |
| B5 | definitive rejection | target FAILED_CONFIRMED | rejected reason | rejection suites |
| B6-B9 | ACK path | raw ID may be persisted ACKED; verifier decides confirmed vs IN_DOUBT | exact result | ACK/round-trip/fill-race suites |
| B10 | durable transport ambiguity | target IN_DOUBT | ambiguous reason | unknown transport suite |
| B11 | unsupported class | target IN_DOUBT | unknown-class reason | caller-class hardening test |
| B12 | composite settlement fails | terminal transaction rolls back; durable pre-send/ACK evidence remains | error | injected late failure suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `beginStrategyDispatchTx` | atomically consume nonce, set core DISPATCH_STARTED and lease SUBMITTING | exact CAS; rollback on any mismatch; zero broker calls | paired pre-send transaction tests |
| `dispatch` | perform the one official mutation attempt | exactly once and only after committed SUBMITTING | duplicate/restart tests |
| `context.WithoutCancel` + bounded timeout | finish durable settlement after broker response despite caller cancellation | 15-second finite window; no resend | cancellation tests |
| `strategyTransitionOnly` | store raw broker ID byte-exact in ACKED before readback | failure returns with no invented confirmation | ACK identity tests |
| `verify` | authoritative official order existence readback | mandatory; failure never retries submit | mandatory verifier/round-trip tests |
| `settleStrategyDispatchTx` | commit core outcome, lease disposition, risk/campaign/strategy links and ACK-window backfill/release together | one immediate transaction; rollback on any failure | composite settlement and injected rollback suites |

## State mutations and fallbacks

- Pre-send mutation is one atomic `RECORDED→DISPATCH_STARTED` plus `CLAIMED→SUBMITTING` transition.
- Nonblank ACK is recorded byte-exact before official readback.
- Terminal settlement owns core attempt, lease disposition, aggregate/five buckets, real order mapping,
  campaign watermark, strategy lineage and missed-fill repair in one transaction.
- There is no automatic resubmit fallback. Unknown or unverified transport remains `IN_DOUBT` plus
  `AMBIGUOUS+HELD` for separately attested reconciliation.

## Safety conclusion

- Safe edit boundary: keep ordinary `DispatchVerified` unchanged and keep this method strategy-only.
- High-risk impact: yes — exact-once real-money dispatch and crash-recovery authority boundary.
