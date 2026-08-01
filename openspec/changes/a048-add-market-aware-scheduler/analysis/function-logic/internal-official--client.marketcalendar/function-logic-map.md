# Function Logic Map: `Client.MarketCalendar`

- Source: `internal/official/calendar_reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| country | trimmed/case-folded KR or US | official endpoint contract | other value returns before request |
| date | empty or exact valid Gregorian `YYYY-MM-DD` | official query contract | malformed/impossible value returns before request |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | country not KR/US | none | validation error | existing country tests |
| B2 | nonempty date is not exact/valid | none | validation error | exact-date test |
| B3 | valid nonempty date | query only | adds one `date` value | calendar request tests |
| B4 | official GET fails | HTTP read only | propagate error | client tests |
| success | empty/valid date and successful response | official GET | decoded map | calendar read tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validateMarketCalendarDate` | exact lexical and Gregorian validation | pure, fail before network | AST + exact-date test |
| `Client.get` | official GET through authenticated client | existing retry/error contract | CodeGraph + HTTP fixture |

## State mutations and fallbacks

- No account mutation; invalid input cannot reach token or calendar endpoints.

## Safety conclusion

- Safe edit boundary: official read query validation.
- High-risk impact: yes, because a wrong day could become entry calendar evidence.
