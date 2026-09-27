# Branch Test Map: Gateway.protection

- Source: `internal/execgw/protection.go` (95-100); base `775c37cb` (HEAD 에 함수 없음)
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 96:2 | `if g.protectionOverride != nil {` | `TestARaisingMutationIsRefusedWhileProtectionIsUnwired` · `TestProductionDefaultRefusesKRAndUSBuyBeforeBroker` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 판독기 `Gateway.protection` 을 지우고 `checkProtection` 이 봉인 어댑터를 읽게 함 |
