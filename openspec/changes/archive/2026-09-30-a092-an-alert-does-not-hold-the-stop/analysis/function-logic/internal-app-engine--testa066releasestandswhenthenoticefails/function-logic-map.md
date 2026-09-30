# Function Logic Map: `TestA066ReleaseStandsWhenTheNoticeFails`

- Source: `internal/app/engine/a066_risk_relaxation_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :221–234, 분기 3, source_sha256 `092f36a777f7…`, 추출 커밋 `0e4f26af`.
- Risk scan: `risk-pattern-report.md`
- 편집: 실패 주입 자리를 원장(`EnqueueAlert` 덮어쓰기)에서 기록자(`service.notices = a066FailingNotices{}`)로. 단언은 그대로. 반증: 변이 R10 · R15 에서 FAIL.
- 비례 원칙: 시험 코드라 High-risk 경로 · 게이트 판정 변경이 아님 → 변이 원장 · 다중 리뷰는 이 함수에 대해 `not-applicable`(생산 쪽 변이 R01~R15 가 이 시험들을 반증 도구로 씀).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at :225 — 픽스처 · 단언 갈래(편집이 바꾸지 않음) | `t.Fatal` 류 | — | 자기 자신 |
| B2 | if at :228 — 픽스처 · 단언 갈래(편집이 바꾸지 않음) | `t.Fatal` 류 | — | 자기 자신 |
| B3 | if at :231 — 픽스처 · 단언 갈래(편집이 바꾸지 않음) | `t.Fatal` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 구성 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 디렉터리 · 임시 원장만.

## Safety conclusion

- Safe edit boundary: 픽스처 배선 · 주입 자리만 바뀜, 단언 불변.
- High-risk impact: no — 시험 코드.
