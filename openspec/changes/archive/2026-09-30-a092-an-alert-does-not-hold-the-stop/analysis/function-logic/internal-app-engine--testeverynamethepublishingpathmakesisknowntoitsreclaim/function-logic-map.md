# Function Logic Map: `TestEveryNameThePublishingPathMakesIsKnownToItsReclaim`

- Source: `internal/app/engine/a109_the_sibling_endpoints_recover_too_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :379–436, 분기 7, source_sha256 `4ed962de949a…`, 추출 커밋 `2714e393`.
- Risk scan: `risk-pattern-report.md`
- 편집: 이름-집합 완전성 표에 모드 제어(`modeControlEndpointNames`)를 더함.
- 비례 원칙: 시험 코드 → 변이 원장은 생산 변이(`--set 25.9`)가 반증 도구로 씀. `not-applicable`: 이 함수 자체의 다중 리뷰.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range at :380 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B2 | if at :410 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B3 | if at :416 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B4 | if at :419 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B5 | range at :424 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B6 | if at :425 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B7 | if at :431 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 자원만.

## Safety conclusion

- Safe edit boundary: 표 · 단언 편집.
- High-risk impact: no — 시험 코드.
