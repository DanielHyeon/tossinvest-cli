# Function Logic Map: `positionsPage.RefreshSeconds`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

change refresh-positions-screen에서 신설된 무상태 leaf 함수다(diff 교차로 evidence 요구). positions 재로드 주기 = `holdingsTTL`에서 유도(30초) — 스펙 '주기는 캐시 TTL 이상' SHALL을 상수 유도로 고정.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `holdingsTTL` | 30s (스펙 하한 15s 이상) | holdings.go 상수 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 | `int(holdingsTTL/time.Second)` = 30 | `TestThePositionsScreenAsksTheBrowserToReloadAtTheCacheTTL` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음 — 상수 나눗셈) | — | — | ast.json |

## State mutations and fallbacks

- 없음. TTL 변경 시 주기가 자동 추종하므로 드리프트 불가.

## Safety conclusion

- Safe edit boundary: 신규 leaf — TTL 유도 상수만
- High-risk impact: no (콘솔 read-only 렌더 경로 — 주문·위험·원장 코드 무접촉)
