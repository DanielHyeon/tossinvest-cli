# Function Logic Map: `TestImportsStayInsideTheAllowlist` (tools/a128-jev-shadow-probe)

- Source: `tools/a128-jev-shadow-probe/static_test.go`(편집 뒤 198~220행). 편집 전 AST `../../pre-edit-ast/a128--testimportsstayinsidetheallowlist.ast.json`(162~181행, 분기 3).
- AST evidence: `ast.json`(편집 뒤, 분기 3 그대로)
- Risk scan: `risk-pattern-report.md`(일치 0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 비시험 소스의 import | 표준 라이브러리 + `allowedModule` | 이 함수 | 경로에 `.` 이 있고 목록 밖이면 Errorf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (212) | `range` 파일 | — | — | 자신 |
| B2 (213) | `range` import | — | — | 자신 |
| B3 (215) | 모듈 경로이고 목록 밖 | — | Errorf | 자신(편집 전 두 예산 경로로 RED, 변이 a128-e-import) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `nonTestSources`·`strconv.Unquote`·`strings.Contains` | 대상 열거·경로 해독 | — | AST |

## State mutations and fallbacks

- 없음.

## 편집

- `allowedModule` 에 `internal/accountbudget`·`internal/accountbudgetruntime` 두 경로와 한글 주석 한 줄. 실패 문구의 괄호 목록에 두 이름을 더해 사실에 맞춤(proposal What Changes). 분기 구조 무변.

- 보강(2026-11-15)에서 이 함수 본문은 바뀌지 않았고 같은 파일의 위쪽 편집으로 행만 내려가 AST 를 다시 뽑음.

## Safety conclusion

- Safe edit boundary: 허용 경로 두 개.
- High-risk impact: 없음(시험).
