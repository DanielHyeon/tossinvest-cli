# Function Logic Map: `Adoption.validate`

- Source: `internal/config/engine.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

'의미 있음'을 enabled ∨ include≠∅로 확장 — include만 있어도 손절폭 검증 요구. absent 블록 경로 불변(§0.2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if | 없음 | — | TestIncludeOnlyBlockNeedsTheStopFraction(RED)·TestAdoptionOffWithNoFractionIsNotARefusal green |
| B2 | if | 없음 | — | TestIncludeOnlyBlockNeedsTheStopFraction(RED)·TestAdoptionOffWithNoFractionIsNotARefusal green |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음

## Safety conclusion

- Safe edit boundary: 첫 조건식 확장만
- High-risk impact: no (콘솔·config·배선 — 주문·원장 무접촉)
