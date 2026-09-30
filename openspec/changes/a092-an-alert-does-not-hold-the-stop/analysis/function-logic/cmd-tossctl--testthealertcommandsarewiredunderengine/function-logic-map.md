# Function Logic Map: `TestTheAlertCommandsAreWiredUnderEngine`

- Source: `cmd/tossctl/a098_the_operator_command_names_a_person_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :135–170, 분기 11, source_sha256 `805b502b1d70…`, 추출 커밋 `2714e393`.
- Risk scan: `risk-pattern-report.md`
- 편집: **교차 change(a098) 시험 단언 반전** — 옛 단언 「ack 는 mutating 이 아니다」는 frozen 24판 22.3 C17(Manager 판정 4: ack 는 원장 쓰기라 `mutating: "true"`)과 정면 충돌. 새 단언은 `mutating == "true"`. 추가 확인 마찰 없음 · `--operator` 기본값 없음 단언은 그대로.
- 비례 원칙: 시험 코드 → 변이 원장은 생산 변이(`--set 25.9`)가 반증 도구로 씀. `not-applicable`: 이 함수 자체의 다중 리뷰.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range at :137 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B2 | switch at :138 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B3 | case at :139 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B4 | case at :141 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B5 | case at :143 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B6 | if at :147 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B7 | if at :150 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B8 | if at :154 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B9 | if at :158 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B10 | if at :161 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |
| B11 | if at :167 — 단언 · 픽스처 갈래 | `t.Error` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 자원만.

## Safety conclusion

- Safe edit boundary: 표 · 단언 편집.
- High-risk impact: no — 시험 코드.
