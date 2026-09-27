# Branch Test Map: Context.OfficialClientForTest

- Source: `internal/app/engine/export_test.go` (26-26); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | branchless happy path at 26:1 | `func (c *Context) OfficialClientForTest() *official.Client { return c.official }` | 패키지 전체 시험 | 5.1 에서 재실행 안 함 | 호출하는 시험(텍스트 탐색); 패키지 시험 439/439 PASS (시험별 실행) |
