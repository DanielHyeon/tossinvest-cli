# Function Logic Map: `applyReservationTransition`

- Source: `internal/weeklyvaluelane/reservation.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| state | maps initialized; scope version/count/ordinals keyed by campaign+market | durable replay state | conflict |
| command | bounded identities, trusted evaluatedAt, exact campaign/market scope | scheduler port | typed conflict/refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | duplicate idempotency key+same fingerprint | none | duplicate prior result | retry test |
| B2 | scoped expected version mismatch | none | VERSION_CONFLICT | KR/US isolation test |
| B3 | reserve invalid/stale/non-next ordinal/existing week | none | conflict | sequence/calendar tests |
| B4 | positive fill missing/terminal/wrong identity or ordinal | none | typed refusal | distinct ordinal tests |
| B5 | positive fill next distinct ordinal | clone scoped state; consume ordinal/count | applied | seven-leg test |
| B6 | authoritative zero release | clone scoped state; release only | applied | zero-fill test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ValidateMarketWeek | trusted-time calendar check | fail closed | CodeGraph + AST |
| CanonicalReservationKey | campaign+market+stable week key | deterministic | CodeGraph + AST |
| cloneReservationState | copy-on-write pure transition | preserve original | CodeGraph + AST |

## State mutations and fallbacks

- State is cloned before mutation. Versions, counts, consumed ordinals and receipts are scope-local.

## Safety conclusion

- Safe edit boundary: pure CAS transition; persistence remains caller-owned.
- High-risk impact: yes; controls weekly admission and seven-leg consumption.
