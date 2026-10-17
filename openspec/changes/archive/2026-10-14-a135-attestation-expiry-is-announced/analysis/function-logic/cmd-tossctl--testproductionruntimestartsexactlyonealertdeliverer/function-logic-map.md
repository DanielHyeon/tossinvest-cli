# Function Logic Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

- Source: `cmd/tossctl/a098_the_engine_registers_a_sender_test.go` (20-31, 편집 뒤 — base `421bf53c` 20-30)
- Qualified: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 91e470893ed867f0… · base d4eefec55df63486…) — AST branches 2 · returns 0(편집 전후 동일)
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 함수 — 생산 조립(`engineRuntime`)의 보조 실행자 이름 집합을 집합 동일성으로 고정. a135 편집은 기대 슬라이스에
`"attestation-watch"` 한 원소 + 주석 한 줄. 분기 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `engineRuntime(...)` 결과 | 오류 없음 · 런타임 | `cmd/tossctl/engine.go` | B1 `t.Fatalf` |

## Branches and early returns

| Branch | AST line | Condition | Effect | a135 |
|---|---|---|---|---|
| B1 | 22 | 조립 오류 | `t.Fatalf` | 무편집 |
| B2 | 28 | 이름 집합 ≠ 기대 | `t.Fatalf` | 조건 무편집, 기대값만 셋 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineRuntime` · `runtimeBranchContext` · `clock.NewFake` | 생산 조립 재현 | 오류 → B1 | AST calls |
| `runtime.AuxiliaryNames` · `reflect.DeepEqual` | 집합 비교 | 없음 | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 기대값 한 원소. 부분 일치로 느슨해지지 않음(DeepEqual 유지).
- High-risk impact: no.
