# Function Logic Map: `Journal.LiveOrdersForSymbol`

- Source: `internal/journal/fills.go` (`1849`–`1914`)
- Qualified: `Journal.LiveOrdersForSymbol`
- AST evidence: `ast.json` (`source_sha256` 6214babe8bfa3047…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 7

**역할.** 엔진 귀속 미체결 주문 목록. a094: 종결 증거 술어를 공유 상수(`confirmedOrderTerminalEvidence`)로 — SQL 바이트는 같은 술어.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `accountRef` · `market` · `symbol` | 범위 | 호출자 |  |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1851` `if err := j.guardTrackedFillIdentity(ctx, accountRef); err != nil {` | 예 |
| B2 | if | `:1881` `if err != nil {` | 아니오 |
| B3 | for | `:1887` `for rows.Next() {` | 예 |
| B4 | if | `:1889` `if err := rows.Scan(&o.OrderID, &o.IntentID, &o.AccountRef, &o.Market, &o.TradingDay,` | — |
| B5 | if | `:1896` `if err := rows.Err(); err != nil {` | 예 |
| B6 | range | `:1900` `for i := range out {` | 예 |
| B7 | if | `:1908` `if err != nil {` | 아니오 |

## Calls and live bindings

`guardTrackedFillIdentity`(같은 범위 소유 모호 = 오류) · 질의 · `ResolveCurrentOrderIDScoped`.

## State mutations and fallbacks

없음.

## Safety conclusion

- 동작 무변(술어 문자열을 상수로 옮김). 해제 판정과 같은 술어를 쓰게 해 판정을 둘로 두지 않는다. High-risk: yes(청소의 목록).
