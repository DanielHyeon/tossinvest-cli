# Function Logic Map: `Adoption.Excludes`

- Source: `internal/config/engine.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

본문을 공용 onSymbolList로 위임 — 동작 동일(정규화 비교), Included와 철자 드리프트 방지.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 | 단일 return | 위 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음

## Safety conclusion

- Safe edit boundary: 위임 리팩터링만(동작 보존)
- High-risk impact: no (콘솔·config·배선 — 주문·원장 무접촉)
