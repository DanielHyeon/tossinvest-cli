# Function Logic Map: `appendObservations`

- Source: `internal/performance/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| tx | active immediate SQLite transaction | caller | every error leaves caller to rollback |
| observations | validated canonical immutable rows | caller-owned observations | divergent identity fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate supplied observations | transaction-local writes | continue | append batch tests |
| B2 | validation fails | none | error | validation tests |
| B3 | existing-row lookup fails | none | wrapped error | store error contract |
| B4 | identity exists | none | exact replay continues or B5 conflicts | replay tests |
| B5 | existing bytes differ | none | `ErrImmutableConflict` | divergence tests |
| B6 | identity absent INSERT fails | pending tx only | error | crash/all-or-none tests |
| B7 | late-backfill cadence reschedule fails | pending tx only | error; otherwise marker clears if overdue | continuous influx test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| row lookup/scan | obtains persisted canonical bytes by primary key | sql.ErrNoRows selects insert; other error fails | current HEAD + AST |
| INSERT | appends one immutable row | uniqueness race rechecked/serialized by immediate tx | concurrent tests |

## State mutations and fallbacks

- Never updates existing observation data and never synthesizes a new observation identity.
- A genuinely new late backfill may clear only the derived cadence marker so it cannot be hidden for 24 hours; exact replay does not touch the marker.

## Safety conclusion

- Safe edit boundary: internal compare-and-append helper.
- High-risk impact: persistence integrity only.
