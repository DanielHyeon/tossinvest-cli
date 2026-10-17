# Branch Test Map: TestExitsStayOpenWhileRecoveryIsIncomplete

- Source: `internal/reconcile/recovery_test.go` (443-485); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 456:2 | `if err != nil {` | `TestExitsStayOpenWhileRecoveryIsIncomplete` | 5.1 에서 재실행 안 함 | 시험 자신; TestExitsStayOpenWhileRecoveryIsIncomplete PASS (시험별 실행) |
| B2 | if at 459:2 | `if _, err := reconcile.New(recoveryOptions(j, gate, recoveryCollector(nil, nil))); err != nil {` | `TestExitsStayOpenWhileRecoveryIsIncomplete` | 5.1 에서 재실행 안 함 | 시험 자신; TestExitsStayOpenWhileRecoveryIsIncomplete PASS (시험별 실행) |
| B3 | if at 462:2 | `if rejected := gate.CheckEntry(); rejected == nil {` | `TestExitsStayOpenWhileRecoveryIsIncomplete` | 5.1 에서 재실행 안 함 | 시험 자신; TestExitsStayOpenWhileRecoveryIsIncomplete PASS (시험별 실행) |
| B4 | if at 472:2 | `if err != nil {` | `TestExitsStayOpenWhileRecoveryIsIncomplete` | 5.1 에서 재실행 안 함 | 시험 자신; TestExitsStayOpenWhileRecoveryIsIncomplete PASS (시험별 실행) |
| B5 | if at 475:2 | `if _, err := gw.Cancel(context.Background(), execgw.CancelRequest{` | `TestExitsStayOpenWhileRecoveryIsIncomplete` | 5.1 에서 재실행 안 함 | 시험 자신; TestExitsStayOpenWhileRecoveryIsIncomplete PASS (시험별 실행) |
| B6 | if at 482:2 | `if broker.cancels != 1 {` | `TestExitsStayOpenWhileRecoveryIsIncomplete` | 5.1 에서 재실행 안 함 | 시험 자신; TestExitsStayOpenWhileRecoveryIsIncomplete PASS (시험별 실행) |
