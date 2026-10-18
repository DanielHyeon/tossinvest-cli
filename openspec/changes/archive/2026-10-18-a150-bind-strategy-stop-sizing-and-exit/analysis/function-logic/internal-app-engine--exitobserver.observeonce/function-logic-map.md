# Function Logic Map: `ExitObserver.ObserveOnce`

Source `internal/app/engine/exitloop.go`; pre-edit extraction `ast.json` and `ast-grep.txt`; risk scan `risk-pattern-report.md`. Base 28a800f5.

## Inputs and invariants

| Input/state | Source of truth | Failure behavior |
|---|---|---|
| Context, journal and clock | observer options | Current cycle records errors; loop remains active. |
| Held positions and exit states | workingSet | Valid protective work must proceed before alert delivery. |
| Prices and fill-detection pressure | retrier and SLO | Defer or report outage; no fabricated price. |

## Branches and early returns

| Branch | Condition | Side effect / return | Required verification |
|---|---|---|---|
| B1 | fill detection behind | deferred and outage check | existing exitloop SLO tests |
| B2 | working set read fails | cycle error and return | existing failure tests |
| B3 | no managed states | settle unobserved, return | existing unmanaged tests and a150 missing binding alert |
| B4 | price read fails | error and outage check | existing price outage tests |
| B5 | quote missing or unusable | record unobserved, skip position | existing quote tests |
| B6 | quote usable | evaluate and submit in order, retain first error | existing protective-exit tests; a150 latch failure remains managed |
| B7 | after all positions | settle unobserved | a150 alert recorded after protective work; no network lock |

## Calls and live bindings

`workingSet` loads and prepares exits; `observe` batches quotes; `judge` evaluates and submits; `settleUnobserved` records monitoring. Deferred `endExitOrderCycle` follows all returns. Add only deferred local outbox recording after judgement; no broker call or new network lock.

## Safety conclusion

High-risk impact: yes. Preserve every early return and protection order; no alarm publishing before existing valid stops are evaluated.

## State mutations and fallbacks

Current transaction and fallback branches are enumerated in the current AST and branch map; mutation failures propagate, with protective-exit behavior verified separately.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
