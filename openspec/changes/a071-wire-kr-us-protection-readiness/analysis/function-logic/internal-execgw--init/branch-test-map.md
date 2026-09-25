# Branch Test Map: init

- Source: `internal/execgw/export_test.go` (135-139); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | branchless happy path at 135:1 | `func init() {` | 패키지 전체 시험 | 5.1 에서 재실행 안 함 | 패키지 시험 바이너리 초기화 — 이 패키지의 모든 시험이 지나감; 패키지 시험 222/222 PASS (시험별 실행) |
