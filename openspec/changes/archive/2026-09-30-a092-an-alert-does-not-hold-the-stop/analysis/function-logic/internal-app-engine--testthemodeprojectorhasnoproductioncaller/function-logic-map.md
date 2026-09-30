# Function Logic Map: `TestTheModeProjectorHasNoProductionCaller`

- Source: `internal/app/engine/a124_the_enforcement_boundary_internal_test.go`
- AST evidence: `ast.json` — **base**(삭제 전), :90–101, 분기 2, source_sha256 `225ebd6478f4…`, 추출 base `721d0338`.
- Risk scan: `risk-pattern-report.md`
- 편집: **삭제(이름 변경)된 a124 핀**(`revision: base`, base `721d0338`). a124 자신이 「배선이 착지하면 빨강이 되고 D10 · spec 경계 문장 · 이 핀을 함께 고친다(정상 경로)」라고 적은 핀 — a092 tasks 21.7(d). 새 핀 `TestTheModeProjectorIsBoundOnlyByTheEngineAssembly`: 비시험 호출 자리가 정확히 둘이고 둘 다 `mode_projection_wiring.go`.
- 비례 원칙: 시험 코드 → 변이 원장은 생산 변이(`--set 25.9`)가 반증 도구로 씀. `not-applicable`: 이 함수 자체의 다중 리뷰.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at :93 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B2 | if at :97 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 자원만.

## Safety conclusion

- Safe edit boundary: 삭제(이름 변경) — 대체 핀이 같은 파일에.
- High-risk impact: no — 시험 코드.
