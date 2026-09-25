# Function Logic Map: `TestNoShippedFileClaimsProtection`

- Source: `internal/execgw/protection_test.go` (121-164)
- Revision: current — HEAD `648df8ef`; source_sha256 `14894b5dae3d64fc2c12b8242f85410b611bb6b2dc90515ca67d9270eccb4def`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 11 · returns 6 · calls 22
- Exact AST return positions: 139:4, 142:5, 144:4, 146:4, 150:4, 159:3
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | TestNoShippedFileClaimsProtection | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | for at 125:2 | `for i := 0; i < optionsType.NumField(); i++ {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B2 | if at 127:3 | `if field.IsExported() && (field.Type == reflect.TypeOf(execgw.ProtectionReadiness("")) \|\|` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B3 | switch at 137:3 | `switch {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B4 | case at 138:3 | `case err != nil:` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B5 | case at 140:3 | `case d.IsDir():` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B6 | if at 141:4 | `if name := d.Name(); name == ".git" \|\| name == "vendor" \|\| name == "node_modules" {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B7 | case at 145:3 | `case !strings.HasSuffix(path, ".go") \|\| strings.HasSuffix(path, "_test.go"):` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B8 | if at 149:3 | `if readErr != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B9 | range at 152:3 | `for _, name := range forbidden {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B10 | if at 153:4 | `if !strings.Contains(string(src), name) {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B11 | if at 161:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | mapped AST control flow | bounded to function | typed return | affected regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mapped dependencies | preserve function contract | caller handles error | CodeGraph + AST |

## State mutations and fallbacks

- No authority broadening; current behavior is covered by focused tests.

## Safety conclusion

- Safe edit boundary: TestNoShippedFileClaimsProtection only.
- High-risk impact: reviewed and regression-tested.
