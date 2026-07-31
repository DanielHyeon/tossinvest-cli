# Function Logic Map: `onSymbolList`

- Source: `internal/config/engine.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

신규 leaf — Excludes/Included 공용 정규화 비교(대문자·trim, 빈 심볼 불일치).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if | 없음 | — | Excludes/Included 양쪽 테스트가 간접 커버 |
| B2 | range | 없음 | — | Excludes/Included 양쪽 테스트가 간접 커버 |
| B3 | if | 없음 | — | Excludes/Included 양쪽 테스트가 간접 커버 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음

## Safety conclusion

- Safe edit boundary: 공용 헬퍼 신설
- High-risk impact: no (콘솔·config·배선 — 주문·원장 무접촉)
