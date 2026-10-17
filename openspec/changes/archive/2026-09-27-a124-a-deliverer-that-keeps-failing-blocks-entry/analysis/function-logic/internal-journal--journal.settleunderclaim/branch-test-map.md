# Branch Test Map: `Journal.settleUnderClaim`

- Source: `internal/journal/alert_claim.go` (314-355); current
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/journal` 의 시험 158 개를 하나씩 돌린 커버 프로필(-covermode=set)에서 그 분기 본문 블록을 실행한 시험. 「합집합」은 패키지 전체 한 판.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 317:2 | `if strings.TrimSpace(token) == "" {` | `TestSettlingWithoutAClaimIsRefused` | a124 RED — 편집 전에 없던 분기 | 블록 317.36-319.3 을 시험 1 개가 실행, 전부 PASS |
| B2 | if at 321:2 | `if err != nil {` | 없음 | 기존 분기(회귀 핀) | 블록 321.16-323.3 을 어느 시험도 실행하지 않음 |
| B3 | if at 327:2 | `if err != nil {` | 없음 | a124 RED — 편집 전에 없던 분기 | 블록 327.16-329.3 을 어느 시험도 실행하지 않음 |
| B4 | if at 331:2 | `if err != nil {` | 없음 | a124 RED — 편집 전에 없던 분기 | 블록 331.16-333.3 을 어느 시험도 실행하지 않음 |
| B5 | if at 334:2 | `if n == 1 {` | `TestAFailedAttemptReturnsTheAttemptsItCommitted` · `TestAFailedAttemptsReadLeavesTheLedgerAsItWas` · `TestARearmedRowCountsFromZero` (+17) | a124 RED — 편집 전에 없던 분기 | 블록 334.12-338.17 을 시험 20 개가 실행, 전부 PASS |
| B6 | if at 338:3 | `if err != nil {` | `TestAFailedAttemptsReadLeavesTheLedgerAsItWas` | 기존 분기(회귀 핀) | 블록 338.17-340.4 을 시험 1 개가 실행, 전부 PASS |
| B7 | if at 341:3 | `if err := tx.Commit(); err != nil {` | `TestAFailedAttemptsReadLeavesTheLedgerAsItWas` | 기존 분기(회귀 핀) | 블록 341.37-343.4 을 시험 1 개가 실행, 전부 PASS |
| B8 | if at 348:2 | `if err != nil {` | 없음 | 기존 분기(회귀 핀) | 블록 348.16-350.3 을 어느 시험도 실행하지 않음 |
| B9 | if at 351:2 | `if err := tx.Commit(); err != nil {` | 없음 | 기존 분기(회귀 핀) | 블록 351.36-353.3 을 어느 시험도 실행하지 않음 |
