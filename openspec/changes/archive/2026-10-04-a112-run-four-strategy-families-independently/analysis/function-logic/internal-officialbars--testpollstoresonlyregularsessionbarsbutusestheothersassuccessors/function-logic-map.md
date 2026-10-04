# Function Logic Map: `TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors`

- Source: `internal/officialbars/producer_test.go`
- Source SHA-256: `6c6f064d3ecceb46ed4df8546cb5d1f16c40e2831dbfa99c223c0c99f2b5eda4` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors(t *testing.T)`
- Source range: `405:1`–`444:2` (ast.json `start`/`end`), revision `current`
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 6, returns 0, calls 39, assignments 7, defers 0, go statements 0.
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- What the correction actually changed here: **only two inline comments**. Every fixture argument, every assertion and every expected value is unchanged, because the test was already written in open-instant terms through the `usBar` helper — `usBar` now emits `Timestamp = open + 1m` and `adoptPage` subtracts it back, so the arithmetic the test asserts is identical. The comment on the bar at `usClose` moved from "16:00 ET 봉" to "16:00 ET**에 여는** 봉" and the successor comment from "15:59 봉의 successor" to "15:59**에 여는** 봉의 successor" — i.e. they now say *opens at* rather than leaving "the 16:00 bar" ambiguous between label and open.
- Risk scan: `risk-pattern-report.md` (no configured risk pattern matched).

## Inputs and invariants

- The regression for design §5 and `PollClosedBars` B36: bars outside `[regularOpen, regularClose)` are **observed and counted and usable as successors**, but never stored.
- Five fixture bars, all named by the instant they **open**: `usClose+1m` (16:01, after hours), `usClose` (16:00, after hours — the close instant itself is excluded by the half-open window), `usClose−1m` (15:59, the last regular minute), `usOpen` (09:30, the first regular minute) and `usOpen−1m` (09:29, pre-market). `pollAt` is `usClose + 5m` so the whole session is in the past.
- Expected result: `Observed == 5`, `Admitted == 2`. Two of the five fall outside the window; of the three inside, the newest observed bar overall (`usClose+1m`) is withheld by decision 6 — and it is outside the window anyway — leaving 09:30 and 15:59 stored.
- The successor assertion is the sharp one: the stored 15:59 bar's `SuccessorOpenAtMS` must be the **16:00 after-hours** bar. That is the whole point of "counted but not stored" — an out-of-window bar is still proof that the bar below it closed.
- Gap expectation: exactly one gap of 388 minutes from `usOpen+1m` to `usClose−2m`, i.e. the entire regular session between the two stored bars.
- Clock and store are deterministic: `marketclock.NewFake(pollAt)` drives a real `strategyevidence.Store` in `t.TempDir()`, and `usSeries(t, store, pollAt, pollAt)` reads it back.

## Branches and early returns

The AST reports 6 branches and **no return nodes** — the function ends by falling off the bottom, and every branch is a `t.Fatalf` assertion guard.

| Branch | AST kind | Source location | What the guard proves | Test disposition |
|---|---|---|---|---|
| B1 | if | 422:2 | `PollClosedBars` returned an error → `t.Fatalf` | proves the mixed in-window/out-of-window page is **accepted**, not refused; an out-of-window bar must not be a contract violation |
| B2 | if | 425:2 | `Observed != 5 \|\| Admitted != 2` → `t.Fatalf` | proves the two counters diverge exactly as design §5 requires: all five bars are observed, only two are admitted |
| B3 | if | 429:2 | `len(series.Bars) != 2` → `t.Fatalf` | proves the store agrees with the counter — `Admitted` is not merely a number the producer reported about itself |
| B4 | if | 432:2 | the two stored `OpenAtMS` are not `usOpen` and `usClose−1m` → `t.Fatalf` | proves **which** two minutes were kept, in open-instant space. This is the row decision 30 could have broken silently: with the conversion removed, the stored opens would be one minute late and this guard fires |
| B5 | if | 437:2 | the 15:59 bar's `SuccessorOpenAtMS` is not `usClose` → `t.Fatalf` | proves an **out-of-window** bar still serves as the successor that makes an in-window bar admissible — the invariant that separates "not stored" from "not seen" |
| B6 | if | 440:2 | the gap report is not exactly one 388-minute gap from `usOpen+1m` to `usClose−2m` → `t.Fatalf` | proves `minuteGaps` clamps to the regular window and reports on converted open instants, and that a gap never suppresses a stored bar |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Parallel` | 406:2 |
| `context.Background` | 407:9 |
| `usClose.Add` | 408:12 |
| `marketclock.NewFake` | 409:11 |
| `openTestStore` | 410:11 |
| `readerOf` | 411:12 |
| `usPage` | 411:21 |
| `usBar` | 412:3 |
| `usClose.Add` | 412:12 |
| `usBar` | 413:3 |
| `usBar` | 414:3 |
| `usClose.Add` | 414:12 |
| `usBar` | 415:3 |
| `usBar` | 416:3 |
| `usOpen.Add` | 416:12 |
| `PollClosedBars` | 418:17 |
| `usCalendarAt` | 420:13 |
| `usClose.Add` | 420:29 |
| `t.Fatalf` | 423:3 |
| `t.Fatalf` | 426:3 |
| `usSeries` | 428:12 |
| `len` | 429:5 |
| `t.Fatalf` | 430:3 |
| `len` | 430:32 |
| `uint64` | 432:40 |
| `usOpen.UnixMilli` | 432:47 |
| `uint64` | 433:38 |
| `UnixMilli` | 433:45 |
| `usClose.Add` | 433:45 |
| `t.Fatalf` | 434:3 |
| `uint64` | 437:49 |
| `usClose.UnixMilli` | 437:56 |
| `t.Fatalf` | 438:3 |
| `len` | 440:5 |
| `result.Gaps.From.Equal` | 441:4 |
| `usOpen.Add` | 441:30 |
| `result.Gaps.To.Equal` | 441:59 |
| `usClose.Add` | 441:83 |
| `t.Fatalf` | 442:3 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `t.Parallel`, `context.Background` | 406, 407 | ordinary harness; the test owns its own store so parallel execution is safe |
| `usClose.Add`, `usOpen.Add` | 408, 412–416, 433, 441 | every fixture instant and every expectation is expressed as an offset from the session edges, in **open** space |
| `marketclock.NewFake(pollAt)`, `openTestStore` | 409, 410 | deterministic clock and a real `strategyevidence.Store` under `t.TempDir()` |
| `readerOf(usPage(…))` | 411 | a single terminal page (empty cursor) carrying all five bars |
| `usBar` ×5 | 412–416 | the re-based fixture builder: takes the open instant, writes `open + 1m` on the wire (`internal-officialbars--usbar`) |
| `usCalendarAt(t, usClose−1h)` | 420 | a snapshot fetched inside the session, so the ruling 24 calendar gate passes |
| `PollClosedBars` | 418 | the function under test; its enumeration lives in `internal-officialbars--pollclosedbars` |
| `usSeries(t, store, pollAt, pollAt)` | 428 | reads the sealed series back for B3–B5 |
| `t.Fatalf` ×6 | 423, 426, 430, 434, 438, 442 | one per guard; the first violation stops the test |

## State mutations and fallbacks

- Locals only (7 AST assignments): `ctx`, `pollAt`, `clock`, `store`, `reader`, `result`/`err`, `series`. No package-level state, no goroutines, no defers.
- The only side effect is a SQLite store under `t.TempDir()`, written by the producer under test and read back once. No HTTP server, no credential, no network.
- No fallback: all six guards are `t.Fatalf`, so the first failed property stops the test rather than letting a later guard report a misleading second failure.

## Safety conclusion

- Safe edit boundary: a test function whose assertions the correction did not touch. Its comments were corrected; its expectations are the same values they were at the frozen base.
- High-risk impact: no. No production branch sits behind these rows — they are assertions over `PollClosedBars` B35/B36 and `minuteGaps`.
- Why it still matters to the decision-30 record: this is the test that would have caught a one-minute shift **without** anyone writing a new test, because B4 pins the stored minutes by value. That it passes unchanged is evidence that the conversion landed in `adoptPage` and nowhere else — the helper re-basing and the producer's subtraction cancel exactly, as they must.
- Its US window boundary is asserted in open space; the complementary label-space boundary (the bar *labelled* 16:00 opening at 15:59) is asserted separately by `TestPollTreatsTheClosingLabelAsTheLastRegularBar`.
