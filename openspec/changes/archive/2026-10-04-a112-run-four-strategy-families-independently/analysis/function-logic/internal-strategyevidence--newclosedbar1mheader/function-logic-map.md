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

| Callee expression | Position |
|---|---|
| `marketAndCode` | 136:33 |
| `strings.ToUpper` | 140:17 |
| `strings.TrimSpace` | 140:33 |
| `canonicalSymbolText` | 141:6 |
| `invalid` | 142:20 |
| `strings.TrimSpace` | 144:20 |
| `sessionDateFor` | 145:22 |
| `invalid` | 147:20 |
| `err.Error` | 147:67 |
| `canonicalPayloadText` | 149:6 |
| `invalid` | 150:20 |
| `checkMarketCurrency` | 152:12 |
| `invalid` | 156:20 |
| `openAt.IsZero` | 158:5 |
| `observedAt.IsZero` | 158:24 |
| `invalid` | 159:20 |
| `openAt.UTC` | 161:15 |
| `openAtUTC.UnixMilli` | 162:14 |
| `invalid` | 164:20 |
| `normalizedMarket.TradingDay` | 166:21 |
| `invalid` | 168:20 |
| `err.Error` | 168:64 |
| `invalid` | 171:20 |
| `openAtUTC.Add` | 174:14 |
| `observedAt.UTC` | 175:17 |
| `observedUTC.Before` | 176:5 |
| `invalid` | 177:20 |
| `closedBar1mRecordID` | 179:14 |
| `uint64` | 179:92 |
| `evidenceIDFor` | 181:31 |
| `revisionIdentityFor` | 190:31 |
| `supersededRevisionIdentity` | 191:31 |
| `strings.ToUpper` | 197:31 |
| `strings.TrimSpace` | 197:47 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
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
