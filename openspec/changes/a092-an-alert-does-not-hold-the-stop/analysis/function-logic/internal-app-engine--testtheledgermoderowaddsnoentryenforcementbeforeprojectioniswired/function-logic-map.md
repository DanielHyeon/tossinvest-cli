# Function Logic Map: `TestTheLedgerModeRowAddsNoEntryEnforcementBeforeProjectionIsWired`

- Source: `internal/app/engine/a124_the_enforcement_boundary_internal_test.go`
- AST evidence: `ast.json` — **base**(삭제 전), :147–183, 분기 8, source_sha256 `225ebd6478f4…`, 추출 base `721d0338`.
- Risk scan: `risk-pattern-report.md`
- 편집: **삭제(이름 변경)된 a124 핀**(`revision: base`). 새 핀 `TestTheLedgerModeRowEnforcesEntryOnceProjectionIsWired`: 승인은 알림 사유만 풀고 모드 사유는 남음 · 관측을 다 줘도 진입은 모드 사유로 거절 · 재시작(원장 재오픈) 뒤 복원이 모드 사유를 세움.
- 비례 원칙: 시험 코드 → 변이 원장은 생산 변이(`--set 25.9`)가 반증 도구로 씀. `not-applicable`: 이 함수 자체의 다중 리뷰.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | for at :150 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B2 | if at :153 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B3 | if at :157 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B4 | if at :161 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B5 | if at :165 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B6 | if at :170 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B7 | if at :177 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B8 | if at :180 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 자원만.

## Safety conclusion

- Safe edit boundary: 삭제(이름 변경) — 대체 핀이 같은 파일에.
- High-risk impact: no — 시험 코드.
