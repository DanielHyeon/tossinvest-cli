# Branch Test Map: `EntryGate.Block`

- Source: `internal/execgw/retry.go` (532-539); current
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/execgw` 의 시험 59 개를 하나씩 돌린 커버 프로필(-covermode=set)에서 그 분기 본문 블록을 실행한 시험. 「합집합」은 패키지 전체 한 판.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 535:2 | `if _, exists := g.latches[reason]; !exists {` | `TestARejectedCredentialTightensTheOperatingMode` · `TestARelaxationClearsTheLatch` · `TestATighteningBetweenIssuanceAndSubmissionIsRefusedAtTheGateway` (+18) | 기존 분기(회귀 핀) | 블록 535.45-538.3 을 시험 21 개가 실행, 전부 PASS |
