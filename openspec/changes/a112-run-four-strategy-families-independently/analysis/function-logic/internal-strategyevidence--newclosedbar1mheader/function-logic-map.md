# Function Logic Map: `newClosedBar1mHeader`

- Source: `internal/strategyevidence/breakout_bar.go`
- Source SHA-256: `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b` (current worktree; verified with `sha256sum` 2026-08-17)
- Signature: `newClosedBar1mHeader(market marketclock.Market, symbol, sessionID, calendarVersion string, openAt time.Time, revision uint64, observedAt time.Time, currency string) (Header, error)`
- Source range: `134:1`–`202:2`
- AST evidence: `ast.json` generated 2026-08-17 (new function, not in the frozen base 016da624); branches 11, returns 12, defers 0, go statements 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Unexported (P2-7 → unexported in the fix round): called from `NewClosedBar1mEnvelope` (breakout_bar.go:300) and from tests. Builds the deterministic `Header` for one closed 1-minute bar; same inputs → identical header (`TestClosedBar1mHeaderCarriesDeterministicBarIdentity`).
- Identity mapping (Manager decisions 3/4/5/8): `SourceRecordID = <MARKET>:<SYMBOL>:<session_id>:60000:<open_at_ms>`, `EvidenceID = <SourceRecordID>:r<revision>`, `IssuerIdentity = <MARKET>:<SYMBOL>`, `SourceEventAt = open_at`, `SourceAvailableAt = open_at + 60 s`, `ObservedAt = IngestedAt = observedAt`, `MarketEffectiveDate` = session date, `Unit = minor`.
- Refusals are typed `ValidationError`s (`invalid(...)`); no partial header is returned on error (`Header{}`).

## Branches and early returns

Exact AST return nodes: `138, 142, 147, 150, 153, 156, 159, 164, 168, 171, 177, 180` (11 refusals + the header literal at 180).

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 137:2 | `marketAndCode` error (market not KR/US) → refuse | `TestClosedBar1mHeaderRefusesInconsistentInput` /unknown market |
| B2 | if | 141:2 | upper-cased trimmed symbol not canonical (empty, `:`, space/NUL) → refuse | `TestClosedBar1mHeaderRefusesInconsistentInput` /empty symbol, `TestClosedBarRefusesSymbolWithRecordSeparatorOrLowerCase` (`AA:PL` via the helper) |
| B3 | if | 146:2 | `sessionDateFor` error (calendar ≠ market, bad date) → refuse | `TestClosedBar1mHeaderRefusesInconsistentInput` /session calendar does not match market |
| B4 | if | 149:2 | `calendarVersion` not canonical text → refuse | `TestClosedBar1mHeaderRefusesInconsistentInput` /empty calendar version |
| B5 | if | 152:2 | `checkMarketCurrency` error → refuse | `TestClosedBar1mHeaderRefusesInconsistentInput` /currency does not match market |
| B6 | if | 155:2 | `revision == 0` → refuse | `TestClosedBar1mHeaderRefusesInconsistentInput` /revision zero |
| B7 | if | 158:2 | zero `openAt` or `observedAt` → refuse | `TestClosedBar1mHeaderRefusesInconsistentInput` /zero observed clock |
| B8 | if | 163:2 | `openAtMS <= 0` or not on a whole minute → refuse | `TestClosedBar1mHeaderRefusesInconsistentInput` /open_at is not on the minute (the `<= 0` sub-clause, a pre-1970 instant, has no dedicated fixture) |
| B9 | if | 167:2 | `TradingDay` error → refuse | not-applicable: unreachable — market pinned to KR/US by B1 and `time/tzdata` embedded (defensive) |
| B10 | if | 170:2 | market-local trading day of `openAt` ≠ session date → refuse | `TestClosedBar1mHeaderRefusesABarFromAnotherCalendarDay`, `TestClosedBarRequiresTheSessionCalendarDay`; accept side `TestClosedBarAcceptsEveryBarOfTheMarketLocalSessionDay` (KR 09:00 KST) |
| B11 | if | 176:2 | `observedAt` before `openAt + 60 s` → refuse | `TestClosedBar1mHeaderRefusesInconsistentInput` /observed before the bar closed |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `marketAndCode(market)` | 136 | normalises the market and yields the upper-case code used in every identity string |
| `canonicalSymbolText`, `sessionDateFor`, `canonicalPayloadText`, `checkMarketCurrency` | 141, 145, 149, 152 | shared text/identity predicates (same rules the payload gate re-applies) |
| `normalizedMarket.TradingDay(openAtUTC)` | 166 | market-local calendar day (`internal/clock`) |
| `closedBar1mRecordID`, `evidenceIDFor`, `revisionIdentityFor`, `supersededRevisionIdentity` | 179, 181, 190, 191 | identity derivation; pinned by `TestClosedBar1mHeaderCarriesDeterministicBarIdentity` (record id, evidence id, r1/r2, supersedes) |
| `strings.ToUpper/TrimSpace`, `time` arithmetic | 140, 144, 161–175, 197 | normalisation only |

## State mutations and fallbacks

- None. Locals only (AST: 11 `:=` assignments, no package state, no defers/goroutines). No fallback or defaulting: any inconsistent input refuses the whole header. `IngestedAt` is set equal to `observedAt` here and overwritten by the store clock on `Append` (store behaviour, not this function).

## Safety conclusion

- The header helper is the only place the bar identity strings and the dual-cutoff clocks are derived, and it is unexported so producers must go through `NewClosedBar1mEnvelope`. Every refusal branch except the defensive B9 is pinned by a named test, and the calendar-day rule (B10) was re-probed by both reviewers at ET/KST edges (review.md 2026-08-17 rechecks). Residual a1 (a hand-built `Header` through public `NewEnvelope` bypasses this helper) is recorded in review.md and closed on the read side by `scanClosedBarRecord`.
