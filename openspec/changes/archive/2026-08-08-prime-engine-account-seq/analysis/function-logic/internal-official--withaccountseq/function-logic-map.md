# Function Logic Map: `WithAccountSeq`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `seq` | any integer; positive explicit, zero unresolved, negative invalid | caller option | negative is retained only so scoped use can refuse it |
| client construction | option applied before `New` returns | `New` option loop | no concurrent access during mutation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | returned option is applied | atomic store plus immutable `seq > 0` provenance | no error | positive/negative option tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `accountSeq.Store` | publish the configured routing state for later lock-free reads | construction-time only; no I/O | CodeGraph + AST |

## State mutations and fallbacks

- The option stores the exact integer. It does not normalize a negative value
  into unresolved zero because negative configuration must fail without
  discovery.
- Only a positive value sets explicit provenance. Zero follows lazy discovery;
  negative remains invalid and is never used as a header.

## Safety conclusion

- Safe edit boundary: atomic construction-time store and positive-only immutable
  provenance; no network, journal, gate, or order mutation occurs.
- High-risk impact: yes — this option selects the header scope for all official
  reads and mutations.
