# Function Logic Map: `adaptCalendarDay`

- Source: `internal/scheduler/calendar.go`
- Source SHA-256: `004b8011b16021f4922a46c54b6a0e3278aaca9876018269dd548fd47ffa33f4` (current worktree; verified with `sha256sum` 2026-08-17, equal to `source_sha256` in `ast.json`)
- Signature: `adaptCalendarDay(market marketclock.Market, loc *time.Location, day official.MarketCalendarDay) (CalendarDay, error)` (`ast.json`: `adaptCalendarDay(params=3, results=2)`)
- Source range: `103:1`–`134:2`
- AST counts: branches 10, returns 7, calls 13, defers 0, go statements 0 (`ast.json` generated 2026-08-17 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

- The per-day adapter behind `AdaptOfficialCalendar` (calendar.go:64, 68, 72), which calls it three times — previous, today, next. It is the only place where an official calendar payload becomes a scheduling fact, which is why the L1b brief cites it: the session window a112 uses to bound a US bar day is produced here.
- `loc` is the exchange's IANA location, already resolved by the caller through `market.Location()` (calendar.go:60–63). That call is what rejects any market outside KR and US, so this function's own `default` arm cannot be reached through the only caller.
- The two markets read different shapes of the same payload: KR sessions arrive under `day.Integrated.RegularMarket` (113–115) and US sessions under `day.RegularMarket` (116–117). A nil `Integrated` for KR leaves `session` nil, which is the holiday encoding, not an error.
- An empty `day.Date` is not a refusal here: it returns a zero `CalendarDay` with a nil error (104–106). The refusal is the caller's — `AdaptOfficialCalendar` then fails to parse that empty date in the chronology check (calendar.go:76–81).
- A present session must be well-formed and exchange-local: both endpoints non-zero and ordered (125), and both must fall on `day.Date` when rendered in `loc` (128). `EarlyClose` is derived, not reported: `EndTime.Sub(StartTime) < regularSessionDuration` where `regularSessionDuration` is 6h30m (calendar.go:21). A full KR 09:00–15:30 and a full US 09:30–16:00 session are both exactly 6h30m, so the boundary is not an early close.

## Branches and early returns

Exact AST return nodes: `105, 108, 119, 123, 126, 129, 133`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 104:2 | `day.Date == ""` → zero `CalendarDay`, nil error (the caller decides) | `TestCalendarRejectsMissingOrNonChronologicalBusinessDayEvidence` (mutation setting `Today.Date` to the empty string; the snapshot is then refused by the chronology check) |
| B2 | if | 107:2 | `day.Date` is not `2006-01-02` → `invalid date %q` | not-applicable: no existing test supplies a non-empty, unparsable date — every payload helper builds well-formed dates or the empty string |
| B3 | switch | 111:2 | dispatch on market to find the regular session | `TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose` (KR), `TestUSCalendarUsesExchangeTimezoneAcrossDST` (US) |
| B4 | case | 112:2 | KR arm | `TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose`, `TestCalendarVersionIsCanonicalAndSemantic` |
| B5 | if | 113:3 | `day.Integrated != nil` → take `day.Integrated.RegularMarket` | taken: `TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose` (regular and early-close cases); untaken (nil `Integrated`, session stays nil): the same test's holiday case and `TestCalendarRejectsMissingOrNonChronologicalBusinessDayEvidence` |
| B6 | case | 116:2 | US arm → `day.RegularMarket` (US payloads carry no `Integrated`) | `TestUSCalendarUsesExchangeTimezoneAcrossDST` (EST and EDT), `TestCalendarRejectsMalformedExchangeDay` |
| B7 | case | 118:2 | any other market → `unsupported market %q` | not-applicable: unreachable through the only caller — `AdaptOfficialCalendar` resolves `market.Location()` first and that refuses anything outside KR/US with `ErrUnknownMarket` (market.go:79–92, `TestUnknownMarketFailsClosed`) |
| B8 | if | 122:2 | `session == nil` → return the day with `Regular` nil and no error (a holiday) | `TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose` (holiday case asserts `Today.Regular == nil` with a nil error) |
| B9 | if | 125:2 | zero start, zero end, or start not before end → `invalid regular session` | not-applicable: no existing test supplies a zero or inverted session window |
| B10 | if | 128:2 | either endpoint's exchange-local date differs from `day.Date` → `regular session is outside exchange-local date %s` | `TestCalendarRejectsMalformedExchangeDay` (date 2026-03-09 with a session on 2026-03-10) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `time.Parse("2006-01-02", day.Date)` | 107 | date shape check; the parsed value is discarded, only the error is used |
| `fmt.Errorf("invalid date %q: %w", …)` | 108 | date refusal |
| `fmt.Errorf("unsupported market %q", market)` | 119 | market refusal (unreachable through `AdaptOfficialCalendar`) |
| `session.StartTime.IsZero()`, `session.EndTime.IsZero()`, `session.StartTime.Before(session.EndTime)` | 125 | well-formedness of the window |
| `fmt.Errorf("invalid regular session")` | 126 | window refusal |
| `session.StartTime.In(loc).Format("2006-01-02")`, `session.EndTime.In(loc).Format("2006-01-02")` | 128 | exchange-local day binding; this is what makes the DST cases in `TestUSCalendarUsesExchangeTimezoneAcrossDST` exact |
| `fmt.Errorf("regular session is outside exchange-local date %s", day.Date)` | 129 | day-binding refusal |
| `session.EndTime.Sub(session.StartTime)` | 132 | early-close derivation against `regularSessionDuration` (6h30m) |

## State mutations and fallbacks

- Pure. No receiver, no package-level state, no I/O. The locals are `session` (a pointer into the caller's payload, never written through) and `out` (6 AST assignments, all local).
- The returned `SessionWindow` holds the payload's absolute instants unchanged (131) — the conversion to `loc` at 128 is used only for the day comparison, so no timezone rewriting reaches the snapshot. `AdaptOfficialCalendar` then hashes the adapted days into a fetch-time-independent version (`TestCalendarVersionIsCanonicalAndSemantic`).
- No fallback: a malformed day returns `CalendarDay{}` with an error, never a partially filled day. The one non-error empty return is B1, and it is empty on purpose so the caller's chronology check refuses it.

## Safety conclusion

- Safe edit boundary: B8 and B9 must stay distinct. Collapsing "no session" (a holiday, a legitimate nil) into "invalid session" (a refusal), or the reverse, would either block a normal trading day or let a garbage window through as a schedule.
- High-risk impact: no direct order, sizing or protection effect — the scheduler is a gating input, and `AdaptOfficialCalendar` fails closed above this function whenever a day is missing, out of order, or lacks session evidence (calendar.go:79–84). It is nevertheless the source of the session boundaries a112 would bind a US bar day to, so a widened B10 would silently move that boundary.
- Untested branches are three: B2 (unparsable date), B7 (unreachable market default) and B9 (zero or inverted window). Every arm that produces a schedule — B1, B3, B4, B5, B6, B8, B10 — is covered. Package suite green (`go test ./internal/scheduler -count=1`, 2026-08-17).
