# Function Logic Map: `Console.handlePositions`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

CSRF 전달 + seam 판독 가능 시 CanDesignate·행별 Designated 스탬프 추가. 렌더 경로 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if | 없음 | — | TestDesignatingASymbolFromThePositionsScreen·portfolio 기존 스위트 green |
| B2 | if | 없음 | — | TestDesignatingASymbolFromThePositionsScreen·portfolio 기존 스위트 green |
| B3 | range | 없음 | — | TestDesignatingASymbolFromThePositionsScreen·portfolio 기존 스위트 green |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음(표시 데이터 구성만; seam은 Load 판독 전용)

## Safety conclusion

- Safe edit boundary: 페이지 구성 확장만
- High-risk impact: no (콘솔·config·배선 — 주문·원장 무접촉)
