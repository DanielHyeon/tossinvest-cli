# Branch Test Map: `TestA091TheReportFitsItsShare`

- Source: `internal/app/engine/a091_replay_test.go`

> 시험 함수의 분기는 그 시험이 돌 때 지나간다 — Test 열은 그 시험(또는 도우미를 부르는 시험)이다.

| Branch | 조건 | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:228` `for i := 0; i < 100; i++ {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B2 | `:229` `if _, err := r.journal.EnqueueAlert(ctx, journal.Alert{EventKey: "a091-backlog-" + time.Now().Format(time.RFC3339Nano) + fmt.Sprint(i),` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B3 | `:245` `for {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B4 | `:246` `select {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B5 | `:252` `if err := r.notifier.Acknowledge(ctx, "a091-operator"); err != nil {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B6 | `:274` `for {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B7 | `:275` `select {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B8 | `:280` `if _, err := r.journal.EnqueueAlert(ctx, journal.Alert{EventKey: "a091-pool-" + time.Now().Format(time.RFC3339Nano),` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B9 | `:290` `for _, c := range []cell{` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B10 | `:301` `if c.fail {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B11 | `:305` `if c.before != nil {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B12 | `:311` `for i := 0; i < reps; i++ {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B13 | `:315` `if d := end.Sub(start); d > worst {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B14 | `:322` `if stop != nil {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B15 | `:324` `if load == 0 {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B16 | `:329` `if c.fail {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B17 | `:333` `} else if n := a091OutboxCount(t, r, obs.EventExitStopSoldNothing); n != 1 {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B18 | `:330` `if got := r.mode(); got != journal.ModeEntryBlocked {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B19 | `:333` `} else if n := a091OutboxCount(t, r, obs.EventExitStopSoldNothing); n != 1 {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B20 | `:337` `if !c.owned {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B21 | `:340` `for _, cy := range cycles {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B22 | `:341` `for _, w := range ackWindows {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B23 | `:342` `if cy[0].Before(w[1]) && w[0].Before(cy[1]) {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B24 | `:349` `if overlap == 0 {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B25 | `:353` `if worst >= a091ObservationPeriod {` | `TestA091TheReportFitsItsShare` | n/a | yes |
| B26 | `:358` `if worst > a091ReportShare {` | `TestA091TheReportFitsItsShare` | n/a | yes |
