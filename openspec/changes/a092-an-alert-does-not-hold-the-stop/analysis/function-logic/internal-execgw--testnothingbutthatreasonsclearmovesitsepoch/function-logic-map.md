# Function Logic Map: `TestNothingButThatReasonsClearMovesItsEpoch`

- Source: `internal/execgw/a124_clear_epoch_internal_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :61–83, 분기 1, source_sha256 `5638ae17b055…`, 추출 커밋 `2714e393`.
- Risk scan: `risk-pattern-report.md`
- 편집: **교차 change(a124) 시험** — 두 투영 호출에 순번 1 · 2 를 붙임. a092 울타리(순번 0 은 적용 안 함) 뒤에는 순번 없는 투영이 아무것도 안 해서 이 시험의 「투영은 해제 세대를 안 바꾼다」가 공허하게 참이 됐을 것.
- 비례 원칙: 시험 코드 → 변이 원장은 생산 변이(`--set 25.9`)가 반증 도구로 씀. `not-applicable`: 이 함수 자체의 다중 리뷰.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at :80 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 자원만.

## Safety conclusion

- Safe edit boundary: 표 · 단언 편집.
- High-risk impact: no — 시험 코드.
