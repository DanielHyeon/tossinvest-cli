# Branch Test Map: isClaim

- Source: `internal/execgw/protection_test.go` (214-240); base `775c37cb` (HEAD 에 함수 없음)
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | switch at 216:2 | `switch {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀 |
| B2 | case at 217:2 | `case strings.HasPrefix(trimmed, name+" "):` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀 |
| B3 | case at 219:2 | `case strings.HasPrefix(trimmed, "var "+name+" "), strings.HasPrefix(trimmed, "var "+name+" ="):` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀 |
| B4 | case at 221:2 | `case strings.HasPrefix(trimmed, "var wiredForTest ="), strings.HasPrefix(trimmed, "var WiredProtectionForTe...` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀 |
| B5 | if at 225:2 | `if at < 0 {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀 |
| B6 | switch at 231:2 | `switch {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀 |
| B7 | case at 232:2 | `case strings.HasSuffix(before, "=="), strings.HasSuffix(before, "!="):` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀 |
| B8 | case at 234:2 | `case strings.HasSuffix(before, "case"), strings.HasSuffix(before, "return"):` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀 |
| B9 | case at 236:2 | `case strings.HasSuffix(before, "="), strings.HasSuffix(before, ":="), strings.HasSuffix(before, ":"):` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀 |
