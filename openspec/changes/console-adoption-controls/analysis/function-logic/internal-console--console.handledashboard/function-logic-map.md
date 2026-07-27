# Function Logic Map: `Console.handleDashboard`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

404 안내 문구의 화면 수만 갱신(다섯→여섯, 편입 설정 추가). 로직 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if | 없음 | — | 콘솔 스위트 green(404 경로 기존 케이스) |
| B2 | if | 없음 | — | 콘솔 스위트 green(404 경로 기존 케이스) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음

## Safety conclusion

- Safe edit boundary: 문자열 1행만
- High-risk impact: no (콘솔·config·배선 — 주문·원장 무접촉)
