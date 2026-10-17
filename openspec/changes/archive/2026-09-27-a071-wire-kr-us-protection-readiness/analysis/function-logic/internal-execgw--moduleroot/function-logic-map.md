# Function Logic Map: `moduleRoot`

- Source: `internal/execgw/protection_test.go` (167-183)
- Revision: current — HEAD `648df8ef`; source_sha256 `14894b5dae3d64fc2c12b8242f85410b611bb6b2dc90515ca67d9270eccb4def`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 4 · returns 1 · calls 7
- Exact AST return positions: 175:4
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | moduleRoot | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 170:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B2 | for at 173:2 | `for {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B3 | if at 174:3 | `if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B4 | if at 178:3 | `if parent == dir {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

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

- Safe edit boundary: moduleRoot only.
- High-risk impact: reviewed and regression-tested.
