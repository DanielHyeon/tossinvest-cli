# Branch Test Map: TestNoShippedFileClaimsProtection

- Source: `internal/execgw/protection_test.go` (121-164); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | for at 125:2 | `for i := 0; i < optionsType.NumField(); i++ {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B2 | if at 127:3 | `if field.IsExported() && (field.Type == reflect.TypeOf(execgw.ProtectionReadiness("")) \|\|` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B3 | switch at 137:3 | `switch {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B4 | case at 138:3 | `case err != nil:` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B5 | case at 140:3 | `case d.IsDir():` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B6 | if at 141:4 | `if name := d.Name(); name == ".git" \|\| name == "vendor" \|\| name == "node_modules" {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B7 | case at 145:3 | `case !strings.HasSuffix(path, ".go") \|\| strings.HasSuffix(path, "_test.go"):` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B8 | if at 149:3 | `if readErr != nil {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B9 | range at 152:3 | `for _, name := range forbidden {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B10 | if at 153:4 | `if !strings.Contains(string(src), name) {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
| B11 | if at 161:2 | `if err != nil {` | `TestNoShippedFileClaimsProtection` | 5.1 에서 재실행 안 함 | 시험 자신; TestNoShippedFileClaimsProtection PASS (시험별 실행) |
