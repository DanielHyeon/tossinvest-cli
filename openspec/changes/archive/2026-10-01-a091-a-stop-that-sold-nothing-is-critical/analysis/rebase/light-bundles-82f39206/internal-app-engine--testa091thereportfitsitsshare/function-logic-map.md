# Function Logic Map: `TestA091TheReportFitsItsShare`

- Source: `internal/app/engine/a091_replay_test.go` (`213`–`363`)
- Qualified: `TestA091TheReportFitsItsShare`
- AST evidence: `ast.json` (`source_sha256` d9ca1043f02ee9e9…) — `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 26

**역할(경량 번들 — 시험 함수).** 보고 몫 실측(소유 칸 ≤ 750ms · 승인 칸 < 5s) — i2 에서 승인 호출 구간과의 겹침 단언 · 연결 풀 B2 칸을 더함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하네스 · 가짜 부품 | 시험 고정 | 시험 파일 | 단언 실패 → 시험 실패 |

## Branches and early returns

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | for | `:228` `for i := 0; i < 100; i++ {` |
| B2 | if | `:229` `if _, err := r.journal.EnqueueAlert(ctx, journal.Alert{EventKey: "a091-backlog-" + time.Now().Format(time.RFC3339Nano) + fmt.Sprint(i),` |
| B3 | for | `:245` `for {` |
| B4 | select | `:246` `select {` |
| B5 | if | `:252` `if err := r.notifier.Acknowledge(ctx, "a091-operator"); err != nil {` |
| B6 | for | `:274` `for {` |
| B7 | select | `:275` `select {` |
| B8 | if | `:280` `if _, err := r.journal.EnqueueAlert(ctx, journal.Alert{EventKey: "a091-pool-" + time.Now().Format(time.RFC3339Nano),` |
| B9 | range | `:290` `for _, c := range []cell{` |
| B10 | if | `:301` `if c.fail {` |
| B11 | if | `:305` `if c.before != nil {` |
| B12 | for | `:311` `for i := 0; i < reps; i++ {` |
| B13 | if | `:315` `if d := end.Sub(start); d > worst {` |
| B14 | if | `:322` `if stop != nil {` |
| B15 | if | `:324` `if load == 0 {` |
| B16 | if | `:329` `if c.fail {` |
| B17 | else | `:333` `} else if n := a091OutboxCount(t, r, obs.EventExitStopSoldNothing); n != 1 {` |
| B18 | if | `:330` `if got := r.mode(); got != journal.ModeEntryBlocked {` |
| B19 | if | `:333` `} else if n := a091OutboxCount(t, r, obs.EventExitStopSoldNothing); n != 1 {` |
| B20 | if | `:337` `if !c.owned {` |
| B21 | range | `:340` `for _, cy := range cycles {` |
| B22 | range | `:341` `for _, w := range ackWindows {` |
| B23 | if | `:342` `if cy[0].Before(w[1]) && w[0].Before(cy[1]) {` |
| B24 | if | `:349` `if overlap == 0 {` |
| B25 | if | `:353` `if worst >= a091ObservationPeriod {` |
| B26 | if | `:358` `if worst > a091ReportShare {` |

Exact AST return positions: `223:36`, `224:36`, `232:6`, `248:6`, `254:6`, `264:3`, `264:77`, `277:6`, `283:6`, `288:3`, `288:47`, `356:5`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(63). 생산 경로에 닿는 것은 하네스가 조립한 `ExitObserver` · 알림기 · 원장뿐 — 브로커 0.

## State mutations and fallbacks

임시 원장 · 버퍼만(시험 범위).

## Safety conclusion

- 생산 코드가 아니다 — 생산 동작 변화 0. 이 번들은 base 재고정 창의 증거 완결을 위한 경량 기록이다.
