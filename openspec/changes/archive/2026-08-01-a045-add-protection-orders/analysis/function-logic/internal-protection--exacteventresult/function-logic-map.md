# Function Logic Map: `exactEventResult`

- Source: `internal/protection/repository.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| durable saga metadata and private typed event | persisted revision is expected+1; last kind, canonical full-event fingerprint, timestamp and event-specific result fields all match | CAS reload + event-specific constructor | false means conflict, never approximate success |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | persisted last kind or full-event fingerprint differs | fingerprint computation only | false | registration-vs-replace active masquerade test |
| B2 | persisted update time differs | none | false | stale different event |
| B3 | dispatch event kind | none | event-specific comparison | all event method table |
| B4 | registration begin outcome | none | exact state/attempt match | idempotent registration retry |
| B5 | registration active outcome | none | exact state/attempt/broker match | lineage retry |
| B6 | mutation unknown outcome | none | exact state/attempt/reason match | unknown method test |
| B7 | replace begin outcome | derives zero quantity default | exact attempt/pending/previous match | replace method test |
| B8 | zero requested replacement quantity | local quantity default only | continue compare | replace default test |
| B9 | replace active outcome | none | exact state/attempt/broker match | replace result test |
| B10 | trigger outcome | none | exact state/broker match | trigger lineage test |
| B11 | close outcome | none | exact state/broker match | close lineage test |
| B12 | discrepancy outcome | none | exact state/reason match | discrepancy test |
| B13 | unknown kind | none | false | private default invariant |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SHA-256 event fingerprint plus event-kind/result comparisons | admit idempotency only when the row proves this exact event already won | pure; no retry/fallback | CodeGraph + AST |

## State mutations and fallbacks

- No mutation or external calls.

## Safety conclusion

- Safe edit boundary: exact durable-result predicate, not semantic or partial equivalence.
- High-risk impact: yes; controls whether CAS loss is reported as success.
