# Function Logic Map: `Console.accountPanelFrom`

- Source: `internal/console/overview.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Cached holdings | absent, fresh, stale, or failed snapshot | `holdings.peek(now)` | never refresh here; unreadable is not zero |
| Live journal positions | zero or more, partial only with named journal failure | one `overviewLedger` read | join preserves whichever source answered |
| Market/currency | KR/KRW, US/USD, or unknown other | broker/journal market plus fixed market table | never sum across currencies; other market remains named/unmeasured |
| Joined rows | broker and journal identities joined by market+symbol | `joinPositions` | duplicates remain visible; no silent overwrite |
| Broker-only notice | readable journal row absent for a broker holding | enriched joined rows | dashboard names the missing protection basis once |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | scan enriched rows for broker-only holding and stop at first | set one request-local notice flag | no error path | a057 dashboard broker-only notice test |
| B3-B4 | each known market; broker cache unreadable | append blocked market row | function continues for every market | cold/failed cache overview tests |
| B5-B10 | each joined row; skip other market; classify unknown/managed/unmanaged | aggregate request-local counters/sums | no side effect | market totals and journal unreadable tests |
| B11-B13 | no symbols vs values; unknown management vs measured split | append truthful market row | no error path | empty/managed/unmanaged tests |
| B14-B17 | broker readable; collect other-market pairs; append only when present | append named unmeasured other row | no currency inference | other-market tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `holdings.peek` | zero-call dashboard cache read | mutex-only read, no broker call | CodeGraph callees + rate-budget tests |
| `joinPositions` | one row model for aggregate and detail | deterministic in-memory join | AST + portfolio tests |
| `decoratePositionRows` | produce exactly the same management/exit facts as `/positions` | read failures stay unknown; request-local mutation only | AST + a057 parity tests |
| `brokerReadable` | distinguish empty, never read, and failed | typed reason, no retry | overview tests |
| `Managed`/`Unknown` | management count classification | unknown never becomes unmanaged | operator-console tests |

## State mutations and fallbacks

- Only the returned `accountPanel` and local aggregates change.
- Retains and enriches the already-built rows but never calls `holdings.get` or a broker method.
- Unknown market/currency remains named but unmeasured.

## Safety conclusion

- Safe edit boundary: preserve `peek`, per-market aggregation, and unknown semantics while attaching request-local rows.
- High-risk impact: no; no mutation capability or broker refresh is available here.
