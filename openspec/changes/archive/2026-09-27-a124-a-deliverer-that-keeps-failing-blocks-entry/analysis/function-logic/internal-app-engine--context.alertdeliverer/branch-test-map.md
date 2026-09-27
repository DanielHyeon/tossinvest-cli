# Branch Test Map: `Context.AlertDeliverer`

- Source: `internal/app/engine/auxiliary.go` (156-186); current
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/app/engine` 의 시험 498 개를 하나씩 돌린 커버 프로필(-covermode=set)에서 그 분기 본문 블록을 실행한 시험. 「합집합」은 패키지 전체 한 판.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 157:2 | `if c == nil \|\| c.Journal == nil \|\| c.Entry == nil {` | 없음 | 기존 분기(회귀 핀) | 블록 157.52-160.3 을 어느 시험도 실행하지 않음 |
| B2 | if at 161:2 | `if clk == nil {` | 없음 | 기존 분기(회귀 핀) | 블록 161.16-163.3 을 어느 시험도 실행하지 않음 |
| B3 | if at 165:2 | `if c.Notifier != nil {` | `TestAStopAlertDoesNotWaitBehindTheBacklog` · `TestJudgingTransactionsDelayTheExitCycleOnlyWithinTheFixedMargin` · `TestTheDeliveredNotificationCarriesTheRecordedAlert` (+3) | 기존 분기(회귀 핀) | 블록 165.23-167.3 을 시험 6 개가 실행, 전부 PASS |
