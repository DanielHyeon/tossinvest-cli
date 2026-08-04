# Pre-edit Function Logic and Branch Test Map — engine market-worker supervisor

## Scope and existing-function disposition

This checkpoint adds a new `internal/app/engine` strategy-entry supervisor and focused tests. It does not edit
`Runtime.Run`, `NewRuntime`, command assembly, journal schema, Guardian, Gateway, activation, or LIVE mutation
paths. Consequently there is no existing Go function body to map before this checkpoint. If implementation
requires an existing body edit, its generated AST and Function Logic Map must be added before that edit.

The existing runtime contract remains the outer safety boundary: any supplied `SupervisedLoop` that returns
abnormally causes `Runtime.Run` to cancel and drain every loop. The new strategy-entry loop therefore absorbs
market-local failures internally and returns only for parent cancellation or an explicitly typed central
integrity failure.

## Function Logic Map — new `StrategyEntrySupervisor`

1. Construction requires exactly one KR and one US descriptor, independent bounded queues, and no combined
   market descriptor. Both descriptors default to effective OFF and may omit an evaluator while OFF.
2. `SupervisedLoop` publishes one outer loop named `strategy-entry`; the existing engine runtime sees no KR/US
   child goroutine and therefore cannot misclassify a market-local failure as a safety-loop failure.
3. `Run` creates independent child contexts/goroutines and releases one shared start barrier only after both
   are installed. A work item from one bounded queue is never read by the peer market.
4. OFF or latched workers reject new triggers without invoking evaluation. A full market queue refuses the
   trigger without blocking the caller or consuming peer/safety capacity. Shutdown flips accepting OFF under
   the same lock as enqueue, cancels children, and drains both bounded queues.
5. A successful cycle leaves the market active. A cycle error, panic or watchdog deadline records an immutable
   market/revision/generation/freshness/evidence-bound fault and latches only that market effective OFF.
6. This leaf has no durable latch or recovery callback and no ON transition after failure. The future central
   owner must persist the fault and consume a separately specified exact durable release receipt; boolean,
   stale, wrong-market or replay recovery cannot be expressed by this API.
7. Evaluation callbacks receive no mutation dependency and run behind a bounded watchdog. A callback that
   ignores cancellation may finish later, but its buffered result has no state transition or writer action.
   Sequential child execution plus the irreversible OFF latch limits abandoned callbacks to one per market;
   no cycle or restart is accepted after abandonment.
8. A peer market continues while the failed market is latched. The outer loop also remains alive, so
   reconciliation, fill detection, protection, exit, and emergency reduction loops are not cancelled.
9. An explicitly typed central integrity failure, saturated fault handoff, or impossible supervisor invariant
   returns from the outer loop. Existing `Runtime.Run` then performs process-wide fail-closed cancellation.
10. Parent cancellation cancels and drains both supervised children and returns the parent context error.
    Subsequent triggers are disabled and queue depth is zero. A context-ignoring evaluation-only callback can
    outlive the child but receives no writer and its late result is discarded.

## Branch Test Map

| Branch | Expected evidence |
|---|---|
| Default dormant KR and US | Both snapshots OFF/unlatched; triggers disabled; zero evaluator calls |
| Concurrent start | Barrier proves KR and US evaluators can enter before either is released |
| KR cycle error | KR OFF/latched with first fault; US work completes; outer loop remains alive |
| US panic | US OFF/latched; panic is contained; KR and safety loop continue |
| No recovery surface | Failed market stays OFF; API exposes no boolean or in-memory ON transition |
| Queue saturation | Submit is non-blocking and returns FULL only for the saturated market |
| Shutdown/enqueue barrier | Racing triggers become DISABLED after acceptance closes; both queues finish depth 0 |
| Ignored cancellation | Run drains without waiting for the evaluation-only callback; late result causes no action |
| Watchdog deadline | Only the timed-out market latches and its late result cannot emit another fault or central action |
| Central integrity error | Outer `strategy-entry` loop returns typed central error; existing runtime drains safety loops |
| Parent cancellation | Both child contexts drain; outer loop returns cancellation and runtime is graceful |
| Invalid assembly | Duplicate/missing/combined market, active nil evaluator/incomplete authority, dormant authority, or invalid bounds refused |

Static dependency checks must continue to show that this file imports no journal, broker, execgw, WTS,
activation writer, operating-setting writer, or order transport package. This checkpoint supplies supervision
only; production command wiring remains dormant until the durable a066/a071/a072 authority adapters land.
