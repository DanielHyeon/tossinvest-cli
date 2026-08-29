# Function Logic Map: `ValidateMarketWeek`

- Source: `internal/weeklyvaluelane/reservation.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| market/provider/timezone | exact official allowlist | canonical calendar port | OFFICIAL_CALENDAR_INVALID |
| session/week identity | local Monday, derived `market-exchange-YYYY-Www` exact equality | IANA calendar | OFFICIAL_CALENDAR_INVALID |
| evaluatedAt | observed <= evaluated <= freshUntil | trusted command/evidence cutoff | OFFICIAL_CALENDAR_INVALID |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | provider/zone/official/identity bounds invalid | none | refusal | bounds test |
| B2 | timestamp/generation/digest invalid | none | refusal | stale test |
| B3 | timezone/session parse or non-Monday | none | refusal | holiday/DST test |
| B4 | derived stable identity mismatch | none | refusal | forged identity test |
| B5 | valid official week | none | success | KR/US valid fixtures |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| time.LoadLocation/ParseInLocation/ISOWeek | exact IANA week derivation | no local fallback | CodeGraph + AST |

## State mutations and fallbacks

- Pure validation only; server timezone is never read.

## Safety conclusion

- Safe edit boundary: exact canonical identity derivation.
- High-risk impact: yes; weekly exposure uniqueness.
