# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

가드 개정: stateChanging 맵에 /settings/save·/settings/include 추가(스펙 델타 반영). 단언 로직 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range | 없음 | — | 자체 실행(RED: 신규 라우트 불일치로 실패 관측 → GREEN) |
| B2 | switch | 없음 | — | 자체 실행(RED: 신규 라우트 불일치로 실패 관측 → GREEN) |
| B3 | case | 없음 | — | 자체 실행(RED: 신규 라우트 불일치로 실패 관측 → GREEN) |
| B4 | case | 없음 | — | 자체 실행(RED: 신규 라우트 불일치로 실패 관측 → GREEN) |
| B5 | range | 없음 | — | 자체 실행(RED: 신규 라우트 불일치로 실패 관측 → GREEN) |
| B6 | if | 없음 | — | 자체 실행(RED: 신규 라우트 불일치로 실패 관측 → GREEN) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음(테스트)

## Safety conclusion

- Safe edit boundary: 허용 목록 2항 추가만
- High-risk impact: no (콘솔·config·배선 — 주문·원장 무접촉)
