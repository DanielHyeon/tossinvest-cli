# Function Logic Map: `positionsPage.Refresh`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json` (구현 후 재추출 — 분기 0, 단일 return)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (없음 — 수신자 무상태) | — | head 템플릿이 `{{if .Refresh}}`로 소비 | 해당 없음 |

불변식: 상수 반환 순수 함수. true로 바뀌면서 head 템플릿이 `RefreshSeconds()`
(= `holdingsTTL`초, TTL에서 유도)를 함께 읽는다. Go 코드는 이 메서드를 호출하지 않는다
(static 가드 `.Refresh(` 금지와 정합 — 템플릿 전용).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (단일 경로) | — | 없음 | `true` — positions 화면이 재로드 지시를 갖는다 | `TestThePositionsScreenAsksTheBrowserToReloadAtTheCacheTTL` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 없음. 재로드의 비용 불변식은 이 함수가 아니라 `holdingsCache.get`의 TTL·hold 판정이
  소유한다(무변경) — `TestRefreshingInsideTheTTLCostsNoBrokerCall`,
  `TestAVerificationInProgressSuspendsTheRefresh`가 가드.

## Safety conclusion

- Safe edit boundary: 반환값 플립 + 자매 메서드 `RefreshSeconds`(신규 leaf) 추가만
- High-risk impact: no (콘솔 read-only 렌더 경로)
