# Function Logic Map: `TestImportsStayInsideTheAllowlist` (tools/a140-breakout-signal-shadow)

- Source: `tools/a140-breakout-signal-shadow/static_test.go`(편집 뒤 123~143행). 편집 전 AST `../../pre-edit-ast/a140--testimportsstayinsidetheallowlist.ast.json`(118~136행, 분기 4).
- AST evidence: `ast.json`(편집 뒤, 분기 4 그대로)
- Risk scan: `risk-pattern-report.md`(일치 0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 비시험 소스의 import | 표준 라이브러리 + `module` 접두사 + `allowed` 접미사 | 이 함수 | 목록 밖 Errorf |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (132) | `range` 파일 | — | — | 자신 |
| B2 (133) | `range` import | — | — | 자신 |
| B3 (135) | 표준 라이브러리 | — | continue | 자신 |
| B4 (138) | 모듈 밖이거나 접미사 목록 밖 | — | Errorf | 자신(편집 전 두 예산 경로 RED, 변이 a140-e-import) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `nonTestSources`·`strconv.Unquote`·`strings.*` | 열거·판정 | — | AST |

## State mutations and fallbacks

- 없음.

## 편집

- `allowed` 에 `accountbudget`·`accountbudgetruntime` 와 한글 주석 한 줄. 분기 무변.

## Safety conclusion

- Safe edit boundary: 허용 접미사 둘.
- High-risk impact: 없음(시험).
