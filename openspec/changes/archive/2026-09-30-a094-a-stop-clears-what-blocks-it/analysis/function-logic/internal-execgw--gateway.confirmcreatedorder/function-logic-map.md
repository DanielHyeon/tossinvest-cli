# Function Logic Map: `Gateway.confirmCreatedOrder`

- Source: `internal/execgw/roundtrip.go` (`85`–`92`)
- Qualified: `Gateway.confirmCreatedOrder`
- AST evidence: `ast.json` (`source_sha256` face1097e68e93a0…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 0

**역할.** 발주 직후 확인 — 이제 `ConfirmPlacedOrder`(기동 ACKED 확정과 공유) 한 호출. 응답 종목이 비면 확인 실패(D−5.1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `brokerOrderID` | 브로커가 준 번호 | dispatch | 바이트 일치 요구 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|

## Calls and live bindings

`ConfirmPlacedOrder`(OrderRaw · roundTripTimeout · parseOrderFacts · 번호 바이트 일치 · 종목 비공백 일치).

## State mutations and fallbacks

없음.

## Safety conclusion

- 보수 방향 강화(종목 공백 = 실패 → IN_DOUBT). High-risk: yes(접수 확정).
