# Function Logic Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

> **시험 코드다.** a193 이 이 함수를 고치는 이유는 하나 — 보조 실행자 기대 목록에 `in-doubt-settle` 를 더한다. 단언 방식(집합 동일성 `reflect.DeepEqual`)은
> 그대로다. 선례: a087·a135·a136 이 같은 줄을 같은 이유로 고쳤다(a102 번들 형식).

- Source: `cmd/tossctl/a098_the_engine_registers_a_sender_test.go` (20-32, 편집 후). 편집 전 원본 `../../pre-edit-ast/cmd-tossctl--testproductionruntimestartsexactlyonealertdeliverer.ast.json`(20-31, sha `a626861e1479`). 재번호 `renumber.txt`(B2 28→29, 주석 1줄).
- AST evidence: `ast.json` — branches **2** / returns 0 / calls 8, source SHA-256 `0084b4e6f4ac39db0e677c5ade76d535fb48928c1df59be5317d92d399ac0c79`(편집 후).
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
| B2 (if@29) | `AuxiliaryNames`(call@29) ≠ 기대 목록 | 없음 | `t.Fatalf`(call@30) | 자기 자신 — **a193 편집: 기대 목록 끝에 `"in-doubt-settle"`** |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineRuntime`(call@21) | 프로덕션 조립 | 오류면 B1 | AST |
| `runtime.AuxiliaryNames`(call@29) | 보조 실행자 이름 열거 | — | AST |

## State mutations and fallbacks

- 편집(완료): 문자열 슬라이스 리터럴에 원소 `"in-doubt-settle"` 1(분기 아님) + 주석 한 줄(a193). 분기 2개 문장 불변(B2 줄 +1).

## Safety conclusion

- Safe edit boundary: 기대 목록 원소 1. 분기 2개 문장 불변.
- High-risk impact: no(시험). 다만 이 시험이 배선 원소 삭제 변이(m11)를 잡는다.
