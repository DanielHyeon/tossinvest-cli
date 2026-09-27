# Branch Test Map: `alertDeliverer.cycle`

- Source: `internal/app/engine/alertdelivery.go` (235-264); current
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/app/engine` 의 시험 498 개를 하나씩 돌린 커버 프로필(-covermode=set)에서 그 분기 본문 블록을 실행한 시험. 「합집합」은 패키지 전체 한 판.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 243:2 | `if err != nil {` | `TestACancelledEngineIsNotALedgerFault` · `TestAClearAroundTheLimitthListingFailureStillEscalates` · `TestASuccessfulListingBreaksTheListingRun` (+2) | a124 RED — 편집 전에 없던 분기 | 블록 243.16-247.3 을 시험 5 개가 실행, 전부 PASS |
| B2 | if at 249:2 | `if len(pending) < d.batch() {` | `TestAClaimThatFindsTheRowSettledEndsItsRun` · `TestAClearAfterTheEvidenceLeavesWhatAnOnTimeLatchWouldHave` · `TestAClearAfterTheLimitthRecordFailureStillEscalates` (+59) | a124 RED — 편집 전에 없던 분기 | 블록 249.30-253.3 을 시험 62 개가 실행, 전부 PASS |
| B3 | range at 254:2 | `for _, alert := range pending {` | `TestAClaimThatFindsTheRowSettledEndsItsRun` · `TestAClearAfterTheEvidenceLeavesWhatAnOnTimeLatchWouldHave` · `TestAClearAfterTheLimitthRecordFailureStillEscalates` (+60) | a124 RED — 편집 전에 없던 분기 | 블록 254.32-258.23 을 시험 63 개가 실행, 전부 PASS |
| B4 | if at 258:3 | `if ctx.Err() != nil {` | `TestAStopAlertDoesNotWaitBehindTheBacklog` · `TestJudgingTransactionsDelayTheExitCycleOnlyWithinTheFixedMargin` · `TestTheExitCycleDoesNotLengthenWhileTheSenderIsStuckInTheTransport` (+1) | a124 RED — 편집 전에 없던 분기 | 블록 258.23-260.4 을 시험 4 개가 실행, 전부 PASS |
