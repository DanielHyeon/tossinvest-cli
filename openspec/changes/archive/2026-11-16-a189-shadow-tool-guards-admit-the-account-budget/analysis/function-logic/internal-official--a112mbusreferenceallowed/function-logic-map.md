# Function Logic Map: `a112MBUSReferenceAllowed`

- Source: `internal/official/a112_mbus_static_test.go`(편집 뒤 489~499행). 편집 전 AST `../../pre-edit-ast/official--a112mbusreferenceallowed.ast.json`(base 7d1e3163a, 489~498행).
- AST evidence: `ast.json`(편집 뒤, 분기 2·return 3)
- Risk scan: `risk-pattern-report.md`(일치 0)
- 작성 시점: 편집 전 AST 는 1단계(tasks 1.4), 이 번들은 편집 뒤(tasks 2.3). 시험 파일의 기존 함수 본문을 고쳐 gate ⑤ 대상임.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 저장소 루트 기준 상대 경로(호출자가 `filepath.Rel` 로 만듦) | `a112MBUSRepositoryForbiddenRefs`·픽스처 시험 | true 면 호출자가 그 파일을 파싱하지 않고 위반 0 |

불변식: 허용은 M-B0 정의 파일 셋, M-B0 시험(`internal/official/a112_mbus_*_test.go`), 정확한 도구(`tools/a112-mb-us-source/`), 그리고 a166 런타임 시험 파일 하나(정확 경로)뿐임.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (491행) | `tools/a112-mb-us-source/` 접두사 | 없음 | true(492) | `TestA112MBUSStaticNoProductCallerOrForbiddenPath`(저장소 순회가 도구 파일을 지남) |
| B2 (494행) | `internal/official/a112_mbus_` 접두사 + `_test.go` 접미사 | 없음 | true(495) | 같은 시험(순회가 `a112_mbus_*_test.go` 를 지남) |
| 마지막 return (498행) | 정의 파일 셋 또는 `internal/accountbudgetruntime/v2_token_test.go` 와 정확히 같음 | 없음 | 비교 결과 | true: 순회의 `v2_token_test.go`(g4 가 되돌리면 실패), `TestA189MBUSGuardAdmitsOnlyTheExactA166RuntimeTestFile` 허용 사례. false: 나머지 모든 파일, 픽스처 시험 4개 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `filepath.ToSlash` | 구분자 통일 | 없음 | AST |
| `strings.HasPrefix`·`HasSuffix` | B1·B2 | 없음 | AST |

호출자: `a112MBUSFindForbiddenRefs` 하나(CodeGraph 1.6.0, `../../code-context/codegraph-baseline.md`).

## State mutations and fallbacks

- 없음(순수 함수).

## 편집

- 마지막 return 의 `||` 사슬 끝에 `path == "internal/accountbudgetruntime/v2_token_test.go"` 를 더하고 위에 한글 주석 한 줄. B1·B2 무편집. 분기 수 2 그대로, return 행이 497→498.

## Safety conclusion

- Safe edit boundary: 마지막 비교식 하나. 접두사·접미사 조건은 그대로.
- High-risk impact: 없음(시험 파일, 운영 바이너리에 안 들어감).
