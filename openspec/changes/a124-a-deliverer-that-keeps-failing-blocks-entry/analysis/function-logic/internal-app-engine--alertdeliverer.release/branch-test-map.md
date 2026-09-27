# Branch Test Map: `alertDeliverer.release`

- Source: `internal/app/engine/alertdelivery.go` (600-606); current
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/app/engine` 의 시험 498 개를 하나씩 돌린 커버 프로필(-covermode=set)에서 그 분기 본문 블록을 실행한 시험. 「합집합」은 패키지 전체 한 판.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 603:2 | `if _, err := d.led().ReleaseAlertClaim(relCtx, id, token); err != nil {` | 없음 | a124 RED — 편집 전에 없던 분기 | 블록 603.72-605.3 을 어느 시험도 실행하지 않음 |
