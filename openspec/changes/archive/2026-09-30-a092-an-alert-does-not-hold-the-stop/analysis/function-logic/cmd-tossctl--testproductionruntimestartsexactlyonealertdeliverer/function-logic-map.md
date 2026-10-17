# Function Logic Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

- Source: `cmd/tossctl/a098_the_engine_registers_a_sender_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :20–30, 분기 2, source_sha256 `d4eefec55df6…`, 추출 커밋 `e55102f0`.
- Risk scan: `risk-pattern-report.md`
- 편집: **교차 change(a098) 시험 표 갱신** — 생산 보조 실행자 집합 `[alert-delivery]` → `[alert-delivery, normal-alert-relay]`(a092 C8). 배달 실행자가 정확히 하나라는 주제(이름)는 그대로이고 집합 동일성 단언도 그대로.
- 비례 원칙: 시험 코드 — 반증은 생산 변이 N08(relay 실행자 미등록 → 이 시험 FAIL).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at :22 — 단언 갈래 | `t.Fatal` 류 | — | 자기 자신 |
| B2 | if at :27 — 단언 갈래 | `t.Fatal` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineRuntime` · `AuxiliaryNames` | 생산 조립 · 집합 | 시험 | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 기대 집합 한 줄.
- High-risk impact: no — 시험 코드.
