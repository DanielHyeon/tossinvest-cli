# Branch Test Map: `adaptCalendarDay`

- Source: `internal/scheduler/calendar.go`, SHA-256 `004b8011b16021f4922a46c54b6a0e3278aaca9876018269dd548fd47ffa33f4`; branch IDs follow `ast.json` (generated 2026-08-17).
- AST counts: branches 10, returns 7, calls 13, defers 0, go statements 0. Source range `103:1`–`134:2`. Signature `adaptCalendarDay(market marketclock.Market, loc *time.Location, day official.MarketCalendarDay) (CalendarDay, error)`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 104:2 — an empty `date` yields a zero day and a nil error, and the caller refuses the snapshot on it | `TestCalendarRejectsMissingOrNonChronologicalBusinessDayEvidence` | n/a (not edited) | existing suite green |
| B2 | if at 107:2 — a non-empty date that is not `2006-01-02` | not-applicable: no existing test supplies a malformed date string | n/a (not edited) | not-applicable |
| B3 | switch at 111:2 — market dispatch for the regular session | `TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose`, `TestUSCalendarUsesExchangeTimezoneAcrossDST` | n/a (not edited) | existing suite green |
| B4 | case at 112:2 — KR reads the session from `Integrated` | `TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose`, `TestCalendarVersionIsCanonicalAndSemantic` | n/a (not edited) | existing suite green |
| B5 | if at 113:3 — `Integrated` present gives the session; `Integrated` nil leaves it nil (holiday) | `TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose` (both sides), `TestCalendarRejectsMissingOrNonChronologicalBusinessDayEvidence` (nil on previous/next) | n/a (not edited) | existing suite green |
| B6 | case at 116:2 — US reads `RegularMarket` directly, across an EST and an EDT day | `TestUSCalendarUsesExchangeTimezoneAcrossDST`, `TestCalendarRejectsMalformedExchangeDay` | n/a (not edited) | existing suite green |
| B7 | case at 118:2 — a market that is neither KR nor US | not-applicable: unreachable through `AdaptOfficialCalendar`, which resolves `market.Location()` first; the unknown-market refusal is pinned upstream by `TestUnknownMarketFailsClosed` | n/a (not edited) | not-applicable |
| B8 | if at 122:2 — no regular session for the day: the day is returned with `Regular` nil and no error | `TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose` (holiday case) | n/a (not edited) | existing suite green |
| B9 | if at 125:2 — a zero or inverted session window | not-applicable: no existing test builds one; the payload helpers always produce ordered non-zero instants | n/a (not edited) | not-applicable |
| B10 | if at 128:2 — a session whose exchange-local day disagrees with the day's `date` | `TestCalendarRejectsMalformedExchangeDay` | n/a (not edited) | existing suite green |

Derived value below the last branch (line 132, not a branch): `EarlyClose` is `EndTime.Sub(StartTime) < regularSessionDuration` (6h30m). Covered by `TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose` (09:00–15:30 KST is not an early close; 09:00–12:30 KST is) and by `TestCalendarVersionIsCanonicalAndSemantic`, which shows that moving the close changes the canonical calendar digest.

Verification: `go test ./internal/scheduler -count=1` green on 2026-08-17 (exit 0). No RED round applies — a112 does not edit this function.
