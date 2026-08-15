# Function Logic Map: `TestReleaseErrorsAreMappedToOperatorLanguage`

- Source: `internal/console/a079_quarantine_release_test.go` (287-331)
- AST evidence: `ast.json` — AST 분기 3 · return 0 · defer 0
  (source_sha256 는 ast.json 이 정본, a109 §2-fix F5 편집 후 생성)
- Risk scan: `risk-pattern-report.md`
- **a109 §2-fix 편집 대상: 표의 한 칸** — 미배선 행의 기대 제목을 문자열 리터럴에서
  **상수 `quarantineUnwiredTitle`** 로 바꾼다. 나머지 네 행과 harness 는 a079 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 표 5행 | (sentinel, status, 화면 문구) | a079 가 정한 매핑 | 어긋나면 `t.Fatalf`/`t.Errorf` (B2·B3) |
| 미배선 행의 기대 제목 | `quarantineUnwiredTitle` | `internal/console/exit_quarantine.go` | 제목이 갈라지면 실패 |
| harness | 콘솔 + 가짜 commander | `quarantineHarness` | preview 오류를 주입해 각 갈래를 만든다 |

**a079 가 지는 요구**: 엔진이 돌려준 거부가 **운영자 언어**로 번역된다 — 상태 코드와
화면 문구가 sentinel 마다 다르다.

**a109 §2-fix F5 가 고친 것**: 미배선 행이 옛 제목(`"판정 격리 command seam 미배선"`)을
**문자열로 다시 적고** 있었다. 제목을 상수 한 벌로 모으는 정정에서 이 줄을 놓치면
테스트가 옛 제목을 되살리는 편집을 통과시킨다 — 정정의 단위는 줄이 아니라 값이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:288) | 표 5행 반복 | 하위 테스트 실행 | — | — |
| B2 (:323) | 상태 코드가 다르다 | 없음 | `t.Fatalf` | — |
| B3 (:326) | 화면에 기대 문구가 없다 | 없음 | `t.Errorf` | 뮤테이션 M20(문구 세 곳 중 하나만 되돌리기) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `quarantineHarness` | 콘솔 + 주입 가능한 commander | 없음 | AST · :302 |
| `h.post` | preview 경로를 실제로 두드린다 | HTTP 응답만 | AST · :322 |
| `strings.Contains` | 화면 문구 확인 | 순수 | AST · :326 |

## State mutations and fallbacks

- 테스트 지역 harness 뿐. 디스크·엔진에 닿지 않는다.

## Safety conclusion

- Safe edit boundary: 표의 **기대값 칸**. 상태 코드 매핑은 a079 소유다.
- High-risk impact: **no** — 콘솔 화면 문구의 회귀 핀이다.
- 금지: 기대 제목을 문자열 리터럴로 되돌리는 것(F5 가 지운 사본), 미배선 행 삭제.
