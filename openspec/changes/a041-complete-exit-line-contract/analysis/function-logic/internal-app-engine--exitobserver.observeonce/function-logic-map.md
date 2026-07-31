# Function Logic Map: `ExitObserver.ObserveOnce`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| cycle observation identity | one immutable instant/sequence per `ObserveOnce` | observer clock + atomic sequence | fail the cycle before judgement |
| quote metadata | positive last price; preserve broker `FetchedAt` | `domain.Quote` | omit unanswered symbol |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | fill detection behind | outage bookkeeping only | deferred cycle | existing outage tests |
| B2 | working-set or price read fails | no judgement/submission | cycle error | existing observation tests |
| B3 | no held state | reset outage latch | empty cycle | existing empty-account test |
| B4 | quote exists | evaluate with one cycle identity | per-position error retained | concurrent snapshot test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `workingSet` | load immutable position/state pair | errors stop observation | CodeGraph + AST |
| `observe` | one price read preserving freshness metadata | retrier owns query freshness | CodeGraph + AST |
| `judge` | evaluate and atomically arm | per-position refusal does not stop other positions | CodeGraph + AST |

## State mutations and fallbacks

- Captures fallback observation identity once; `FetchedAt` replaces it when authoritative.

## Safety conclusion

- Safe edit boundary: thread preserved quote metadata and one fallback identity into judgement.
- High-risk impact: yes — drives protective exit submission.
