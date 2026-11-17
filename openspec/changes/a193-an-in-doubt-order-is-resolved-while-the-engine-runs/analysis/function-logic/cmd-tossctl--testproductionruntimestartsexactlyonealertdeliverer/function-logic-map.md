# Function Logic Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

> **시험 코드다.** a193 이 이 함수를 고치는 이유는 하나 — 보조 실행자 기대 목록에 `in-doubt-settle` 를 더한다. 단언 방식(집합 동일성 `reflect.DeepEqual`)은
> 그대로다. 선례: a087·a135·a136 이 같은 줄을 같은 이유로 고쳤다(a102 번들 형식).

- Source: `cmd/tossctl/a098_the_engine_registers_a_sender_test.go` (20-31)
- AST evidence: `ast.json` — branches **2** / returns 0 / calls 8, source SHA-256 `a626861e1479fde89301ff26b0dc4af80f0647498d3cd1bfc9805b7eb94b4387`(HEAD `791fb6ef1`).
- Risk scan: `risk-pattern-report.md`(일치 0).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `runtimeBranchContext()` | 영값 위주 조립 문맥 | engine_runtime_branch_test.go:32-47 | — |
| 불변식 | 보조 실행자 집합이 **정확히** 기대 목록(순서 포함) | 이 시험 | 다르면 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@22) | `engineRuntime`(call@21) 오류 | 없음 | `t.Fatalf`(call@23) | 자기 자신(성공 경로가 정상) |
| B2 (if@28) | `AuxiliaryNames`(call@28) ≠ 기대 목록 | 없음 | `t.Fatalf`(call@29) | 자기 자신 — **a193 편집: 기대 목록 끝에 `"in-doubt-settle"`** |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineRuntime`(call@21) | 프로덕션 조립 | 오류면 B1 | AST |
| `runtime.AuxiliaryNames`(call@28) | 보조 실행자 이름 열거 | — | AST |

## State mutations and fallbacks

- 편집: 문자열 슬라이스 리터럴에 원소 1(분기 아님). 주석 한 줄(a193).

## Safety conclusion

- Safe edit boundary: 기대 목록 원소 1. 분기 2개 문장 불변.
- High-risk impact: no(시험). 다만 이 시험이 배선 원소 삭제 변이(m11)를 잡는다.
