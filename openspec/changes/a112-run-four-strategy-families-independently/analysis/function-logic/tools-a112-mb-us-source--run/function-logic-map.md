# Function Logic Map: `run`

- Source: `tools/a112-mb-us-source/main.go` (lines 71–223, `source_sha256` 1945343d39b89e45ccc0fbe0accca9a3e3d5a157cab9e5a2ef838d3e8caecac9 — post-amendment; the run-3 executable's source was `2aa28733…`, lines 71–211)
- AST evidence: `ast.json` (38 branches, 35 returns, 5 defers)
- Risk scan: `risk-pattern-report.md` (no configured pattern matched)
- Revision: post-amendment (lot 0.7b.3 GREEN). This function is not in the frozen base (`base-commit.txt` = 016da624 has no `tools/a112-mb-us-source`), so the gate does not classify it as a modified existing function; the map exists because the 2026-08-16 review section and the approved amendment cite its branches as evidence.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cfg` (tokenCache, goBinary, receiptRoot absolute; sessionDate `YYYY-MM-DD`; before bytes, may be empty) | `validateConfig` (main.go:257) | CLI flags via `cli.go` | B1 → HOLD "invalid collector input", 0 requests |
| `deps` (reader, now, identity, builder, optional openReceipt) | all non-nil | `newProductionDependencies` / test doubles | B2 → HOLD "collector dependencies are incomplete" |
| clock `deps.now()` | non-zero | injected clock | B3 → HOLD "clock returned zero time" |
| overall budget | 120 s from `started`; `collectorLive` also requires ≥15 s left before each request | `overallBudget`, `collectorLive`, `boundedRequestContext` | every `live()` failure → HOLD; deferred taint |
| receipt root | existing current-UID 0700 dir under `/tmp` | `openReceipt` (receipt_unix.go:35) | B6 → HOLD "receipt root" |
| pre-request identity | source/compiled/untracked/Go/Git/base/executable hashes | `snapshotIdentity` + `verifyPrescribedBinary` | B10/B11 → HOLD before any request |
| candle page cap | exactly 4 (`for page < 4`, B16) | main.go:135 | after 4 pages without raw-null terminal → `candle-crawl.json` records `cap_exhausted` (B26 @175) and the run continues; no 5th request |
| cursor | raw JSON `null` (terminal) or non-empty UTF-8 string equal to its decoded value, not previously seen | `nextCursor` (main.go:301) | B23 → HOLD |
| rate gate | prior candle page `X-Ratelimit-Remaining` exactly one value ≥ 1 | `requireRemaining` (main.go:321) | B25 → HOLD |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 @72 | `validateConfig(cfg) != nil` | none | HOLD invalid input | validate-config tests (main_test.go:300,306) |
| B2 @75 | ctx/reader/now/identity/builder nil | none | HOLD deps incomplete | — (constructor guard) |
| B3 @79 | `started.IsZero()` | none | HOLD zero clock | — |
| B4 @85 | `live()` before receipt open fails | none | HOLD collector unavailable | budget tests (:156) |
| B5 @89 | `deps.openReceipt == nil` | choose production `openReceipt` | continue | all run tests (nil → prod) |
| B6 @93 | receipt open error | none | HOLD receipt root | insecure receipt (:111) |
| B7 @98 (deferred) | `err != nil` at return | `store.taint()` writes `TAINTED` | — | every HOLD test asserts TAINTED |
| B8 @102 | store implements `setLive` | store rechecks liveness on each write | continue | implicit |
| B9 @108 | `live()` before identity fails | none | HOLD | budget tests |
| B10 @112 | `snapshotIdentity` error | none | HOLD pre-request identity | identity tests (:317,:354) |
| B11 @116 | `verifyPrescribedBinary` error | none | HOLD prescribed build identity | build identity tests |
| B12 @119 | `live()` after identity fails | none | HOLD | budget tests |
| B13 @122 | `writeJSON(preflight.json)` error | receipt write | HOLD write preflight | receipt tests |
| B14 @125 | `live()` after preflight fails | none | HOLD | — |
| B15 @130 | `len(before) != 0` | seed `seen` with initial cursor | continue | (:185) explicit-before repeat |
| B16 @135 | `for page := 0; page < 4` | loop | — | 4-page hard bound (`TestRunRecordsCapExhaustionAndStillCollectsOrderbookCalendarAndSeals` asserts exactly 4 candle calls) |
| B17 @137 | `boundedRequestContext` error (<15 s left) | none | HOLD candle page N | (:156) |
| B18 @142 | `reader.Candle` error | none | HOLD candle page N | `TestRunStopsOnReaderErrorWithoutFallback` |
| B19 @145 | `live()` after read fails | none | HOLD | `TestRunHoldsWhenClockExceedsOverallBudgetAfterFinalResponse` |
| B20 @148 | `validateSourceResult` error | none | HOLD page result | `TestRunHoldsOnNonStringRawCursorBeforeOrderbook` |
| B21 @151 | `writeResult(candle-NN)` error | receipt write | HOLD page receipt | `TestRunHoldsOnSecretLikeRawBodyBeforeNextRequest` |
| B22 @154 | `live()` after receipt fails | none | HOLD | — |
| B23 @159 | `nextCursor` error (loop / malformed) | none | HOLD page cursor | `TestRunHoldsOnCursorLoopBeforeOrderbook`, `TestRunStillHoldsOnCursorLoopAfterCapChange` |
| B24 @162 | `done` (raw null terminal) | `terminal = true`, break | continue | `TestRunCollectsTerminalCandleThenSingleOrderbookAndCalendar`, `TestRunRecordsNullTerminalInCandleCrawlRecord` |
| B25 @166 | `requireRemaining` error (also evaluated after page 4 — intentional strictness) | none | HOLD rate budget | `TestRunHoldsWhenEarlierPageRateBudgetIsExhausted`, `TestRunHoldsWhenPageFourRateBudgetIsExhausted` |
| B26 @175 | `!terminal` after loop (cap exhausted) | crawl record `terminal=cap_exhausted`, `last_cursor_sha256=sha256(page-4 cursor)` | continue (amended: **no HOLD**) | `TestRunRecordsCapExhaustionAndStillCollectsOrderbookCalendarAndSeals` |
| B27 @179 | `writeJSON(candle-crawl.json)` error | receipt write | HOLD candle crawl receipt | `TestRunHoldsAndTaintsWhenCandleCrawlRecordWriteFails` |
| B28 @182 | `live()` after crawl record fails | none | HOLD | not-applicable: redundant with the store's post-write live guard, unobservable through `run()` (adversary mutation h) |
| B29 @185 | orderbook `collectSingle` error | none | HOLD (propagated) | (:126 style) |
| B30 @190 | calendar `collectSingle` error | none | HOLD | — |
| B31 @196 | `live()` before post identity fails | none | HOLD | (:222) |
| B32 @200 | post `snapshotIdentity` error | none | HOLD post identity | — |
| B33 @204 | `!pre.equal(post)` | none | HOLD identity drift | (:212) |
| B34 @207 | `live()` after post identity fails | none | HOLD | (:222,:235) |
| B35 @210 | `store.seal` error | manifest write | HOLD seal | (:266) unexpected entry |
| B36 @213 | `live()` after seal fails | none | HOLD | — |
| B37 @216 | `verifySealed` error | none | HOLD sealed verification | (:278) |
| B38 @219 | `live()` before success fails | none | HOLD | — |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validateConfig` | 72:12 |
| `holdf` | 73:10 |
| `holdf` | 76:10 |
| `deps.now` | 78:13 |
| `started.IsZero` | 79:5 |
| `holdf` | 80:10 |
| `context.WithDeadline` | 82:35 |
| `Add` | 82:61 |
| `time.Now` | 82:61 |
| `cancelCollector` | 83:8 |
| `collectorLive` | 84:32 |
| `live` | 85:12 |
| `holdf` | 86:10 |
| `opener` | 92:16 |
| `holdf` | 94:10 |
| `(unnamed)` | 96:8 |
| `store.close` | 96:21 |
| `(unnamed)` | 97:8 |
| `store.taint` | 99:8 |
| `guarded.setLive` | 103:3 |
| `live` | 108:12 |
| `holdf` | 109:10 |
| `snapshotIdentity` | 111:14 |
| `holdf` | 113:10 |
| `(unnamed)` | 115:8 |
| `pre.close` | 115:21 |
| `verifyPrescribedBinary` | 116:12 |
| `holdf` | 117:10 |
| `live` | 119:12 |
| `holdf` | 120:10 |
| `store.writeJSON` | 122:12 |
| `receiptConfigFrom` | 122:71 |
| `holdf` | 123:10 |
| `live` | 125:12 |
| `holdf` | 126:10 |
| `bytes.Clone` | 129:12 |
| `len` | 130:5 |
| `string` | 131:8 |
| `boundedRequestContext` | 136:37 |
| `holdf` | 138:11 |
| `deps.reader.Candle` | 140:22 |
| `cancel` | 141:3 |
| `holdf` | 143:11 |
| `live` | 145:13 |
| `holdf` | 146:11 |
| `validateSourceResult` | 148:13 |
| `holdf` | 149:11 |
| `store.writeResult` | 151:13 |
| `fmt.Sprintf` | 151:31 |
| `holdf` | 152:11 |
| `live` | 154:13 |
| `holdf` | 155:11 |
| `nextCursor` | 158:24 |
| `holdf` | 160:11 |
| `requireRemaining` | 166:13 |
| `holdf` | 167:11 |
| `string` | 169:8 |
| `digestBytes` | 177:28 |
| `store.writeJSON` | 179:12 |
| `holdf` | 180:10 |
| `live` | 182:12 |
| `holdf` | 183:10 |
| `collectSingle` | 185:12 |
| `deps.reader.Orderbook` | 186:10 |
| `collectSingle` | 190:12 |
| `deps.reader.Calendar` | 191:10 |
| `live` | 196:12 |
| `holdf` | 197:10 |
| `snapshotIdentity` | 199:15 |
| `holdf` | 201:10 |
| `(unnamed)` | 203:8 |
| `post.close` | 203:21 |
| `pre.equal` | 204:6 |
| `holdf` | 205:10 |
| `live` | 207:12 |
| `holdf` | 208:10 |
| `store.seal` | 210:12 |
| `holdf` | 211:10 |
| `live` | 213:12 |
| `holdf` | 214:10 |
| `store.verifySealed` | 216:12 |
| `holdf` | 217:10 |
| `live` | 219:12 |
| `holdf` | 220:10 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `deps.reader.Candle/Orderbook/Calendar` | the only network reads (M-B0 seam) | one request per call, no retry/fallback; error → HOLD | ast.json calls; a112_mbus_read.go single `hc.Do` |
| `store.write*/seal/taint/verifySealed` | receipt capability | O_EXCL 0600 payloads, self-excluding manifest, taint on any HOLD | receipt_unix.go |
| `snapshotIdentity/verifyPrescribedBinary` | pre/post identity binding | drift → HOLD; no ambient GOROOT/PATH | identity.go |
| `collectorLive/boundedRequestContext` | deadline enforcement | 120 s overall, 15 s per request, ≥15 s remaining to start | main.go |

## State mutations and fallbacks

- No product/runtime state is touched; only the private `/tmp` receipt directory is written. There is no fallback path anywhere: every error is HOLD + taint.
- Amended 2026-08-16 (lot 0.7b.3): cap exhaustion (B26 @175) writes `candle-crawl.json` (`schema a112-mb-us-candle-crawl:v1`, `pages`, `terminal`, `last_cursor_sha256`) through the live-guarded `store.writeJSON` so it is part of the sealed manifest set, then falls through to orderbook/calendar/post identity/seal. Pre-amendment the same branch returned HOLD ("candle pagination did not reach raw null within four pages") — observed 3/3 on the 2026-08-16 human runs — because `nextBefore null` means end of retained history and an empty/session-end cursor can never satisfy B24 within four pages. The terminal-null path also writes the record (`terminal=null`, `pages` = pages read). Known intentional strictness: page-4 `requireRemaining` (B25) still HOLDs on remaining < 1 even though no 5th request follows.

## Safety conclusion

- Safe edit boundary (approved amendment, now applied): B26 plus the new B27/B28 write/live checks only; every other branch, the 4-page cap, the rate gate, the null-only terminal typing and the no-retry rule are unchanged in intent.
- High-risk impact: no — uninstalled measurement tool with zero production callers; no order/auth/config/runtime surface.
