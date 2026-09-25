# Branch Test Map: buildGateway

- Source: `internal/app/engine/gateway.go` (234-355); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 239:2 | `if err := checkProjectionWired(in.journal); err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 239.57-241.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B2 | if at 261:2 | `if err := tracker.Restore(ctx); err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 261.45-263.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B3 | if at 269:2 | `if err := restoreAlertEntryLatch(ctx, in.journal, entry); err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 269.71-271.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B4 | if at 293:2 | `if err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 293.16-295.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B5 | if at 317:2 | `if err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 317.16-319.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
