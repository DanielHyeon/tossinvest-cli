# Function Logic Map: `TestMutatingAnnotationOnTradeCommands`

- Source: `cmd/tossctl/help_convention_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :96–154, 분기 3, source_sha256 `fb6a9804667c…`, 추출 커밋 `2714e393`.
- Risk scan: `risk-pattern-report.md`
- 편집: 허용 표에 `tossctl engine mode-release` · `tossctl engine alerts ack` 두 항목을 더함(a066 의 `*-release` 가족과 같은 근거 — 노출 증가를 다시 여는 완화 · 원장 쓰기).
- 비례 원칙: 시험 코드 → 변이 원장은 생산 변이(`--set 25.9`)가 반증 도구로 씀. `not-applicable`: 이 함수 자체의 다중 리뷰.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range at :144 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B2 | if at :147 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B3 | if at :150 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 자원만.

## Safety conclusion

- Safe edit boundary: 표 · 단언 편집.
- High-risk impact: no — 시험 코드.
